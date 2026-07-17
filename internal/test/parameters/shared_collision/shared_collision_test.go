package parameterssharedcollision

import (
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Referencing these names asserts, at compile time, that non-colliding shared
// (path-item-level) parameters keep their historical undecorated form, so
// existing code is unaffected (issue #2090). The generated file compiling is
// itself the primary guard — the "redeclared in this block" bug would fail the
// build.
//
//   - Wid0/Wid1: shared across two methods of /widget/{wid}. Declared once for
//     the path item (not once per method), and bare because the name is unique.
//   - Ref0/Ref1: the single-method /sprocket/{ref}, also unique and bare.
//   - Whid0/Whid1: shared across the two methods of the onEvent webhook, and
//     Cbid0/Cbid1 across the two methods of the onData callback — the webhook
//     and callback emit-once paths, declared once each rather than per method.
var (
	_ Wid0
	_ Wid1
	_ Ref0
	_ Ref1
	_ Whid0
	_ Whid1
	_ Cbid0
	_ Cbid1
)

// TestSharedParamCollisionHashed asserts that the same {id} parameter, reused by
// the two sibling paths /gadget/{id} and /gadget/{id}/part, is disambiguated by
// a per-path hash prefix rather than colliding on a bare name. We read the
// generated source because the hash values are derived from the paths and
// shouldn't be hard-coded into the test.
func TestSharedParamCollisionHashed(t *testing.T) {
	src, err := os.ReadFile("shared_collision.gen.go")
	require.NoError(t, err)

	// The bare, colliding names must NOT be declared.
	assert.NotRegexp(t, regexp.MustCompile(`(?m)^type Id0 `), string(src))
	assert.NotRegexp(t, regexp.MustCompile(`(?m)^type Id1 `), string(src))

	// Exactly two distinct hash-prefixed variants of the id union member must be
	// declared — one per colliding path.
	variants := regexp.MustCompile(`(?m)^type (H[0-9a-f]+Id0) `).FindAllStringSubmatch(string(src), -1)
	seen := map[string]bool{}
	for _, m := range variants {
		seen[m[1]] = true
	}
	assert.Len(t, seen, 2, "expected two per-path hash-disambiguated id types, got %v", seen)
}
