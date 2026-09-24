package matrixv2oneofuntypedproperty

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

// TestFromWithUntypedProperty: the union's own Meta is an interface holding
// nil before the first From*, which reflect cannot ask IsZero of.
func TestFromWithUntypedProperty(t *testing.T) {
	var s Subject
	require.NotPanics(t, func() {
		require.NoError(t, s.FromCat(Cat{Meow: "purr", Meta: map[string]any{"k": float64(1)}}))
	})
	s.Id = "1"

	b, err := json.Marshal(s)
	require.NoError(t, err)
	assert.JSONEq(t, `{"id": "1", "meow": "purr", "meta": {"k": 1}}`, string(b))

	require.NotPanics(t, func() {
		require.NoError(t, s.MergeDog(Dog{Bark: "woof"}))
	})
}
