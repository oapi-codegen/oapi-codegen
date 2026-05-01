package issue518

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubServer struct{}

func (stubServer) GetAlpha(ctx echo.Context) error      { return ctx.NoContent(http.StatusOK) }
func (stubServer) GetBeta(ctx echo.Context) error       { return ctx.NoContent(http.StatusOK) }
func (stubServer) GetGamma(ctx echo.Context) error      { return ctx.NoContent(http.StatusOK) }

// recordingMiddleware appends a tag to the slice it closes over each time it
// runs. Lets the test assert which routes the middleware ran on.
func recordingMiddleware(record *[]string, tag string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			*record = append(*record, tag)
			return next(c)
		}
	}
}

func TestRegisterHandlersWithOptions_PerOperationMiddleware(t *testing.T) {
	var calls []string

	e := echo.New()
	RegisterHandlersWithOptions(e, stubServer{}, RegisterHandlersOptions{
		OperationMiddlewares: map[string][]echo.MiddlewareFunc{
			// Spec-form key (kebab-case), proving the map is NOT keyed on the
			// normalized Go identifier "GetAlpha".
			"get-alpha": {recordingMiddleware(&calls, "alpha-mw")},
		},
	})

	srv := httptest.NewServer(e)
	defer srv.Close()

	for _, path := range []string{"/alpha", "/beta", "/gamma"} {
		resp, err := http.Get(srv.URL + path)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
		require.Equal(t, http.StatusOK, resp.StatusCode, "path %s", path)
	}

	// Middleware fires only on /alpha, not /beta or /gamma.
	assert.Equal(t, []string{"alpha-mw"}, calls)
}

func TestRegisterHandlersWithOptions_FallbackKeyForGeneratedOperationId(t *testing.T) {
	var calls []string

	e := echo.New()
	// /gamma has no spec operationId, so the codegen generated one. Per the
	// MiddlewareKey() fallback, the map key in this case is the normalized
	// OperationId — copy whatever the wrapper method is named.
	RegisterHandlersWithOptions(e, stubServer{}, RegisterHandlersOptions{
		OperationMiddlewares: map[string][]echo.MiddlewareFunc{
			"GetGamma": {recordingMiddleware(&calls, "gamma-mw")},
		},
	})

	srv := httptest.NewServer(e)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/gamma")
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, []string{"gamma-mw"}, calls)
}

func TestRegisterHandlers_BackwardsCompatible(t *testing.T) {
	// Existing call sites using RegisterHandlers / RegisterHandlersWithBaseURL
	// must keep working — they delegate to RegisterHandlersWithOptions with no
	// per-operation middleware.
	e := echo.New()
	RegisterHandlers(e, stubServer{})

	srv := httptest.NewServer(e)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/alpha")
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRegisterHandlersWithBaseURL_BackwardsCompatible(t *testing.T) {
	e := echo.New()
	RegisterHandlersWithBaseURL(e, stubServer{}, "/api")

	srv := httptest.NewServer(e)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/alpha")
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
