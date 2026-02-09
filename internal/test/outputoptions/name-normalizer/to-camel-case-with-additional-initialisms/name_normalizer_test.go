package tocamelcasewithadditionalinitialisms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenCodeHasCorrectNames(t *testing.T) {
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
