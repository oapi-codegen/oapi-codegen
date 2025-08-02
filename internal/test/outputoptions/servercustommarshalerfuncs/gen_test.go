package servercustommarshalerfuncs

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type strictServer struct{}

func (s strictServer) JSONExample(_ context.Context, request JSONExampleRequestObject) (JSONExampleResponseObject, error) {
	return JSONExample200JSONResponse(*request.Body), nil
}

type testEncoder struct {
	w io.Writer
}

func newTestEncoder(w io.Writer) *testEncoder {
	return &testEncoder{w: w}
}

func (e *testEncoder) Encode(val any) error {
	_, err := e.w.Write([]byte(`{"value":"100500"}`))

	return err
}

type testDecoder struct {
	r io.Reader
}

func newTestDecoder(r io.Reader) *testDecoder {
	return &testDecoder{r: r}
}

func (d *testDecoder) Decode(val any) error {
	raw, err := io.ReadAll(d.r)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(raw, val); err != nil {
		return err
	}

	req := val.(*JSONExampleJSONRequestBody)
	v := "999999"
	req.Value = &v

	return nil
}

func Test_setJSONEncoder(t *testing.T) {
	defer func() {
		SetJSONEncoder(func(w io.Writer) Encoder {
			return json.NewEncoder(w)
		})
	}()

	val := "test_string"

	request := &JSONExampleJSONRequestBody{
		Value: &val,
	}
	raw, err := json.Marshal(request)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/json", bytes.NewBuffer(raw))

	handler := NewStrictHandlerWithOptions(&strictServer{}, nil, StrictHTTPServerOptions{})
	handler.JSONExample(w, r)

	resp := &JSONExample200JSONResponse{}

	rawResp, err := io.ReadAll(w.Result().Body)
	assert.NoError(t, err)
	assert.NoError(t, json.Unmarshal(rawResp, resp))
	assert.Equal(t, request.Value, resp.Value)

	SetJSONEncoder(func(w io.Writer) Encoder {
		return newTestEncoder(w)
	})

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/json", bytes.NewBuffer(raw))

	handler.JSONExample(w, r)

	resp = &JSONExample200JSONResponse{}

	rawResp, err = io.ReadAll(w.Result().Body)
	assert.NoError(t, err)
	assert.NoError(t, json.Unmarshal(rawResp, resp))
	assert.Equal(t, "100500", *resp.Value)
}

func Test_setJSONDecoder(t *testing.T) {
	defer func() {
		SetJSONDecoder(func(r io.Reader) Decoder {
			return json.NewDecoder(r)
		})
	}()

	val := "test_string"

	request := &JSONExampleJSONRequestBody{
		Value: &val,
	}
	raw, err := json.Marshal(request)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/json", bytes.NewBuffer(raw))

	handler := NewStrictHandlerWithOptions(&strictServer{}, nil, StrictHTTPServerOptions{})
	handler.JSONExample(w, r)

	resp := &JSONExample200JSONResponse{}

	rawResp, err := io.ReadAll(w.Result().Body)
	assert.NoError(t, err)
	assert.NoError(t, json.Unmarshal(rawResp, resp))
	assert.Equal(t, request.Value, resp.Value)

	SetJSONDecoder(func(r io.Reader) Decoder {
		return newTestDecoder(r)
	})

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/json", bytes.NewBuffer(raw))

	handler.JSONExample(w, r)

	resp = &JSONExample200JSONResponse{}

	rawResp, err = io.ReadAll(w.Result().Body)
	assert.NoError(t, err)
	assert.NoError(t, json.Unmarshal(rawResp, resp))
	assert.Equal(t, "999999", *resp.Value)
}
