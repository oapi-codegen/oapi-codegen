// Package enum_via_oneof tests the OpenAPI 3.1 enum-via-oneOf idiom: a
// scalar schema whose oneOf branches each carry `title` + `const` is
// emitted as a Go typed enum with named constants.
//
// The tests are structural -- they instantiate the generated types and
// assert constant values and JSON round-trips -- rather than string-
// matching the generated source.
package enum_via_oneof

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSeverityConstants verifies that an integer enum-via-oneOf produces
// `type Severity int` with the right per-branch constant values.
func TestSeverityConstants(t *testing.T) {
	assert.Equal(t, 2, int(HIGH))
	assert.Equal(t, 1, int(MEDIUM))
	assert.Equal(t, 0, int(LOW))
}

// TestSeverityJSONRoundTrip confirms Severity marshals as its integer
// value, not as a wrapped union or as a string.
func TestSeverityJSONRoundTrip(t *testing.T) {
	data, err := json.Marshal(HIGH)
	require.NoError(t, err)
	assert.JSONEq(t, `2`, string(data))

	var got Severity
	require.NoError(t, json.Unmarshal([]byte(`1`), &got))
	assert.Equal(t, MEDIUM, got)
}

// TestColorConstants verifies that a string enum-via-oneOf produces
// `type Color string` with the right per-branch constant values.
func TestColorConstants(t *testing.T) {
	assert.Equal(t, "r", string(Red))
	assert.Equal(t, "g", string(Green))
	assert.Equal(t, "b", string(Blue))
}

// TestColorJSONRoundTrip confirms Color marshals as its string value.
func TestColorJSONRoundTrip(t *testing.T) {
	data, err := json.Marshal(Red)
	require.NoError(t, err)
	assert.JSONEq(t, `"r"`, string(data))

	var got Color
	require.NoError(t, json.Unmarshal([]byte(`"b"`), &got))
	assert.Equal(t, Blue, got)
}

// TestMixedOneOfFallsThrough verifies the negative path: a oneOf where
// any branch lacks `title` must NOT trigger enum-via-oneOf detection.
// MixedOneOf is emitted by the standard handler as `type MixedOneOf =
// string` (an alias), so a plain string is directly assignable. If
// detection were over-eager, MixedOneOf would become a newtype `type
// MixedOneOf string` and the assignment below would fail to compile
// (a string would not be directly assignable to a newtype value).
func TestMixedOneOfFallsThrough(t *testing.T) {
	var s = "anything"
	var m MixedOneOf = s
	assert.Equal(t, "anything", string(m))
}
