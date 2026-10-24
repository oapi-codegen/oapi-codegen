package optionsunexpectedresponseskipped

import (
	"bytes"
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

// outputoptions/unexpected-response/skipped: with the option unset (the default),
// an undeclared response yields a response object and a nil error, so existing
// behaviour is preserved.
func TestUnexpectedResponseHasNoError(t *testing.T) {
	res, err := ParseGetThingResponse(newResponse(http.StatusInternalServerError, "application/json", `{"oops":true}`))
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Nil(t, res.JSON200)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode())
}
