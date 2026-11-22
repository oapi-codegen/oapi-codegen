package matrixv3oneofdiscriminatorinline

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Each inline variant takes the value its petType enum pins.
func TestInlineVariantsAreDiscriminated(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"petType": "dog", "bark": "woof"}`), &s))
	v, err := s.ValueByDiscriminator()
	require.NoError(t, err)
	dog, ok := v.(Subject1)
	require.True(t, ok, "%T", v)
	assert.Equal(t, "woof", dog.Bark)

	require.NoError(t, s.FromSubject0(Subject0{Meow: "purr"}))
	discriminator, err := s.Discriminator()
	require.NoError(t, err)
	assert.Equal(t, "cat", discriminator, "From stamps the variant's value")
}

// A variant that pins no value is left to the caller: the discriminator
// doesn't lead to it, and setting it stamps nothing.
func TestUnpinnedVariantIsLeftToTheCaller(t *testing.T) {
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(`{"petType": "finch", "chirp": "tweet"}`), &s))
	_, err := s.ValueByDiscriminator()
	assert.EqualError(t, err, "unknown discriminator value: finch")
	bird, err := s.AsSubject2()
	require.NoError(t, err)
	assert.Equal(t, "tweet", bird.Chirp)

	require.NoError(t, s.FromSubject2(Subject2{PetType: "robin", Chirp: "chirp"}))
	discriminator, err := s.Discriminator()
	require.NoError(t, err)
	assert.Equal(t, "robin", discriminator)
}
