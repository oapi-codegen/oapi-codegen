//go:build go1.22

package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clientAPI "github.com/oapi-codegen/oapi-codegen/v2/internal/test/servers/strict/client"
	"github.com/oapi-codegen/runtime"
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
