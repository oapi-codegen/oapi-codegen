package optionsnamenormalizertocamelcasewithdigits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestToCamelCaseWithDigitsNormalizer verifies that name-normalizer:
// ToCamelCaseWithDigits treats digit sequences as word boundaries so
// OneOf2things → OneOf2Things, but does NOT expand initialisms (uuid stays
// Uuid, and operationId getHttpPet → GetHttpPet on both the client and the
// server wrapper).
//
// Sources: outputoptions/name-normalizer/to-camel-case-with-digits
func TestToCamelCaseWithDigitsNormalizer(t *testing.T) {
	pet := &Pet{}
	assert.Equal(t, "", pet.Name)
	assert.Equal(t, "", pet.Uuid)

	uri := "https://my-api.com/some-base-url/v1/"
	client, err := NewClient(uri)
	assert.Nil(t, err)
	assert.NotNil(t, client.GetHttpPet)

	server := &ServerInterfaceWrapper{}
	assert.NotNil(t, server.GetHttpPet)

	oneOf := OneOf2Things{}
	assert.Zero(t, oneOf)
}
