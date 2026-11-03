package optionsdefaultresponseerrorenabled

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

// outputoptions/default-response-error/enabled: a declared response still parses
// normally, so enabling the option does not change the matched paths.
func TestDeclaredResponseStillParses(t *testing.T) {
	resp, err := ParseGetThingResponse(newResponse(http.StatusOK, `{"id":"thing-1"}`))
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.JSON200)
	assert.Equal(t, "thing-1", resp.JSON200.Id)
}

// outputoptions/default-response-error/enabled: a response that matches no
// declared case must return an error naming the status code.
func TestUnmatchedResponseIsAnError(t *testing.T) {
	resp, err := ParseGetThingResponse(newResponse(http.StatusInternalServerError, `{"message":"boom"}`))

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "unexpected response status: 500")
}
