package api

import (
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oapi-codegen/testutil"
)

// validatingHandler wraps handler in a request validator built directly on
// kin-openapi, the way the oapi-codegen validation middlewares are, so the
// test does not depend on those modules.
func validatingHandler(t *testing.T, handler http.Handler, options openapi3filter.Options) http.Handler {
	t.Helper()
	spec, err := GetSwagger()
	require.NoError(t, err)
	// The spec declares an absolute server URL; drop it so the router matches
	// on the path alone.
	spec.Servers = nil
	router, err := gorillamux.NewRouter(spec)
	require.NoError(t, err)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route, pathParams, err := router.FindRoute(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		err = openapi3filter.ValidateRequest(r.Context(), &openapi3filter.RequestValidationInput{
			Request:    r,
			PathParams: pathParams,
			Route:      route,
			Options:    &options,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func newStrictServer() http.Handler {
	e := echo.New()
	RegisterHandlers(e, NewStrictHandler(StrictServer{}, nil))
	return e
}

// TestEchoServer_BindsValidatedBody checks that, behind a validator that
// keeps its validated value, the strict handler receives exactly the value
// the validator saw: the framework binder would match "VALUE" to the `value`
// field case-insensitively and bind the unvalidated member instead.
func TestEchoServer_BindsValidatedBody(t *testing.T) {
	const body = `{"value":"123","VALUE":"456"}`

	t.Run("with KeepValidatedValue", func(t *testing.T) {
		handler := validatingHandler(t, newStrictServer(), openapi3filter.Options{KeepValidatedValue: true})
		rr := testutil.NewRequest().Post("/json").WithContentType("application/json").WithBody([]byte(body)).GoWithHTTPHandler(t, handler).Recorder
		require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
		assert.JSONEq(t, `{"value":"123"}`, rr.Body.String())
	})

	t.Run("without KeepValidatedValue the body is decoded as before", func(t *testing.T) {
		handler := validatingHandler(t, newStrictServer(), openapi3filter.Options{})
		rr := testutil.NewRequest().Post("/json").WithContentType("application/json").WithBody([]byte(body)).GoWithHTTPHandler(t, handler).Recorder
		require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
		// The framework binder's case-insensitive matching lets the later
		// "VALUE" member win; this is the pre-existing behaviour the fallback keeps.
		assert.JSONEq(t, `{"value":"456"}`, rr.Body.String())
	})
}
