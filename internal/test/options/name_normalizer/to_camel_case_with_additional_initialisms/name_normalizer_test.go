package optionsnamenormalizertocamelcasewithadditionalinitialisms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestToCamelCaseWithAdditionalInitialismsNormalizer verifies that
// name-normalizer: ToCamelCaseWithInitialisms combined with
// additional-initialisms: [NAME] causes the "name" field to be rendered as
// NAME in addition to the standard initialism expansions (uuid → UUID,
// getHttpPet → GetHTTPPet on both the client and the server wrapper), with
// digit "2" treated as a word boundary (OneOf2Things).
//
// Sources: outputoptions/name-normalizer/to-camel-case-with-additional-initialisms
func TestToCamelCaseWithAdditionalInitialismsNormalizer(t *testing.T) {
	pet := &Pet{}
	assert.Equal(t, "", pet.NAME)
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
