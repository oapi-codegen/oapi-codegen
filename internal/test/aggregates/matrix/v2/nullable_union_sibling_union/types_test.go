package matrixv2nullableunionsiblingunion

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

// TestSiblingUnionKeepsVariants: the nullable oneOf collapses into its object
// branch, whose property is a field, and the anyOf beside it keeps its
// variants and their accessors (PR #2566).
func TestSiblingUnionKeepsVariants(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"meow": "purr", "id": "1"}`), &s))
	require.NotNil(t, s.Id)
	assert.Equal(t, "1", *s.Id)
	cat, err := s.AsCat()
	require.NoError(t, err)
	assert.Equal(t, Cat{Meow: "purr"}, cat)

	require.NoError(t, s.FromDog(Dog{Bark: "woof"}))
	b, err := json.Marshal(s)
	require.NoError(t, err)
	assert.JSONEq(t, `{"bark": "woof", "id": "1"}`, string(b))
}
