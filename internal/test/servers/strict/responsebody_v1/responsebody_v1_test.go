package serversstrictresponsebodyv1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Handler code written against the direct response types still compiles.
var (
	_ GetAllOfResponseObject        = GetAllOf200JSONResponse{Untyped: "value"}
	_ GetAllOfSiblingResponseObject = GetAllOfSibling200JSONResponse{Untyped: "value"}
	_ GetWrappedResponseObject      = GetWrapped200JSONResponse(Wrapped{Untyped: "value"})
)

type strictServer struct {
	body any
}

func (s strictServer) GetAllOf(context.Context, GetAllOfRequestObject) (GetAllOfResponseObject, error) {
	return GetAllOf200JSONResponse{Untyped: s.body}, nil
}

func (s strictServer) GetAllOfSibling(context.Context, GetAllOfSiblingRequestObject) (GetAllOfSiblingResponseObject, error) {
	return GetAllOfSibling200JSONResponse{Untyped: s.body}, nil
}

func (s strictServer) GetWrapped(context.Context, GetWrappedRequestObject) (GetWrappedResponseObject, error) {
	return GetWrapped200JSONResponse{Untyped: s.body}, nil
}

func (s strictServer) GetNullable(context.Context, GetNullableRequestObject) (GetNullableResponseObject, error) {
	return GetNullable200JSONResponse{Body: s.body}, nil
}

func serve(t *testing.T, body any, path string) *httptest.ResponseRecorder {
	t.Helper()
	handler := Handler(NewStrictHandler(strictServer{body: body}, nil))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

func TestDirectResponsesServeJSON(t *testing.T) {
	for _, path := range []string{"/allof", "/allof-sibling", "/wrapped"} {
		t.Run(path, func(t *testing.T) {
			w := serve(t, "value", path)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.True(t, json.Valid(w.Body.Bytes()), "body %q is not JSON", w.Body.String())
		})
	}
}

func TestNullableResponseBody(t *testing.T) {
	for _, tt := range []struct {
		name string
		body any
	}{
		{name: "object", body: map[string]any{"value": "object"}},
		{name: "null", body: nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := serve(t, tt.body, "/nullable")
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			var got any
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
			assert.Equal(t, tt.body, got)
		})
	}
}
