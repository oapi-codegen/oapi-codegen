package client

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"

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

type InterceptMock struct {
	mock.Mock
}

func (intr *InterceptMock) Before(r *http.Request) {
	intr.Called(r)
}
func (intr *InterceptMock) After(resp *http.Response) {
	intr.Called(resp)
}

func newInterceptMiddleware(interceptor *InterceptMock) RoundTripMiddleware {
	return func(next DoFn) DoFn {
		return func(r *http.Request) (*http.Response, error) {
			interceptor.Before(r)
			resp, err := next(r)
			interceptor.After(resp)
			return resp, err
		}
	}
}

func TestSharedRoundTripMiddleware(t *testing.T) {
	interceptor := new(InterceptMock)
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

	interceptor.On("Before", mock.Anything).
		Run(func(args mock.Arguments) {
			httpReq := args.Get(0).(*http.Request)
			assert.Equal(t, httpReq.URL.String(), "https://my-url.com/with_both_responses")
			assert.Empty(t, httpReq.Body)
		}).
		Once()
	interceptor.On("After", mock.Anything).Once()

	clientWithResponses, err := NewClientWithResponses(
		"https://my-url.com",
		WithHTTPClient(httpClient),
		WithSharedRoundTripMiddleware(newInterceptMiddleware(interceptor)),
	)
	require.NoError(t, err)

	resp, err := clientWithResponses.GetBothWithResponse(context.TODO())
	require.NoError(t, err)

	assert.Equal(t, "some-role", resp.JSON200.Role)
	assert.Equal(t, "first-name", resp.JSON200.FirstName)

	interceptor.AssertExpectations(t)
	httpClient.AssertExpectations(t)
}

func TestOperationRoundTripMiddleware(t *testing.T) {

	returnSuccess := func(c *HTTPClientMock) {
		c.
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
			}, nil).
			Once()
	}

	t.Run("should not exectue middleware on different operation", func(t *testing.T) {
		interceptor := new(InterceptMock)
		httpClient := new(HTTPClientMock)

		returnSuccess(httpClient)

		// Interceptor should never be called because it's on a different operation
		// interceptor.On("Before", mock.Anything).Once()
		// interceptor.On("After", mock.Anything).Once()

		clientWithResponses, err := NewClientWithResponses(
			"https://my-url.com",
			WithHTTPClient(httpClient),
			WithRoundTripMiddlewares(RoundTripMiddlewares{
				GetJson: newInterceptMiddleware(interceptor),
			}),
		)
		require.NoError(t, err)

		resp, err := clientWithResponses.GetBothWithResponse(context.TODO())
		require.NoError(t, err)
		assert.Equal(t, "some-role", resp.JSON200.Role)
		assert.Equal(t, "first-name", resp.JSON200.FirstName)

		interceptor.AssertExpectations(t)
		httpClient.AssertExpectations(t)

	})
	t.Run("should exectue middleware on operation", func(t *testing.T) {
		interceptor := new(InterceptMock)
		httpClient := new(HTTPClientMock)

		returnSuccess(httpClient)

		interceptor.On("Before", mock.Anything).Once()
		interceptor.On("After", mock.Anything).Once()

		clientWithResponses, err := NewClientWithResponses(
			"https://my-url.com",
			WithHTTPClient(httpClient),
			WithRoundTripMiddlewares(RoundTripMiddlewares{
				GetBoth: newInterceptMiddleware(interceptor),
			}),
		)
		require.NoError(t, err)

		resp, err := clientWithResponses.GetBothWithResponse(context.TODO())
		require.NoError(t, err)

		assert.Equal(t, "some-role", resp.JSON200.Role)
		assert.Equal(t, "first-name", resp.JSON200.FirstName)

		interceptor.AssertExpectations(t)
		httpClient.AssertExpectations(t)
	})
}

func TestTemp(t *testing.T) {

	var (
		withTrailingSlash    string = "https://my-api.com/some-base-url/v1/"
		withoutTrailingSlash string = "https://my-api.com/some-base-url/v1"
	)

	client1, err := NewClient(
		withTrailingSlash,
	)
	assert.NoError(t, err)

	client2, err := NewClient(
		withoutTrailingSlash,
	)
	assert.NoError(t, err)

	client3, err := NewClient(
		"",
		WithBaseURL(withTrailingSlash),
	)
	assert.NoError(t, err)

	client4, err := NewClient(
		"",
		WithBaseURL(withoutTrailingSlash),
	)
	assert.NoError(t, err)

	expectedURL := withTrailingSlash

	assert.Equal(t, expectedURL, client1.Server)
	assert.Equal(t, expectedURL, client2.Server)
	assert.Equal(t, expectedURL, client3.Server)
	assert.Equal(t, expectedURL, client4.Server)
}
