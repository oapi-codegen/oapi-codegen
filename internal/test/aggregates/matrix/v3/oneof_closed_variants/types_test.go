package matrixv3oneofclosedvariants

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func subject(t *testing.T, document string) Subject {
	t.Helper()
	var s Subject
	require.NoError(t, json.Unmarshal([]byte(document), &s))
	return s
}

// A closed variant allows its own keys and the union's: its properties and
// discriminator.
func TestClosedVariantAllowsTheUnionsKeys(t *testing.T) {
	s := subject(t, `{"kind": "cat", "meow": "purr", "note": "indoor", "collar": {"color": "red"}}`)
	cat, err := s.AsCat()
	require.NoError(t, err)
	assert.Equal(t, "purr", cat.Meow)
	require.NotNil(t, cat.Collar)
	assert.Equal(t, "red", *cat.Collar.Color)

	v, err := s.ValueByDiscriminator()
	require.NoError(t, err)
	assert.IsType(t, Cat{}, v)
}

// A closed variant rejects another variant's data, and a key nothing declares.
func TestClosedVariantRejectsOtherKeys(t *testing.T) {
	s := subject(t, `{"kind": "dog", "bark": "woof"}`)
	_, err := s.AsCat()
	assert.EqualError(t, err, `Cat doesn't allow the property "bark"`)

	s = subject(t, `{"kind": "cat", "meow": "purr", "zebra": 1, "apple": 2}`)
	_, err = s.AsCat()
	assert.EqualError(t, err, `Cat doesn't allow the property "apple"`, "the first unknown key in sorted order")
	_, err = s.ValueByDiscriminator()
	assert.EqualError(t, err, `Cat doesn't allow the property "apple"`)

	s = subject(t, `{"kind": "cat", "meow": "purr", "collar": {"color": "red"}}`)
	_, err = s.AsDog()
	assert.EqualError(t, err, `Dog doesn't allow the property "collar"`)
}

// Only the top level is checked: a nested object without
// additionalProperties: false still takes keys it doesn't declare.
func TestClosedVariantChecksTheTopLevelOnly(t *testing.T) {
	s := subject(t, `{"kind": "cat", "meow": "purr", "collar": {"color": "red", "size": "S"}}`)
	cat, err := s.AsCat()
	require.NoError(t, err)
	assert.Equal(t, "red", *cat.Collar.Color)
}

// A variant without additionalProperties: false ignores keys it doesn't
// declare, as before.
func TestOpenVariantIgnoresOtherKeys(t *testing.T) {
	s := subject(t, `{"kind": "cat", "meow": "purr", "chirp": "tweet"}`)
	bird, err := s.AsBird()
	require.NoError(t, err)
	assert.Equal(t, "tweet", bird.Chirp)
}

// A closed variant reads what its From* wrote, discriminator included.
func TestClosedVariantReadsItsOwnData(t *testing.T) {
	var s Subject
	note := "sleepy"
	s.Note = &note
	require.NoError(t, s.FromDog(Dog{Bark: "woof"}))
	dog, err := s.AsDog()
	require.NoError(t, err)
	assert.Equal(t, "woof", dog.Bark)

	b, err := json.Marshal(s)
	require.NoError(t, err)
	var back Subject
	require.NoError(t, json.Unmarshal(b, &back))
	dog, err = back.AsDog()
	require.NoError(t, err)
	assert.Equal(t, "woof", dog.Bark)
}

// A union holding something other than an object fails to decode as before,
// not on the key check.
func TestClosedVariantOfNonObject(t *testing.T) {
	s := subject(t, `{"kind": "cat", "meow": "purr"}`)
	s.union = json.RawMessage(`"text"`)
	_, err := s.AsCat()
	var typeErr *json.UnmarshalTypeError
	assert.ErrorAs(t, err, &typeErr)
}
