package server

import (
	"net/http"
	"testing"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/authenticated-api/stdhttp/api"
	"github.com/oapi-codegen/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPI(t *testing.T) {
	r := http.NewServeMux()
	s := NewServer()

	fa, err := NewFakeAuthenticator()
	require.NoError(t, err)

	mw, err := CreateMiddleware(fa)
	require.NoError(t, err)

	h := api.HandlerFromMux(s, r)
	// wrap the existing handler with our global middleware
	h = mw(h)

	// Let's create a JWT with no scopes, which allows access to listing things.
	readerJWT, err := fa.CreateJWSWithClaims([]string{})
	require.NoError(t, err)
	t.Logf("reader jwt: %s", string(readerJWT))

	// Now, create a JWT with write permission.
	writerJWT, err := fa.CreateJWSWithClaims([]string{"things:w"})
	require.NoError(t, err)
	t.Logf("writer jwt: %s", string(writerJWT))

	// ListPets should return 401 Unauthorized without credentials
	response := testutil.NewRequest().Get("/things").GoWithHTTPHandler(t, h)
	assert.Equal(t, http.StatusUnauthorized, response.Code())

	// Using the writer JWT should allow us to insert a thing.
	response = testutil.NewRequest().Post("/things").
		WithJWSAuth(string(writerJWT)).
		WithAcceptJson().
		WithJsonBody(api.Thing{Name: "Thing 1"}).GoWithHTTPHandler(t, h)
	require.Equal(t, http.StatusCreated, response.Code())

	// Using the reader JWT should forbid inserting a thing.
	response = testutil.NewRequest().Post("/things").
		WithJWSAuth(string(readerJWT)).
		WithAcceptJson().
		WithJsonBody(api.Thing{Name: "Thing 2"}).GoWithHTTPHandler(t, h)
	require.Equal(t, http.StatusUnauthorized, response.Code())

	// Both JWT's should allow reading the list of things.
	jwts := []string{string(readerJWT), string(writerJWT)}
	for _, jwt := range jwts {
		response := testutil.NewRequest().Get("/things").
			WithJWSAuth(jwt).
			WithAcceptJson().GoWithHTTPHandler(t, h)
		assert.Equal(t, http.StatusOK, response.Code())
	}
}
