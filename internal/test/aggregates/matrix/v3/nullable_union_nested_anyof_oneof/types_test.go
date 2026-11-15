package matrixv3nullableunionnestedanyofoneof

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNestedNullableUnionFromKeepsOtherUnion: the inline schema the null
// branch wraps keeps the keys each of its unions owns, so From* of a oneOf
// variant replaces the oneOf's data and keeps the anyOf's (issue #2565).
func TestNestedNullableUnionFromKeepsOtherUnion(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"meow": "purr", "red": "crimson"}`), &s))
	require.NoError(t, s.FromDog(Dog{Bark: "woof"}))

	b, err := json.Marshal(s)
	require.NoError(t, err)
	assert.JSONEq(t, `{"bark": "woof", "red": "crimson"}`, string(b))

	require.NoError(t, s.FromBlue(Blue{Blue: "navy"}))
	b, err = json.Marshal(s)
	require.NoError(t, err)
	assert.JSONEq(t, `{"bark": "woof", "blue": "navy"}`, string(b))
}
