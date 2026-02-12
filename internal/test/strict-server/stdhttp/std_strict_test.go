//go:build go1.22

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	clientAPI "github.com/oapi-codegen/oapi-codegen/v2/internal/test/strict-server/client"
	"github.com/oapi-codegen/runtime"
	"github.com/oapi-codegen/testutil"
	"github.com/stretchr/testify/assert"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestStdHTTPServer(t *testing.T) {
	server := StrictServer{}
	strictHandler := NewStrictHandler(server, nil)
	m := http.NewServeMux()
	_ = HandlerFromMux(strictHandler, m)
	testImpl(t, m)
}

func TestStreaming(t *testing.T) {
	server := StrictServer{}
	strictHandler := NewStrictHandler(server, nil)
	m := http.NewServeMux()
	handler := HandlerFromMux(strictHandler, m)

	// create a request with a context that will be canceled after 3 second
	req := httptest.NewRequest(http.MethodGet, "/streaming", nil)
	rr := newStreamResponseWriter()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "text/event-stream", rr.Header().Get("Content-Type"))
	assert.Equal(t, 3, len(rr.ops))
	for i := 1; i < len(rr.ops); i++ { // make sure there is at least 10 ms between the writes
		assert.True(t, rr.ops[i].ts-rr.ops[i-1].ts >= 10*time.Millisecond)
	}
	assert.Equal(t, "first write\n", string(rr.ops[0].bytes))
	assert.Equal(t, "second write\n", string(rr.ops[1].bytes))
	assert.Equal(t, "third write\n", string(rr.ops[2].bytes))
}

