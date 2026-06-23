package issue2190

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetTest401TextResponse verifies that the generated VisitGetTestResponse
// method on GetTest401TextResponse produces a valid text/plain 401 response.
// This is a regression test for https://github.com/oapi-codegen/oapi-codegen/issues/2190
// where the generated code tried to do []byte(response) on a struct type,
// which does not compile.
func TestGetTest401TextResponse(t *testing.T) {
	resp := GetTest401TextResponse("Unauthorized")
	w := httptest.NewRecorder()

	err := resp.VisitGetTestResponse(w)
	require.NoError(t, err)
	assert.Equal(t, 401, w.Code)
	assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
	assert.Equal(t, "Unauthorized", w.Body.String())
}

// TestGetTest200JSONResponse verifies that the 200 JSON response path also works.
func TestGetTest200JSONResponse(t *testing.T) {
	resp := GetTest200JSONResponse{SuccessJSONResponse("hello")}
	w := httptest.NewRecorder()

	err := resp.VisitGetTestResponse(w)
	require.NoError(t, err)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "hello")
}
