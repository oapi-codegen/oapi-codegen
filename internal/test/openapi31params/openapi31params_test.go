package openapi31params

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureServer records what the generated binding layer delivered to the
// handler, so tests can assert on the bound dynamic types.
type captureServer struct {
	id     any
	params UpdateThingParams
	body   any
}

func (s *captureServer) UpdateThing(w http.ResponseWriter, r *http.Request, id any, params UpdateThingParams) {
	s.id = id
	s.params = params
	s.body = nil
	_ = json.NewDecoder(r.Body).Decode(&s.body)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Echo", fmt.Sprint(id))
	_ = json.NewEncoder(w).Encode(map[string]any{"id": id})
}

// A union parameter binds to the first member that parses, in specificity
// order (boolean, integer, number, string), in every position the runtime
// binds: path, query, and header.
func TestUnionParamMemberSelection(t *testing.T) {
	capture := &captureServer{}
	srv := httptest.NewServer(Handler(capture))
	defer srv.Close()

	post := func(t *testing.T, path string, header map[string]string) {
		t.Helper()
		req, err := http.NewRequest(http.MethodPost, srv.URL+path, bytes.NewReader([]byte(`"hello"`)))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		for k, v := range header {
			req.Header.Set(k, v)
		}
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// Numeric tokens bind as numbers, not strings.
	post(t, "/things/42?filter=1.5&limit=10", map[string]string{"X-Trace": "7"})
	assert.Equal(t, int64(42), capture.id, "path [string, integer]: integer member wins for an integer token")
	assert.Equal(t, float64(1.5), capture.params.Filter, "query [string, number, boolean]: number member takes the fractional token")
	assert.Equal(t, int64(10), capture.params.Limit, `query [integer, string, "null"]: null is stripped, integer member wins`)
	assert.Equal(t, int64(7), capture.params.XTrace, "header [integer, string]: integer member wins")
	assert.Equal(t, "hello", capture.body, "union request body arrives as the JSON value")

	// Non-numeric tokens fall to the string member; JSON booleans take the
	// boolean member.
	post(t, "/things/abc?filter=true&limit=007", nil)
	assert.Equal(t, "abc", capture.id, "path: string member for a non-numeric token")
	assert.Equal(t, true, capture.params.Filter, "query: boolean member for a JSON boolean literal")
	assert.Equal(t, "007", capture.params.Limit, "query: leading zeros are not a JSON number, string member wins")
	assert.Nil(t, capture.params.XTrace, "absent optional header stays nil")

	// Absent optional parameters stay nil.
	post(t, "/things/1", nil)
	assert.Equal(t, int64(1), capture.id)
	assert.Nil(t, capture.params.Filter)
	assert.Nil(t, capture.params.Limit)
}

// The generated client serializes union (`any`) parameter values by
// reflecting on the concrete value, so typed values round-trip through the
// wire back into the same dynamic types on the server.
func TestUnionParamClientRoundTrip(t *testing.T) {
	capture := &captureServer{}
	srv := httptest.NewServer(Handler(capture))
	defer srv.Close()

	client, err := NewClient(srv.URL)
	require.NoError(t, err)

	resp, err := client.UpdateThing(context.Background(), 42,
		&UpdateThingParams{Filter: 1.5, XTrace: "span-1"},
		UpdateThingJSONRequestBody(3.25))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, int64(42), capture.id, "client int 42 serializes to '42', binds back as int64")
	assert.Equal(t, float64(1.5), capture.params.Filter)
	assert.Equal(t, "span-1", capture.params.XTrace, "client string round-trips as string")
	assert.Nil(t, capture.params.Limit, "unset optional union param is omitted from the request")
	assert.Equal(t, 3.25, capture.body, "union JSON body round-trips")

	var echoed map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&echoed))
	assert.Equal(t, float64(42), echoed["id"], "response body is plain JSON, so the id comes back as a JSON number")
}

// The generated ClientWithResponses binds declared response headers through
// the styled binder with Types, so a union-typed header comes back as the
// first member that parses.
func TestUnionResponseHeaderClientBinding(t *testing.T) {
	capture := &captureServer{}
	srv := httptest.NewServer(Handler(capture))
	defer srv.Close()

	client, err := NewClientWithResponses(srv.URL)
	require.NoError(t, err)

	resp, err := client.UpdateThingWithResponse(context.Background(), 42, nil,
		UpdateThingJSONRequestBody("x"))
	require.NoError(t, err)
	require.NotNil(t, resp.Headers200)
	assert.Equal(t, int64(42), resp.Headers200.XEcho, "header [integer, string]: integer member wins for a numeric token")

	resp, err = client.UpdateThingWithResponse(context.Background(), "abc", nil,
		UpdateThingJSONRequestBody("x"))
	require.NoError(t, err)
	require.NotNil(t, resp.Headers200)
	assert.Equal(t, "abc", resp.Headers200.XEcho, "non-numeric token falls to the string member")
}
