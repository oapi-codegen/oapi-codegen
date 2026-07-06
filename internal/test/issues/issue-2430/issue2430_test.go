package issue2430

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression test for https://github.com/oapi-codegen/oapi-codegen/issues/2430.
//
// OpenAPI 3.1 allows a schema whose only type is "null" (as opposed to a
// type array that merely includes "null"). Such a schema validates exactly
// the JSON value null. The generator used to fail with "unhandled Schema
// type: &[null]"; it now maps the schema to `any`, with no optional
// pointer, since nil is already `any`'s zero value.

func TestNullTypePropertyRoundTrip(t *testing.T) {
	var v ChallengeOpenJson
	require.NoError(t, json.Unmarshal([]byte(`{"id":"1","challenger":null}`), &v))
	assert.Equal(t, "1", v.Id)
	assert.Nil(t, v.Challenger)

	out, err := json.Marshal(RequiredNull{})
	require.NoError(t, err)
	// The required null-typed field must serialize as an explicit null.
	assert.JSONEq(t, `{"value":null}`, string(out))
}

func TestNullTypeComponentIsAny(t *testing.T) {
	// NullOnly is an alias for `any`; assigning an arbitrary value must
	// compile, and nil round-trips through JSON as null.
	var n NullOnly
	require.NoError(t, json.Unmarshal([]byte(`null`), &n))
	assert.Nil(t, n)
}
