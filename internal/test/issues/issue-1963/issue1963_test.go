package issue1963

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests are regression tests for https://github.com/oapi-codegen/oapi-codegen/issues/1963.
//
// The issue: in generated strict server Visit*Response functions, WriteHeader
// was called before marshalling the response body. If JSON encoding failed,
// ResponseErrorHandlerFunc could not set a different status code because 200
// was already written.
//
// The fix: JSON responses are buffered via bytes.Buffer before headers are sent.
// Non-JSON responses retain the original headers-first ordering.

// TestJsonEndpoint_Success verifies JSON responses buffer the body and set
// Content-Type and status code correctly.
func TestJsonEndpoint_Success(t *testing.T) {
	value := "hello"
	resp := JsonEndpoint200JSONResponse{Value: &value}
	w := httptest.NewRecorder()

	err := resp.VisitJsonEndpointResponse(w)
	require.NoError(t, err)
	assert.Equal(t, 200, w.Code)
	assert.True(t, strings.HasPrefix(w.Header().Get("Content-Type"), "application/json"))

	var body Response
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.Equal(t, &value, body.Value)
}

// TestTextEndpoint_Success verifies text responses set Content-Type and status code.
func TestTextEndpoint_Success(t *testing.T) {
	resp := TextEndpoint200TextResponse("hello world")
	w := httptest.NewRecorder()

	err := resp.VisitTextEndpointResponse(w)
	require.NoError(t, err)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
	assert.Equal(t, "hello world", w.Body.String())
}

// TestFormdataEndpoint_Success verifies formdata responses set Content-Type and status code.
func TestFormdataEndpoint_Success(t *testing.T) {
	value := "test"
	resp := FormdataEndpoint200FormdataResponse{Value: &value}
	w := httptest.NewRecorder()

	err := resp.VisitFormdataEndpointResponse(w)
	require.NoError(t, err)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "application/x-www-form-urlencoded", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "value=test")
}

// TestMultipartEndpoint_Success verifies multipart responses set Content-Type and status code.
func TestMultipartEndpoint_Success(t *testing.T) {
	resp := MultipartEndpoint200MultipartResponse(func(writer *multipart.Writer) error {
		return writer.WriteField("field", "value")
	})
	w := httptest.NewRecorder()

	err := resp.VisitMultipartEndpointResponse(w)
	require.NoError(t, err)
	assert.Equal(t, 200, w.Code)
	ct, params, err := mime.ParseMediaType(w.Header().Get("Content-Type"))
	require.NoError(t, err)
	assert.Equal(t, "multipart/form-data", ct)

	reader := multipart.NewReader(w.Body, params["boundary"])
	part, err := reader.NextPart()
	require.NoError(t, err)
	assert.Equal(t, "field", part.FormName())
	data, err := io.ReadAll(part)
	require.NoError(t, err)
	assert.Equal(t, "value", string(data))
}

// TestBinaryEndpoint_Success verifies binary (io.Reader) responses set Content-Type and status code.
func TestBinaryEndpoint_Success(t *testing.T) {
	body := strings.NewReader("binary data")
	resp := BinaryEndpoint200ApplicationoctetStreamResponse{
		Body:          body,
		ContentLength: int64(len("binary data")),
	}
	w := httptest.NewRecorder()

	err := resp.VisitBinaryEndpointResponse(w)
	require.NoError(t, err)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "application/octet-stream", w.Header().Get("Content-Type"))
	assert.Equal(t, "11", w.Header().Get("Content-Length"))
	assert.Equal(t, "binary data", w.Body.String())
}

// TestJsonEndpoint_EncodingError_ErrorHandlerCanSetStatus is the core
// regression test. It verifies that when JSON encoding fails, the
// ResponseErrorHandlerFunc can still set the HTTP status code because nothing
// has been written to the ResponseWriter yet.
//
// We use a custom StrictServerInterface implementation that returns a response
// object whose Visit method will fail during JSON encoding.
func TestJsonEndpoint_EncodingError_ErrorHandlerCanSetStatus(t *testing.T) {
	server := &errorServer{}
	var errHandlerCalled bool
	handler := NewStrictHandlerWithOptions(server, nil, StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			errHandlerCalled = true
			w.WriteHeader(http.StatusInternalServerError)
		},
	})
	mux := http.NewServeMux()
	HandlerFromMux(handler, mux)

	body := `{"value": "test"}`
	req := httptest.NewRequest(http.MethodPost, "/json", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.True(t, errHandlerCalled, "ResponseErrorHandlerFunc should have been called")
	assert.Equal(t, http.StatusInternalServerError, w.Code,
		"error handler should be able to set status code to 500 when JSON encoding fails")
	assert.Empty(t, w.Header().Get("Content-Type"),
		"Content-Type should not be set when encoding fails before headers are written")
}

// errorServer implements StrictServerInterface and returns a response whose
// JSON encoding will fail.
type errorServer struct{}

func (s *errorServer) JsonEndpoint(_ context.Context, _ JsonEndpointRequestObject) (JsonEndpointResponseObject, error) {
	return &unmarshalableJSONResponse{}, nil
}

func (s *errorServer) TextEndpoint(_ context.Context, _ TextEndpointRequestObject) (TextEndpointResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *errorServer) FormdataEndpoint(_ context.Context, _ FormdataEndpointRequestObject) (FormdataEndpointResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *errorServer) MultipartEndpoint(_ context.Context, _ MultipartEndpointRequestObject) (MultipartEndpointResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *errorServer) BinaryEndpoint(_ context.Context, _ BinaryEndpointRequestObject) (BinaryEndpointResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}

// unmarshalableJSONResponse implements JsonEndpointResponseObject with a Visit
// method that follows the exact same pattern as the generated code but
// encodes a value that json.Encoder cannot marshal (a channel).
type unmarshalableJSONResponse struct{}

func (u *unmarshalableJSONResponse) VisitJsonEndpointResponse(w http.ResponseWriter) error {
	var buf bytes.Buffer
	// Channels cannot be JSON-encoded; this will return an error.
	if err := json.NewEncoder(&buf).Encode(map[string]any{
		"bad": make(chan int),
	}); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_, err := buf.WriteTo(w)
	return err
}
