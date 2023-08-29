package client

import (
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/stretchr/testify/assert"
)

var (
	withTrailingSlash = "https://my-api.com/some-base-url/v1/"
)

func TestTemp(t *testing.T) {

	var (
		withTrailingSlash    = "https://my-api.com/some-base-url/v1/"
		withoutTrailingSlash = "https://my-api.com/some-base-url/v1"
	)

	client1, err := NewClient(
		withTrailingSlash,
	)
	assert.NoError(t, err)

	client2, err := NewClient(
		withoutTrailingSlash,
	)
	assert.NoError(t, err)

	client3, err := NewClient(
		"",
		WithBaseURL(withTrailingSlash),
	)
	assert.NoError(t, err)

	client4, err := NewClient(
		"",
		WithBaseURL(withoutTrailingSlash),
	)
	assert.NoError(t, err)

	expectedURL := withTrailingSlash

	assert.Equal(t, expectedURL, client1.Server)
	assert.Equal(t, expectedURL, client2.Server)
	assert.Equal(t, expectedURL, client3.Server)
	assert.Equal(t, expectedURL, client4.Server)
}

func TestSecurityProviders(t *testing.T) {
	bearer, err := securityprovider.NewSecurityProviderBearerToken("mytoken")
	assert.NoError(t, err)
	client1, err := NewClient(
		withTrailingSlash,
		WithRequestEditorFn(bearer.Intercept),
	)
	assert.NoError(t, err)

	apiKey, err := securityprovider.NewSecurityProviderApiKey("cookie", "apikey", "mykey")
	assert.NoError(t, err)
	client2, err := NewClient(
		withTrailingSlash,
		WithRequestEditorFn(apiKey.Intercept),
	)
	assert.NoError(t, err)

	basicAuth, err := securityprovider.NewSecurityProviderBasicAuth("username", "password")
	assert.NoError(t, err)
	client3, err := NewClient(
		withTrailingSlash,
		WithRequestEditorFn(basicAuth.Intercept),
	)
	assert.NoError(t, err)

	assert.Equal(t, withTrailingSlash, client1.Server)
	assert.Equal(t, withTrailingSlash, client2.Server)
	assert.Equal(t, withTrailingSlash, client3.Server)
}
