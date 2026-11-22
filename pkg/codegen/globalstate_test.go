package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// resetGlobalState puts the package's global generator state into what
// Generate sets up for a zero Configuration, before it has seen a spec, and
// puts back whatever was there when the test ends.
//
// A test that reads or writes that state without going through Generate calls
// this first, and then sets only what it is about: the test neither inherits
// what an earlier test left behind nor leaves anything for a later one. A
// subtest that changes the state calls it again, so its siblings start clean
// too. Tests that use it can't run in parallel, as the state is global.
//
// The state is globalState, and the package-level variables Generate sets
// alongside it: nameNormalizer, targetWordRegex, responseTypeSuffix, and the
// `opts` entry of TemplateFunctions.
func resetGlobalState(t *testing.T) {
	t.Helper()

	prevState := globalState
	prevNormalizer := nameNormalizer
	prevRegex := targetWordRegex
	prevSuffix := responseTypeSuffix
	prevOpts, hadOpts := TemplateFunctions["opts"]
	t.Cleanup(func() {
		globalState = prevState
		nameNormalizer = prevNormalizer
		targetWordRegex = prevRegex
		responseTypeSuffix = prevSuffix
		if hadOpts {
			TemplateFunctions["opts"] = prevOpts
		} else {
			delete(TemplateFunctions, "opts")
		}
	})

	// The struct tag generators and the content-type name tags stay nil:
	// their accessors build them from the options on first use, so a test
	// that sets options after this gets generators that match them.
	setZero(&globalState)
	globalState.options.OutputOptions.ClientTypeName = defaultClientTypeName
	globalState.typeMapping = DefaultTypeMapping
	// This sets targetWordRegex as well.
	globalState.initialismsMap = makeInitialismsMap(nil)
	streaming, err := compileStreamingContentTypes(nil)
	require.NoError(t, err)
	globalState.streamingContentTypeRegexes = streaming

	nameNormalizer = ToCamelCase
	responseTypeSuffix = "Response"
	TemplateFunctions["opts"] = func() Configuration { return globalState.options }
}

// setZero sets *p to its zero value. globalState's type has no name to spell
// a zero literal with.
func setZero[T any](p *T) {
	var zero T
	*p = zero
}
