package optionsunexpectedresponseenabled

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newResponse(status int, contentType, body string) *http.Response {
	h := http.Header{}
	if contentType != "" {
		h.Set("Content-Type", contentType)
	}
	return &http.Response{
		StatusCode: status,
		Header:     h,
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
	}
}

// outputoptions/unexpected-response/enabled: a response the spec does not declare
// (here a 500) hits the generated default case and returns ErrUnexpectedResponse.
func TestUnexpectedResponseReturnsSentinel(t *testing.T) {
	res, err := ParseGetThingResponse(newResponse(http.StatusInternalServerError, "application/json", `{"oops":true}`))
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnexpectedResponse), "expected ErrUnexpectedResponse, got %v", err)
	assert.Nil(t, res)
}

// A declared response (200) is still parsed normally, with no error.
func TestExpectedResponseParsesWithoutError(t *testing.T) {
	res, err := ParseGetThingResponse(newResponse(http.StatusOK, "application/json", `{"id":"1","name":"rock"}`))
	require.NoError(t, err)
	require.NotNil(t, res.JSON200)
	assert.Equal(t, "rock", res.JSON200.Name)
}
