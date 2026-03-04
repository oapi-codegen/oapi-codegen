package tocamelcasewithdigits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenCodeHasCorrectNames(t *testing.T) {
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
