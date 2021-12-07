package server

import (
	"net/http"
	"testing"

	"github.com/deepmap/oapi-codegen/examples/authenticated-api/echo/api"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPI(t *testing.T) {
	e := echo.New()
	s := NewServer()

	fa, err := NewFakeAuthenticator()
	require.NoError(t, err)

	mw, err := CreateMiddleware(fa)
	require.NoError(t, err)
	e.Use(mw...)
	api.RegisterHandlers(e, s)

	// Let's create a JWT with no scopes, which allows access to listing things.
	readerJWT, err := fa.CreateJWSWithClaims([]string{})
	require.NoError(t, err)
	t.Logf("reader jwt: %s", string(readerJWT))

	// Now, create a JWT with write permission.
	writerJWT, err := fa.CreateJWSWithClaims([]string{"things:w"})
	require.NoError(t, err)
	t.Logf("writer jwt: %s", string(writerJWT))

	// ListPets should return 403 forbidden without credentials
	response := testutil.NewRequest().Get("/things").Go(t, e)
	assert.Equal(t, http.StatusForbidden, response.Code())

	// Using the writer JWT should allow us to insert a thing.
	response = testutil.NewRequest().Post("/things").
		WithJWSAuth(string(writerJWT)).
		WithAcceptJson().
		WithJsonBody(api.Thing{Name: "Thing 1"}).Go(t, e)
	require.Equal(t, http.StatusCreated, response.Code())

	// Using the reader JWT should forbid inserting a thing.
	response = testutil.NewRequest().Post("/things").
		WithJWSAuth(string(readerJWT)).
		WithAcceptJson().
		WithJsonBody(api.Thing{Name: "Thing 2"}).Go(t, e)
	require.Equal(t, http.StatusForbidden, response.Code())

	// Both JWT's should allow reading the list of things.
	jwts := []string{string(readerJWT), string(writerJWT)}
	for _, jwt := range jwts {
		response := testutil.NewRequest().Get("/things").
			WithJWSAuth(string(jwt)).
			WithAcceptJson().Go(t, e)
		assert.Equal(t, http.StatusOK, response.Code())
	}
}
