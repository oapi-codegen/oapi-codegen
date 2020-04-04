package pkg_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	pkg "github.com/deepmap/oapi-codegen/internal/test/client/private"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type HTTPClientMock struct {
	mock.Mock
}

func (c *HTTPClientMock) Do(req *http.Request) (*http.Response, error) {
	args := c.Called(req)

	arg1, arg2 := args.Get(0), args.Error(1)
	if arg1 == nil {
		return nil, arg2
	}
	return arg1.(*http.Response), arg2
}

func TestManualClientSuccess(t *testing.T) {
	httpClient := new(HTTPClientMock)

	httpClient.
		On("Do", mock.Anything).
		Return(&http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
				"role": "some-role",
				"firstName": "first-name"
			}`))),
		}, nil)

	client, err := pkg.NewClient(
		"https://myapi.com/v1",
		pkg.WithHTTPClient(httpClient),
	)
	require.NoError(t, err)
	assert.NotNil(t, client)

	resp, err := client.PostJSON(context.TODO(), pkg.PostJsonJSONRequestBody{
		Role:      "some-role",
		FirstName: "first-name",
	})
	require.NoError(t, err)
	assert.NotNil(t, resp)

	assert.Equal(t, "some-role", resp.Role)
	assert.Equal(t, "first-name", resp.FirstName)
}

func TestManualClientError(t *testing.T) {
	httpClient := new(HTTPClientMock)

	httpClient.
		On("Do", mock.Anything).
		Return(&http.Response{
			Status:     "422 Unprocessable Entity",
			StatusCode: 422,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
				"httpStatusCode": 422,
				"errorCode": "SOME_ERROR",
				"message": "something went wrong :("
			}`))),
		}, nil)

	client, err := pkg.NewClient(
		"https://myapi.com/v1",
		pkg.WithHTTPClient(httpClient),
	)
	require.NoError(t, err)
	assert.NotNil(t, client)

	resp, err := client.PostJSON(context.TODO(), pkg.PostJsonJSONRequestBody{})
	require.Error(t, err)
	assert.Nil(t, resp)

	errorResp, ok := err.(*pkg.Error)
	require.Truef(t, ok, "PostJSON error should be of type '*pkg.Error'. instead got: %v", err.Error())
	assert.Equal(t, 422, errorResp.HttpStatusCode)
	assert.Equal(t, "SOME_ERROR", errorResp.ErrorCode)
	assert.Equal(t, "something went wrong :(", errorResp.Message)
}
