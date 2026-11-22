package responseheaderformats

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	expiresAt = time.Date(2026, 8, 7, 11, 0, 43, 0, time.FixedZone("CEST", 2*60*60))
	traceID   = uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
)

type server struct{}

func (server) GetThing(ctx context.Context, request GetThingRequestObject) (GetThingResponseObject, error) {
	var resp GetThing200JSONResponse
	ok := true
	resp.Body.Ok = &ok
	resp.Headers.XExpiresAt = expiresAt
	resp.Headers.XOptionalExpiresAt = &expiresAt
	resp.Headers.XTraceId = traceID
	resp.Headers.XPlain = "plain"
	return resp, nil
}

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	HandlerFromMux(NewStrictHandler(server{}, nil), mux)
	s := httptest.NewServer(mux)
	t.Cleanup(s.Close)
	return s
}

// Before issue #2512 the strict server wrote every response header with
// fmt.Sprint, so a time.Time landed on the wire in Go's default layout
// ("2026-08-07 11:00:43 +0200 CEST") instead of RFC 3339. Headers must be
// styled through the runtime helpers, honouring the schema's format.
func TestResponseHeadersUseSchemaFormat(t *testing.T) {
	s := newServer(t)

	res, err := http.Get(s.URL + "/thing")
	require.NoError(t, err)
	t.Cleanup(func() { _ = res.Body.Close() })

	assert.Equal(t, expiresAt.Format(time.RFC3339Nano), res.Header.Get("X-Expires-At"))
	assert.Equal(t, expiresAt.Format(time.RFC3339Nano), res.Header.Get("X-Optional-Expires-At"))
	assert.Equal(t, traceID.String(), res.Header.Get("X-Trace-Id"))
	assert.Equal(t, "plain", res.Header.Get("X-Plain"))
}

// The client generated from the same spec parses response headers with
// runtime.BindStyledParameterWithOptions, so a strict server that wrote them
// with fmt.Sprint produced output its own client could not read.
func TestGeneratedClientParsesGeneratedServerHeaders(t *testing.T) {
	s := newServer(t)

	c, err := NewClientWithResponses(s.URL)
	require.NoError(t, err)

	res, err := c.GetThingWithResponse(context.Background())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode())
	require.NotNil(t, res.Headers200)

	assert.True(t, expiresAt.Equal(res.Headers200.XExpiresAt),
		"round-tripped X-Expires-At: got %s, want %s", res.Headers200.XExpiresAt, expiresAt)
	require.NotNil(t, res.Headers200.XOptionalExpiresAt)
	assert.True(t, expiresAt.Equal(*res.Headers200.XOptionalExpiresAt))
	assert.Equal(t, traceID, res.Headers200.XTraceId)
	assert.Equal(t, "plain", res.Headers200.XPlain)
}
