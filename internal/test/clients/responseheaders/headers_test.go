package responseheaders

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeResponse(status int, headers map[string]string, body string) *http.Response {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	for k, v := range headers {
		h.Set(k, v)
	}
	return &http.Response{
		StatusCode: status,
		Header:     h,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

func TestParsesDeclaredResponseHeaders(t *testing.T) {
	rsp, err := ParseGetFooResponse(makeResponse(200, map[string]string{
		"bar":          "bar-value",
		"x-request-id": "42",
	}, `{"foo":"hello"}`))
	require.NoError(t, err)

	require.NotNil(t, rsp.Headers200)
	require.NotNil(t, rsp.Headers200.Bar)
	assert.Equal(t, "bar-value", *rsp.Headers200.Bar)
	assert.Equal(t, 42, rsp.Headers200.XRequestId)
	assert.Nil(t, rsp.Headers404)
	assert.Nil(t, rsp.HeadersDefault)
}

func TestParsesNullableResponseHeader(t *testing.T) {
	rsp, err := ParseGetFooResponse(makeResponse(200, map[string]string{
		"x-request-id":  "42",
		"x-next-cursor": "cursor-value",
	}, `{"foo":"hello"}`))
	require.NoError(t, err)

	require.NotNil(t, rsp.Headers200)
	require.True(t, rsp.Headers200.XNextCursor.IsSpecified())
	assert.Equal(t, "cursor-value", rsp.Headers200.XNextCursor.MustGet())
}

func TestAbsentNullableResponseHeaderIsUnspecified(t *testing.T) {
	rsp, err := ParseGetFooResponse(makeResponse(200, nil, `{"foo":"hello"}`))
	require.NoError(t, err)

	require.NotNil(t, rsp.Headers200)
	assert.False(t, rsp.Headers200.XNextCursor.IsSpecified())
}

func TestParsesComponentResponseHeaders(t *testing.T) {
	rsp, err := ParseGetFooResponse(makeResponse(404, map[string]string{
		"trace-id": "abc",
	}, `{}`))
	require.NoError(t, err)

	require.NotNil(t, rsp.Headers404)
	assert.Equal(t, "abc", rsp.Headers404.TraceId)
	assert.Nil(t, rsp.Headers200)
}

func TestAbsentOptionalHeaderLeavesFieldNil(t *testing.T) {
	rsp, err := ParseGetFooResponse(makeResponse(200, nil, `{"foo":"hello"}`))
	require.NoError(t, err)

	require.NotNil(t, rsp.Headers200)
	assert.Nil(t, rsp.Headers200.Bar)
	// Absent required headers are tolerated (spec-violating server); the
	// field is left at its zero value rather than failing the parse.
	assert.Equal(t, 0, rsp.Headers200.XRequestId)
}

func TestParsesDefaultResponseHeaders(t *testing.T) {
	rsp, err := ParseGetFooResponse(makeResponse(500, map[string]string{
		"retry-after": "3",
	}, ``))
	require.NoError(t, err)

	require.NotNil(t, rsp.HeadersDefault)
	require.NotNil(t, rsp.HeadersDefault.RetryAfter)
	assert.Equal(t, 3, *rsp.HeadersDefault.RetryAfter)
}

func TestMalformedHeaderValueErrors(t *testing.T) {
	_, err := ParseGetFooResponse(makeResponse(200, map[string]string{
		"x-request-id": "not-a-number",
	}, `{"foo":"hello"}`))
	assert.Error(t, err)
}
