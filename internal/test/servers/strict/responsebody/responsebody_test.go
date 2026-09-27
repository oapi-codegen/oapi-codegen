package serversstrictresponsebody

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type responseBodyStrictServer struct {
	body any
}

func (s responseBodyStrictServer) GetUnion(context.Context, GetUnionRequestObject) (GetUnionResponseObject, error) {
	return GetUnion200JSONResponse{Body: s.body}, nil
}

func (s responseBodyStrictServer) GetNull(context.Context, GetNullRequestObject) (GetNullResponseObject, error) {
	return GetNull200JSONResponse{Body: s.body}, nil
}

func (s responseBodyStrictServer) GetNullableRef(context.Context, GetNullableRefRequestObject) (GetNullableRefResponseObject, error) {
	return GetNullableRef200JSONResponse{Body: s.body}, nil
}

func TestOpenAPI31StrictResponseBody(t *testing.T) {
	tests := []struct {
		name string
		path string
		body any
	}{
		{name: "union string", path: "/union", body: "value"},
		{name: "union number", path: "/union", body: 1.5},
		{name: "union boolean", path: "/union", body: true},
		{name: "null", path: "/null", body: nil},
		{name: "nullable ref object", path: "/nullable-ref", body: map[string]any{"value": "object"}},
		{name: "nullable ref null", path: "/nullable-ref", body: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := Handler(NewStrictHandler(responseBodyStrictServer{body: tt.body}, nil))
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if w.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
			}
			if got := w.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}
			var body any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode JSON response: %v", err)
			}
			if !reflect.DeepEqual(body, tt.body) {
				t.Errorf("JSON body = %#v, want %#v", body, tt.body)
			}
		})
	}
}
