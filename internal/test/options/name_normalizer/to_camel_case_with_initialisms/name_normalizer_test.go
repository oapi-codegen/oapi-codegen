package optionsnamenormalizertocamelcasewithinitialisms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestToCamelCaseWithInitialismsNormalizer verifies that name-normalizer:
// ToCamelCaseWithInitialisms expands common Go initialisms: uuid → UUID and
// operationId getHttpPet → GetHTTPPet on both the client and the server
// wrapper; and digit "2" becomes a word boundary so OneOf2things →
// OneOf2Things.
//
// Sources: outputoptions/name-normalizer/to-camel-case-with-initialisms
func TestToCamelCaseWithInitialismsNormalizer(t *testing.T) {
	pet := &Pet{}
	assert.Equal(t, "", pet.Name)
	assert.Equal(t, "", pet.UUID)

	uri := "https://my-api.com/some-base-url/v1/"
	client, err := NewClient(uri)
	assert.Nil(t, err)
	assert.NotNil(t, client.GetHTTPPet)

	server := &ServerInterfaceWrapper{}
	assert.NotNil(t, server.GetHTTPPet)

	oneOf := OneOf2Things{}
	assert.Zero(t, oneOf)
}
