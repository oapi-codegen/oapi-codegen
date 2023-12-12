package issue1381_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	issue1381 "github.com/deepmap/oapi-codegen/v2/internal/test/issues/issue-1381"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type testStrictServerInterface struct {
	t *testing.T
}

// (GET /test)
func (s *testStrictServerInterface) Test(ctx context.Context, request issue1381.TestRequestObject) (issue1381.TestResponseObject, error) {
	jsonOk := false
	binaryOk := false
	for {
		if p, err := request.Body.NextPart(); err == io.EOF {
			break
		} else {
			assert.NoError(s.t, err)
			switch p.Header.Get("Content-Type") {
			case "application/json":
				var t issue1381.Test
				err := json.NewDecoder(p).Decode(&t)
				assert.NoError(s.t, err)
				assert.Equal(s.t, "request1", t.Field1)
				assert.Equal(s.t, "request2", t.Field2)
				jsonOk = true
			case "application/octet-stream":
				buf, err := io.ReadAll(p)
				assert.NoError(s.t, err)
				assert.Equal(s.t, []byte("REQUEST-BINARY"), buf)
				binaryOk = true
			default:
				assert.Fail(s.t, "Bad Content-Type: %s", p.Header.Get("Content-Type"))
			}
		}
	}

	assert.True(s.t, jsonOk)
	assert.True(s.t, binaryOk)
	return issue1381.Test200MultipartResponse(func(writer *multipart.Writer) error {
		p, err := writer.CreatePart(textproto.MIMEHeader{
			"Content-Type": []string{"application/json"},
		})
		assert.NoError(s.t, err)
		err = json.NewEncoder(p).Encode(issue1381.Test{
			Field1: "response1",
			Field2: "response2",
		})
		assert.NoError(s.t, err)
		p, err = writer.CreatePart(textproto.MIMEHeader{
			"Content-Type": []string{"application/octet-stream"},
		})
		assert.NoError(s.t, err)
		_, err = io.WriteString(p, "RESPONSE-BINARY")
		assert.NoError(s.t, err)
		return nil
	}), nil
}

func TestIssue1381(t *testing.T) {
	g := gin.Default()
	issue1381.RegisterHandlersWithOptions(g,
		issue1381.NewStrictHandler(&testStrictServerInterface{t: t}, nil),
		issue1381.GinServerOptions{})
	ts := httptest.NewServer(g)
	defer ts.Close()

	c, err := issue1381.NewClientWithResponses(ts.URL)
	assert.NoError(t, err)
	body := bytes.NewBuffer(nil)
	w := multipart.NewWriter(body)
	p, err := w.CreatePart(textproto.MIMEHeader{
		"Content-Type": []string{"application/json"},
	})
	assert.NoError(t, err)
	err = json.NewEncoder(p).Encode(issue1381.Test{
		Field1: "request1",
		Field2: "request2",
	})
	assert.NoError(t, err)
	p, err = w.CreatePart(textproto.MIMEHeader{
		"Content-Type": []string{"application/octet-stream"},
	})
	assert.NoError(t, err)
	_, err = io.WriteString(p, "REQUEST-BINARY")
	assert.NoError(t, err)
	w.Close()

	res, err := c.TestWithBodyWithResponse(
		context.TODO(),
		mime.FormatMediaType("multipart/related", map[string]string{
			"boundary": w.Boundary(),
		}),
		body,
	)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
	mediaType, params, err := mime.ParseMediaType(res.HTTPResponse.Header.Get("Content-Type"))
	if assert.NoError(t, err) {
		assert.Equal(t, "multipart/related", mediaType)
		assert.NotEmpty(t, params["boundary"])
		reader := multipart.NewReader(bytes.NewReader(res.Body), params["boundary"])
		jsonOk := false
		binaryOk := false
		for {
			if p, err := reader.NextPart(); err == io.EOF {
				break
			} else {
				assert.NoError(t, err)
				switch p.Header.Get("Content-Type") {
				case "application/json":
					var j issue1381.Test
					err := json.NewDecoder(p).Decode(&j)
					assert.NoError(t, err)
					assert.Equal(t, "response1", j.Field1)
					assert.Equal(t, "response2", j.Field2)
					jsonOk = true
				case "application/octet-stream":
					buf, err := io.ReadAll(p)
					assert.NoError(t, err)
					assert.Equal(t, []byte("RESPONSE-BINARY"), buf)
					binaryOk = true
				default:
					assert.Fail(t, "Bad Content-Type: %s", p.Header.Get("Content-Type"))
				}
			}
		}
		assert.True(t, jsonOk)
		assert.True(t, binaryOk)
	}
}
