package issue1262_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	issue1262 "github.com/oapi-codegen/oapi-codegen/v2/internal/test/issues/issue-1262"
	"github.com/stretchr/testify/require"
)

// DummyServer implements ServerInterface to illustrate the desired behaviour.
type DummyServer struct {
	ParsedUpload *issue1262.PostAttachmentMultipartRequestBody
}

// PostAttachment parses the request body, storing the result in ParsedUpload
// for the test to inspect.
func (d *DummyServer) PostAttachment(ctx echo.Context) error {
	var requestBody issue1262.PostAttachmentMultipartRequestBody
	if err := ctx.Bind(&requestBody); err != nil {
		return err
	}

	d.ParsedUpload = &requestBody
	return nil
}

func Test_EchoBindToMultipartForDataBody(t *testing.T) {
	var server DummyServer
	e := echo.New()
	issue1262.RegisterHandlers(e, &server)

	requestBody := strings.NewReader(`--DELIMITER
Content-Disposition: form-data; name="description"

Movie script draft
--DELIMITER
Content-Disposition: form-data; name="data"; filename="script.txt"

Once upon a time in a galaxy far, far away
--DELIMITER
`)
	req := httptest.NewRequest(http.MethodPost, "/attachment", requestBody)
	req.Header.Set(echo.HeaderContentType, echo.MIMEMultipartForm+"; boundary=DELIMITER")
	resp := httptest.NewRecorder()
	e.ServeHTTP(resp, req)

	require.NotNil(t, server.ParsedUpload)

	t.Run("parses form values", func(t *testing.T) {
		require.NotNil(t, server.ParsedUpload.Description)
		require.Equal(t, "Movie script draft", *server.ParsedUpload.Description)
	})

	/* Disabled: echo doesn't recognise openapi_types.File as something that
		   a multipart.Reader---see echo.DefaultBinder.bindData.
	t.Run("parses form files", func(t *testing.T) {
		assert.Equal(t, "script.txt", server.ParsedUpload.Data.Filename())
		dataBytes, err := server.ParsedUpload.Data.Bytes()
		require.NoError(t, err)
		assert.Equal(t, "Once upon a time in a galaxy far, far away\n", string(dataBytes))
	})
	*/
}
