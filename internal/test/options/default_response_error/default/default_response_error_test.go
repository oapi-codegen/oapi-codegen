package optionsdefaultresponseerrordefault

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

// outputoptions/default-response-error/default: with the option unset, an
// unmatched response is returned with no error, which is the pre-existing
// behaviour the option is opt-in to.
func TestUnmatchedResponseIsNotAnErrorByDefault(t *testing.T) {
	resp, err := ParseGetThingResponse(newResponse(http.StatusInternalServerError, `{"message":"boom"}`))

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Nil(t, resp.JSON200)
	assert.Nil(t, resp.JSON404)
}
