package matrixv2oneofadditionalproperties

// This test records how schema-merging-behavior v2 handles this shape. A bug fix
// may change it, but it must not be changed to accept a regression: code that
// generated and worked before must keep doing so. A commit that changes it must
// say why.

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFromReplacesStaleAdditionalProperty: UnmarshalJSON copies the variant's
// keys into AdditionalProperties as well as the union, and MarshalJSON writes
// AdditionalProperties last. A later From* must not be shadowed by that copy.
func TestFromReplacesStaleAdditionalProperty(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"kind": "cat", "meow": "old", "extra": 7}`), &s))
	require.NoError(t, s.FromCat(Cat{Meow: "new"}))

	cat, err := s.AsCat()
	require.NoError(t, err)
	assert.Equal(t, "new", cat.Meow)

	b, err := json.Marshal(s)
	require.NoError(t, err)
	assert.JSONEq(t, `{"kind": "cat", "meow": "new", "extra": 7}`, string(b))
}

// TestAdditionalPropertySetAfterFromWins: the caller's own additional
// properties keep the last word, as they always have.
func TestAdditionalPropertySetAfterFromWins(t *testing.T) {
	s := Subject{Kind: "cat"}
	require.NoError(t, s.FromCat(Cat{Meow: "purr"}))
	s.Set("extra", 8)

	b, err := json.Marshal(s)
	require.NoError(t, err)
	assert.JSONEq(t, `{"kind": "cat", "meow": "purr", "extra": 8}`, string(b))
}

// TestFromDropsPreviousVariantKeys: after UnmarshalJSON, the previous
// variant's keys are in AdditionalProperties too. Switching variants drops
// them, and keeps genuine additional properties.
func TestFromDropsPreviousVariantKeys(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"kind": "dog", "bark": "woof", "extra": 7}`), &s))
	require.NoError(t, s.FromCat(Cat{Meow: "purr"}))

	b, err := json.Marshal(s)
	require.NoError(t, err)
	assert.JSONEq(t, `{"kind": "dog", "meow": "purr", "extra": 7}`, string(b))
}
