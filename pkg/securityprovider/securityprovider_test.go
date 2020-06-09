package securityprovider

import (
	"testing"

	"github.com/tidepool-org/oapi-codegen/internal/test/client"
	"github.com/stretchr/testify/assert"
)

var (
	withTrailingSlash string = "https://my-api.com/some-base-url/v1/"
)

func TestSecurityProviders(t *testing.T) {
	bearer, err := NewSecurityProviderBearerToken("mytoken")
	assert.NoError(t, err)
	client1, err := client.NewClient(
		withTrailingSlash,
		client.WithRequestEditorFn(bearer.Intercept),
	)
	assert.NoError(t, err)

	apiKey, err := NewSecurityProviderApiKey("cookie", "apikey", "mykey")
	assert.NoError(t, err)
	client2, err := client.NewClient(
		withTrailingSlash,
		client.WithRequestEditorFn(apiKey.Intercept),
	)
	assert.NoError(t, err)

	basicAuth, err := NewSecurityProviderBasicAuth("username", "password")
	assert.NoError(t, err)
	client3, err := client.NewClient(
		withTrailingSlash,
		client.WithRequestEditorFn(basicAuth.Intercept),
	)
	assert.NoError(t, err)

	assert.Equal(t, withTrailingSlash, client1.Server)
	assert.Equal(t, withTrailingSlash, client2.Server)
	assert.Equal(t, withTrailingSlash, client3.Server)
}
