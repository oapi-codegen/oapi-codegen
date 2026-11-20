package matrixv2nullableunionnesteddiscriminator

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

// TestNestedNullableUnionDiscriminator: the inline oneOf the null branch wraps
// keeps its discriminator, so From* stamps it and ValueByDiscriminator reads
// it (issue #2565).
func TestNestedNullableUnionDiscriminator(t *testing.T) {
	var s Subject
	require.NoError(t, s.FromDog(Dog{Bark: "woof"}))
	b, err := json.Marshal(s)
	require.NoError(t, err)
	assert.JSONEq(t, `{"petType": "dog", "bark": "woof"}`, string(b))

	v, err := s.ValueByDiscriminator()
	require.NoError(t, err)
	assert.Equal(t, Dog{PetType: "dog", Bark: "woof"}, v)
}
