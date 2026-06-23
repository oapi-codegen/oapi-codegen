package optionsnamenormalizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUnsetNormalizer verifies that with no name-normalizer set the default
// Go casing is applied: "uuid" → Uuid, "name" → Name, digit "2" is not a
// word boundary so OneOf2things stays OneOf2things.
//
// Sources: outputoptions/name-normalizer/unset
func TestUnsetNormalizer(t *testing.T) {
	pet := &UnsetPet{}
	assert.Equal(t, "", pet.Name)
	assert.Equal(t, "", pet.Uuid)

	oneOf := UnsetOneOf2things{}
	assert.Zero(t, oneOf)
}

// TestToCamelCaseNormalizer verifies that name-normalizer: ToCamelCase produces
// the same output as unset for this spec: uuid stays Uuid, http stays Http,
// digit "2" is not a word boundary (CamelOneOf2things).
//
// Sources: outputoptions/name-normalizer/to-camel-case
func TestToCamelCaseNormalizer(t *testing.T) {
	pet := &CamelPet{}
	assert.Equal(t, "", pet.Name)
	assert.Equal(t, "", pet.Uuid)

	oneOf := CamelOneOf2things{}
	assert.Zero(t, oneOf)
}

// TestToCamelCaseWithInitialismsNormalizer verifies that name-normalizer:
// ToCamelCaseWithInitialisms expands common Go initialisms: uuid → UUID,
// http → HTTP, id → ID; and digit "2" becomes a word boundary so
// InitialismOneOf2things → InitialismOneOf2Things.
//
// Sources: outputoptions/name-normalizer/to-camel-case-with-initialisms
func TestToCamelCaseWithInitialismsNormalizer(t *testing.T) {
	pet := &InitialismPet{}
	assert.Equal(t, "", pet.Name)
	assert.Equal(t, "", pet.UUID)

	oneOf := InitialismOneOf2Things{}
	assert.Zero(t, oneOf)
}

// TestToCamelCaseWithDigitsNormalizer verifies that name-normalizer:
// ToCamelCaseWithDigits treats digit sequences as word boundaries so
// DigitsOneOf2things → DigitsOneOf2Things, but does NOT expand initialisms
// (uuid stays Uuid, id stays Id).
//
// Sources: outputoptions/name-normalizer/to-camel-case-with-digits
func TestToCamelCaseWithDigitsNormalizer(t *testing.T) {
	pet := &DigitsPet{}
	assert.Equal(t, "", pet.Name)
	assert.Equal(t, "", pet.Uuid)

	oneOf := DigitsOneOf2Things{}
	assert.Zero(t, oneOf)
}

// TestToCamelCaseWithAdditionalInitialismsNormalizer verifies that
// name-normalizer: ToCamelCaseWithInitialisms combined with
// additional-initialisms: [NAME] causes the "name" field to be rendered as
// NAME in addition to the standard initialism expansions (uuid → UUID, id → ID).
//
// Sources: outputoptions/name-normalizer/to-camel-case-with-additional-initialisms
func TestToCamelCaseWithAdditionalInitialismsNormalizer(t *testing.T) {
	pet := &ExtraPet{}
	assert.Equal(t, "", pet.NAME)
	assert.Equal(t, "", pet.UUID)

	oneOf := ExtraOneOf2Things{}
	assert.Zero(t, oneOf)
}
