package matrixv3nullableunionnesteddiscriminator

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
