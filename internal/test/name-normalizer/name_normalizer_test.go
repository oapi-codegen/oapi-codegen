package name_normalizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const hostname = "http://host"

func TestGenCodeHasCorrectNamesWithInitialisms(t *testing.T) {
	pet := &Pet{}
	assert.Equal(t, "", pet.Name)
	assert.Equal(t, "", pet.UUID)

	uri := "https://my-api.com/some-base-url/v1/"
	client, err := NewClient(uri)
	assert.Nil(t, err)
	assert.NotNil(t, client.GetHTTPPet)

	server := &ServerInterfaceWrapper{}
	assert.NotNil(t, server.GetHTTPPet)
}
