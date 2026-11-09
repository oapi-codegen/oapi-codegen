//go:build go1.22

package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	chiAPI "github.com/oapi-codegen/oapi-codegen/v2/internal/test/servers/strict/chi"
	gorillaAPI "github.com/oapi-codegen/oapi-codegen/v2/internal/test/servers/strict/gorilla"
	stdAPI "github.com/oapi-codegen/oapi-codegen/v2/internal/test/servers/strict/stdhttp"
	"github.com/stretchr/testify/assert"
)

// Applications can share these aliases without importing a generated package.
type sharedHandler = func(context.Context, http.ResponseWriter, *http.Request, any) (any, error)
type sharedMiddleware = func(sharedHandler, string) sharedHandler

// Slices of generated aliases must also be interchangeable across packages.
var (
	_ []chiAPI.StrictMiddlewareFunc     = []stdAPI.StrictMiddlewareFunc(nil)
	_ []gorillaAPI.StrictMiddlewareFunc = []stdAPI.StrictMiddlewareFunc(nil)
	_ []sharedMiddleware                = []stdAPI.StrictMiddlewareFunc(nil)
)

type gorillaServer struct {
	gorillaAPI.StrictServerInterface
}

func (gorillaServer) TextExample(_ context.Context, request gorillaAPI.TextExampleRequestObject) (gorillaAPI.TextExampleResponseObject, error) {
	return gorillaAPI.TextExample200TextResponse(*request.Body), nil
}

func TestSharedStrictMiddleware(t *testing.T) {
	var operations []string
	middlewares := []sharedMiddleware{
		func(next sharedHandler, operationID string) sharedHandler {
			return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
				operations = append(operations, operationID)
				w.Header().Set("X-Shared-Middleware", "called")
				return next(ctx, w, r, request)
			}
		},
	}

	for _, tt := range []struct {
		name    string
		handler http.Handler
	}{
		{"std-http", stdAPI.Handler(stdAPI.NewStrictHandler(stdAPI.StrictServer{}, middlewares))},
		{"std-http-with-options", stdAPI.Handler(stdAPI.NewStrictHandlerWithOptions(stdAPI.StrictServer{}, middlewares, stdAPI.StrictHTTPServerOptions{}))},
		{"chi", chiAPI.Handler(chiAPI.NewStrictHandler(chiAPI.StrictServer{}, middlewares))},
		{"chi-with-options", chiAPI.Handler(chiAPI.NewStrictHandlerWithOptions(chiAPI.StrictServer{}, middlewares, chiAPI.StrictHTTPServerOptions{}))},
		{"gorilla", gorillaAPI.Handler(gorillaAPI.NewStrictHandler(gorillaServer{}, middlewares))},
		{"gorilla-with-options", gorillaAPI.Handler(gorillaAPI.NewStrictHandlerWithOptions(gorillaServer{}, middlewares, gorillaAPI.StrictHTTPServerOptions{}))},
	} {
		t.Run(tt.name, func(t *testing.T) {
			operations = nil
			request := httptest.NewRequest(http.MethodPost, "/text", strings.NewReader("shared middleware"))
			request.Header.Set("Content-Type", "text/plain")
			response := httptest.NewRecorder()
			tt.handler.ServeHTTP(response, request)

			assert.Equal(t, http.StatusOK, response.Code)
			assert.Equal(t, "shared middleware", response.Body.String())
			assert.Equal(t, "called", response.Header().Get("X-Shared-Middleware"))
			assert.Equal(t, []string{"TextExample"}, operations)
		})
	}
}
