package issue2412

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These compile-time assignments are the core of the regression: an
// `anyOf`/`oneOf` of a primitive and that same primitive constrained by an
// `enum` must keep the Go type as the bare primitive, so any value is
// assignable without a conversion and the generated constants are usable as
// plain strings.
//
// This is a regression test for
// https://github.com/oapi-codegen/oapi-codegen/issues/2412
var (
	_ Source   = "any free-form string"
	_ Source   = Rootly
	_ string   = Rootly
	_ Priority = "anything goes"
	_ Priority = High
)

// TestUntypedEnumStaysString verifies that the generated type is a string and
// that the known enum values are emitted as (untyped) constants — for both the
// `anyOf` and `oneOf` spellings.
func TestUntypedEnumStaysString(t *testing.T) {
	assert.Equal(t, reflect.String, reflect.TypeOf(Source("")).Kind(),
		"Source should be a string, not a distinct enum type")
	assert.Equal(t, reflect.String, reflect.TypeOf(Priority("")).Kind(),
		"Priority should be a string, not a distinct enum type")

	// anyOf: string | enum(string)
	assert.Equal(t, "rootly", Rootly)
	assert.Equal(t, "manual", Manual)
	assert.Equal(t, "api", Api)
	assert.Equal(t, "heartbeat", Heartbeat)

	// oneOf, enum-bearing member first.
	assert.Equal(t, "low", Low)
	assert.Equal(t, "high", High)
}

// TestInlinePropertyStaysString verifies that the idiom nested inside a
// property keeps the field as a string. Event.Source is *EventSource where
// `type EventSource = string`, so a plain *string is assignable.
func TestInlinePropertyStaysString(t *testing.T) {
	s := "free-form"
	e := Event{Source: &s}
	require.NotNil(t, e.Source)
	assert.Equal(t, "free-form", string(*e.Source))

	// The known values are still emitted as constants.
	assert.Equal(t, "webhook", Webhook)
	assert.Equal(t, "poll", Poll)
}

// TestMoreThanTwoMembersStaysUnion verifies the special case is deliberately
// strict: a union with more than two members is not collapsed to a string but
// rendered as a normal union type.
func TestMoreThanTwoMembersStaysUnion(t *testing.T) {
	assert.Equal(t, reflect.Struct, reflect.TypeOf(Mixed{}).Kind(),
		"a 3-member anyOf should remain a union struct, not collapse to a string")
}
