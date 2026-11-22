package matrixv1nullableunionnested

// This test records how schema-merging-behavior v1 handles this shape. A bug fix
// may change it, but it must not be changed to accept a regression: code that
// generated and worked before must keep doing so. A commit that changes it must
// say why.

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNestedNullableUnion: the null branch makes the field a nil pointer, and
// the inline oneOf it wraps is the union, with its accessors and JSON methods
// (issue #2565).
func TestNestedNullableUnion(t *testing.T) {
	var h Holder
	require.NoError(t, json.Unmarshal([]byte(`{"one": null}`), &h))
	assert.Nil(t, h.One)
	b, err := json.Marshal(h)
	require.NoError(t, err)
	assert.JSONEq(t, `{"one": null}`, string(b))

	require.NoError(t, json.Unmarshal([]byte(`{"one": {"meow": "purr"}}`), &h))
	require.NotNil(t, h.One)
	cat, err := h.One.AsCat()
	require.NoError(t, err)
	assert.Equal(t, Cat{Meow: "purr"}, cat)
	b, err = json.Marshal(h)
	require.NoError(t, err)
	assert.JSONEq(t, `{"one": {"meow": "purr"}}`, string(b))
}
