package textnonstring

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegerTextResponse is the regression test for issue-1897: a text/plain
// response with an integer schema must write the decimal representation of the
// value, not fail to compile and not emit the Unicode code point.
func TestIntegerTextResponse(t *testing.T) {
	resp := GetPing201TextResponse(201)
	w := httptest.NewRecorder()

	require.NoError(t, resp.VisitGetPingResponse(w))
	assert.Equal(t, 201, w.Code)
	assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
	assert.Equal(t, "201", w.Body.String())
}

// TestBooleanTextResponse covers the other non-string-convertible primitive.
func TestBooleanTextResponse(t *testing.T) {
	resp := GetPing202TextResponse(true)
	w := httptest.NewRecorder()

	require.NoError(t, resp.VisitGetPingResponse(w))
	assert.Equal(t, 202, w.Code)
	assert.Equal(t, "true", w.Body.String())
}

// TestStringTextResponse confirms the common string case is unchanged.
func TestStringTextResponse(t *testing.T) {
	resp := GetPing200TextResponse("pong")
	w := httptest.NewRecorder()

	require.NoError(t, resp.VisitGetPingResponse(w))
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}