func testImpl(t *testing.T, handler http.Handler) {
	t.Run("JSONExample", func(t *testing.T) {
		value := "123"
		requestBody := clientAPI.Example{Value: &value}
		rr := testutil.NewRequest().Post("/json").WithJsonBody(requestBody).GoWithHTTPHandler(t, handler).Recorder
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.True(t, strings.HasPrefix(rr.Header().Get("Content-Type"), "application/json"))
		var responseBody clientAPI.Example
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, requestBody, responseBody)
	})
	t.Run("URLEncodedExample", func(t *testing.T) {
		value := "456"
		requestBody := clientAPI.Example{Value: &value}
		requestBodyEncoded, err := runtime.MarshalForm(&requestBody, nil)
		assert.NoError(t, err)
		rr := testutil.NewRequest().Post("/urlencoded").WithContentType("application/x-www-form-urlencoded").WithBody([]byte(requestBodyEncoded.Encode())).GoWithHTTPHandler(t, handler).Recorder
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/x-www-form-urlencoded", rr.Header().Get("Content-Type"))
		values, err := url.ParseQuery(rr.Body.String())
		assert.NoError(t, err)
		var responseBody clientAPI.Example
		err = runtime.BindForm(&responseBody, values, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, requestBody, responseBody)
	})
	t.Run("MultipartExample", func(t *testing.T) {
		value := "789"
		fieldName := "value"
		var writer bytes.Buffer
		mw := multipart.NewWriter(&writer)
		field, err := mw.CreateFormField(fieldName)
		assert.NoError(t, err)
		_, _ = field.Write([]byte(value))
		assert.NoError(t, mw.Close())
		rr := testutil.NewRequest().Post("/multipart").WithContentType(mw.FormDataContentType()).WithBody(writer.Bytes()).GoWithHTTPHandler(t, handler).Recorder
		assert.Equal(t, http.StatusOK, rr.Code)
		contentType, params, err := mime.ParseMediaType(rr.Header().Get("Content-Type"))
		assert.NoError(t, err)
		assert.Equal(t, "multipart/form-data", contentType)
		reader := multipart.NewReader(rr.Body, params["boundary"])
		part, err := reader.NextPart()
		assert.NoError(t, err)
		assert.Equal(t, part.FormName(), fieldName)
		readValue, err := io.ReadAll(part)
		assert.NoError(t, err)
		assert.Equal(t, value, string(readValue))
		_, err = reader.NextPart()
		assert.Equal(t, io.EOF, err)
	})
	t.Run("MultipartRelatedExample", func(t *testing.T) {
		value := "789"
		fieldName := "value"
		var writer bytes.Buffer
		mw := multipart.NewWriter(&writer)
		field, err := mw.CreateFormField(fieldName)
		assert.NoError(t, err)
		_, _ = field.Write([]byte(value))
		assert.NoError(t, mw.Close())
		rr := testutil.NewRequest().Post("/multipart-related").WithContentType(mime.FormatMediaType("multipart/related", map[string]string{"boundary": mw.Boundary()})).WithBody(writer.Bytes()).GoWithHTTPHandler(t, handler).Recorder
		assert.Equal(t, http.StatusOK, rr.Code)
		contentType, params, err := mime.ParseMediaType(rr.Header().Get("Content-Type"))
		assert.NoError(t, err)
		assert.Equal(t, "multipart/related", contentType)
		reader := multipart.NewReader(rr.Body, params["boundary"])
		part, err := reader.NextPart()
		assert.NoError(t, err)
		assert.Equal(t, part.FormName(), fieldName)
		readValue, err := io.ReadAll(part)
		assert.NoError(t, err)
		assert.Equal(t, value, string(readValue))
		_, err = reader.NextPart()
		assert.Equal(t, io.EOF, err)
	})
	t.Run("TextExample", func(t *testing.T) {
		value := "text"
		rr := testutil.NewRequest().Post("/text").WithContentType("text/plain").WithBody([]byte(value)).GoWithHTTPHandler(t, handler).Recorder
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
		assert.Equal(t, value, rr.Body.String())
	})
	t.Run("UnknownExample", func(t *testing.T) {
		data := []byte("unknown data")
		rr := testutil.NewRequest().Post("/unknown").WithContentType("image/png").WithBody(data).GoWithHTTPHandler(t, handler).Recorder
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "video/mp4", rr.Header().Get("Content-Type"))
		assert.Equal(t, data, rr.Body.Bytes())
	})
	t.Run("MultipleRequestAndResponseTypesJSON", func(t *testing.T) {
		value := "123"
		requestBody := clientAPI.Example{Value: &value}
		rr := testutil.NewRequest().Post("/multiple").WithJsonBody(requestBody).GoWithHTTPHandler(t, handler).Recorder
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.True(t, strings.HasPrefix(rr.Header().Get("Content-Type"), "application/json"))
		var responseBody clientAPI.Example
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, requestBody, responseBody)
	})
	t.Run("MultipleRequestAndResponseTypesFormdata", func(t *testing.T) {
		value := "456"
		requestBody := clientAPI.Example{Value: &value}
		requestBodyEncoded, err := runtime.MarshalForm(&requestBody, nil)
		assert.NoError(t, err)
		rr := testutil.NewRequest().Post("/multiple").WithContentType("application/x-www-form-urlencoded").WithBody([]byte(requestBodyEncoded.Encode())).GoWithHTTPHandler(t, handler).Recorder
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/x-www-form-urlencoded", rr.Header().Get("Content-Type"))
		values, err := url.ParseQuery(rr.Body.String())
		assert.NoError(t, err)
		var responseBody clientAPI.Example
		err = runtime.BindForm(&responseBody, values, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, requestBody, responseBody)
	})
	t.Run("MultipleRequestAndResponseTypesMultipart", func(t *testing.T) {
		value := "789"
		fieldName := "value"
		var writer bytes.Buffer
		mw := multipart.NewWriter(&writer)
		field, err := mw.CreateFormField(fieldName)
		assert.NoError(t, err)
		_, _ = field.Write([]byte(value))
		assert.NoError(t, mw.Close())
		rr := testutil.NewRequest().Post("/multiple").WithContentType(mw.FormDataContentType()).WithBody(writer.Bytes()).GoWithHTTPHandler(t, handler).Recorder
		assert.Equal(t, http.StatusOK, rr.Code)
		contentType, params, err := mime.ParseMediaType(rr.Header().Get("Content-Type"))
		assert.NoError(t, err)
		assert.Equal(t, "multipart/form-data", contentType)
		reader := multipart.NewReader(rr.Body, params["boundary"])
		part, err := reader.NextPart()
		assert.NoError(t, err)
		assert.Equal(t, part.FormName(), fieldName)
		readValue, err := io.ReadAll(part)
		assert.NoError(t, err)
		assert.Equal(t, value, string(readValue))
		_, err = reader.NextPart()
		assert.Equal(t, io.EOF, err)
	})
	t.Run("MultipleRequestAndResponseTypesText", func(t *testing.T) {
		value := "text"
		rr := testutil.NewRequest().Post("/multiple").WithContentType("text/plain").WithBody([]byte(value)).GoWithHTTPHandler(t, handler).Recorder
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
		assert.Equal(t, value, rr.Body.String())
	})
	t.Run("MultipleRequestAndResponseTypesImage", func(t *testing.T) {
		data := []byte("unknown data")
		rr := testutil.NewRequest().Post("/multiple").WithContentType("image/png").WithBody(data).GoWithHTTPHandler(t, handler).Recorder
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "image/png", rr.Header().Get("Content-Type"))
		assert.Equal(t, data, rr.Body.Bytes())
	})
	t.Run("HeadersExample", func(t *testing.T) {
		header1 := "value1"
		header2 := "890"
		value := "asdf"
		requestBody := clientAPI.Example{Value: &value}
		rr := testutil.NewRequest().Post("/with-headers").WithHeader("header1", header1).WithHeader("header2", header2).WithJsonBody(requestBody).GoWithHTTPHandler(t, handler).Recorder
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.True(t, strings.HasPrefix(rr.Header().Get("Content-Type"), "application/json"))
		var responseBody clientAPI.Example
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, requestBody, responseBody)
		assert.Equal(t, header1, rr.Header().Get("header1"))
		assert.Equal(t, header2, rr.Header().Get("header2"))
	})
	t.Run("UnspecifiedContentType", func(t *testing.T) {
		data := []byte("image data")
		contentType := "image/jpeg"
		rr := testutil.NewRequest().Post("/unspecified-content-type").WithContentType(contentType).WithBody(data).GoWithHTTPHandler(t, handler).Recorder
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, contentType, rr.Header().Get("Content-Type"))
		assert.Equal(t, data, rr.Body.Bytes())
	})
	t.Run("ReusableResponses", func(t *testing.T) {
		value := "jkl;"
		requestBody := clientAPI.Example{Value: &value}
		rr := testutil.NewRequest().Post("/reusable-responses").WithJsonBody(requestBody).GoWithHTTPHandler(t, handler).Recorder
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.True(t, strings.HasPrefix(rr.Header().Get("Content-Type"), "application/json"))
		var responseBody clientAPI.Example
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, requestBody, responseBody)
	})
	t.Run("UnionResponses", func(t *testing.T) {
		value := "union"
		requestBody := clientAPI.Example{Value: &value}
		rr := testutil.NewRequest().Post("/with-union").WithJsonBody(requestBody).GoWithHTTPHandler(t, handler).Recorder
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.True(t, strings.HasPrefix(rr.Header().Get("Content-Type"), "application/json"))
		var responseBody clientAPI.Example
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		assert.NoError(t, err)
		assert.Equal(t, requestBody, responseBody)
	})
	t.Run("StreamingResponse", func(t *testing.T) {
		// the /streaming endpoint will issue two writes with 100ms between them.
		rr := testutil.NewRequest().Get("/streaming").GoWithHTTPHandler(t, handler).Recorder
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "text/event-stream", rr.Header().Get("Content-Type"))
		fmt.Println("body length", len(rr.Body.String()))
	})
}
