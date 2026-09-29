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

type strictServer struct {
	body any
}

func (s strictServer) GetAllOf(context.Context, GetAllOfRequestObject) (GetAllOfResponseObject, error) {
	return GetAllOf200JSONResponse{Body: s.body}, nil
}

func (s strictServer) GetAllOfSibling(context.Context, GetAllOfSiblingRequestObject) (GetAllOfSiblingResponseObject, error) {
	return GetAllOfSibling200JSONResponse{Body: s.body}, nil
}

func (s strictServer) GetWrapped(context.Context, GetWrappedRequestObject) (GetWrappedResponseObject, error) {
	return GetWrapped200JSONResponse{Body: s.body}, nil
}

func (s strictServer) GetNullable(context.Context, GetNullableRequestObject) (GetNullableResponseObject, error) {
	return GetNullable200JSONResponse{Body: s.body}, nil
}

// Each response sends its value itself, not an object keyed by the type.
func TestResponsesServeTheirValue(t *testing.T) {
	for _, path := range []string{"/allof", "/allof-sibling", "/wrapped", "/nullable"} {
		for _, tt := range []struct {
			name string
			body any
		}{
			{name: "string", body: "value"},
			{name: "object", body: map[string]any{"value": "object"}},
			{name: "null", body: nil},
		} {
			t.Run(path+"/"+tt.name, func(t *testing.T) {
				handler := Handler(NewStrictHandler(strictServer{body: tt.body}, nil))
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))

				assert.Equal(t, http.StatusOK, w.Code)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				var got any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
				assert.Equal(t, tt.body, got)
			})
		}
	}
}
