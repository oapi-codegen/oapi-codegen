package optionsnamenormalizertocamelcase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestToCamelCaseNormalizer verifies that name-normalizer: ToCamelCase produces
// the same output as unset for this spec: uuid stays Uuid, operationId
// getHttpPet → GetHttpPet on both the client and the server wrapper, and digit
// "2" is not a word boundary (OneOf2things).
//
// Sources: outputoptions/name-normalizer/to-camel-case
func TestToCamelCaseNormalizer(t *testing.T) {
	pet := &Pet{}
	assert.Equal(t, "", pet.Name)
	assert.Equal(t, "", pet.Uuid)

	uri := "https://my-api.com/some-base-url/v1/"
	client, err := NewClient(uri)
	assert.Nil(t, err)
	assert.NotNil(t, client.GetHttpPet)

	server := &ServerInterfaceWrapper{}
	assert.NotNil(t, server.GetHttpPet)

	oneOf := OneOf2things{}
	assert.Zero(t, oneOf)
}
