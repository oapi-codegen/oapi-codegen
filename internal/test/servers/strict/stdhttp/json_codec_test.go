//go:build go1.22

package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clientAPI "github.com/oapi-codegen/oapi-codegen/v2/internal/test/servers/strict/client"
	"github.com/oapi-codegen/runtime"
	"github.com/oapi-codegen/testutil"
)

// countingCodec counts the calls that reach it. It doesn't implement
// runtime.StreamingJSONCodec, so bodies go through Marshal and Unmarshal too.
type countingCodec struct {
	runtime.StdJSONCodec
	marshals, unmarshals *int
}

func (c countingCodec) Marshal(v any) ([]byte, error) {
	*c.marshals++
	return c.StdJSONCodec.Marshal(v)
}

func (c countingCodec) Unmarshal(data []byte, v any) error {
	*c.unmarshals++
	return c.StdJSONCodec.Unmarshal(data, v)
}

// TestJSONCodec checks that a codec set with runtime.SetJSONCodec is the one
// generated code uses on every step of a JSON round trip: the client's request
// body, the strict server's request binding and response, and the client's
// response parsing.
func TestJSONCodec(t *testing.T) {
	var marshals, unmarshals int
	runtime.SetJSONCodec(countingCodec{marshals: &marshals, unmarshals: &unmarshals})
	t.Cleanup(func() { runtime.SetJSONCodec(nil) })

	m := http.NewServeMux()
	HandlerFromMux(NewStrictHandler(StrictServer{}, nil), m)
	srv := httptest.NewServer(m)
	t.Cleanup(srv.Close)

	client, err := clientAPI.NewClientWithResponses(srv.URL)
	require.NoError(t, err)

	value := "123"
	body := clientAPI.Example{Value: &value}
	rsp, err := client.JSONExampleWithResponse(context.Background(), body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rsp.StatusCode())
	require.NotNil(t, rsp.JSON200)
	assert.Equal(t, body, *rsp.JSON200)

	assert.Equal(t, 2, marshals, "client request body and server response")
	assert.Equal(t, 2, unmarshals, "server request body and client response")
}

// streamingCodec implements runtime.StreamingJSONCodec on top of countingCodec.
type streamingCodec struct {
	countingCodec
	writes, reads *int
}

func (c streamingCodec) MarshalWrite(w io.Writer, v any) error {
	*c.writes++
	return json.NewEncoder(w).Encode(v)
}

func (c streamingCodec) UnmarshalRead(r io.Reader, v any) error {
	*c.reads++
	return json.NewDecoder(r).Decode(v)
}

// TestStreamingJSONCodec checks that the strict server hands a codec that
// implements runtime.StreamingJSONCodec the request body and the response as
// streams, and doesn't go through the codec's Marshal and Unmarshal for them.
func TestStreamingJSONCodec(t *testing.T) {
	var marshals, unmarshals, writes, reads int
	runtime.SetJSONCodec(streamingCodec{
		countingCodec: countingCodec{marshals: &marshals, unmarshals: &unmarshals},
		writes:        &writes,
		reads:         &reads,
	})
	t.Cleanup(func() { runtime.SetJSONCodec(nil) })

	m := http.NewServeMux()
	HandlerFromMux(NewStrictHandler(StrictServer{}, nil), m)

	// The request is built by hand, so that only the server reaches the codec.
	rr := testutil.NewRequest().Post("/json").WithContentType("application/json").WithBody([]byte(`{"value":"123"}`)).GoWithHTTPHandler(t, m).Recorder
	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	assert.JSONEq(t, `{"value":"123"}`, rr.Body.String())

	assert.Equal(t, 1, reads, "request body")
	assert.Equal(t, 1, writes, "response")
	assert.Zero(t, marshals)
	assert.Zero(t, unmarshals)
}
