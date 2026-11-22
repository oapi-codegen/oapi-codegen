package optionsnamenormalizerunset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUnsetNormalizer verifies that with no name-normalizer set the default
// Go casing is applied: "uuid" → Uuid, "name" → Name, operationId getHttpPet
// → GetHttpPet on both the client and the server wrapper, and digit "2" is
// not a word boundary so OneOf2things stays OneOf2things.
//
// Sources: outputoptions/name-normalizer/unset
func TestUnsetNormalizer(t *testing.T) {
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
