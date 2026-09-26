package matrixv3oneofclosedparentlistschildren

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A child closed in its own allOf member is closed as a whole. The parent's
// oneOf, which lists it, doesn't lend it the other children's keys.
func TestClosedChildRejectsItsSiblingsKeys(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"petType": "Dog", "bark": "b"}`), &s))
	_, err := s.AsCat()
	assert.EqualError(t, err, `Cat doesn't allow the property "bark"`)
	v, err := s.ValueByDiscriminator()
	require.NoError(t, err)
	dog, ok := v.(Dog)
	require.True(t, ok, "%T", v)
	assert.Equal(t, "b", *dog.Bark)
	assert.Equal(t, "Dog", dog.PetType)
}
