//go:build go1.22

package stdhttp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testServer struct {
	receivedParam string
}

func (s *testServer) GetResource(w http.ResponseWriter, r *http.Request, addressingIdentifier string) {
	s.receivedParam = addressingIdentifier
	_, _ = fmt.Fprint(w, addressingIdentifier)
}

func TestDashedPathParam(t *testing.T) {
	server := &testServer{}
	handler := Handler(server)

	req := httptest.NewRequest(http.MethodGet, "/resources/my-value", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "expected 200 OK, got %d; body: %s", rec.Code, rec.Body.String())
	assert.Equal(t, "my-value", server.receivedParam, "path parameter was not correctly extracted")
}
