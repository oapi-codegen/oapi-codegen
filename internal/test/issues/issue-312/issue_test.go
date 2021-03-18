package issue_312

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const hostname = "host"

func TestClient_WhenPathHasColon_RequestHasCorrectPath(t *testing.T) {
	t.Parallel()
	doer := &HTTPRequestDoerMock{}
	client, _ := NewClientWithResponses(hostname, WithHTTPClient(doer))
	_ = client

	doer.On("Do", mock.Anything).Return(nil, errors.New("something went wrong")).Run(func(args mock.Arguments) {
		req, ok := args.Get(0).(*http.Request)
		assert.True(t, ok)
		assert.Equal(t, "/host/pets:validate", req.URL.Path)
	})

	client.ValidatePetsWithResponse(context.Background(), ValidatePetsJSONRequestBody{
		Names: []string{"fido"},
	})
	doer.AssertExpectations(t)
}

func TestClient_WhenPathHasId_RequestHasCorrectPath(t *testing.T) {
	t.Parallel()
	doer := &HTTPRequestDoerMock{}
	client, _ := NewClientWithResponses(hostname, WithHTTPClient(doer))
	_ = client

	doer.On("Do", mock.Anything).Return(nil, errors.New("something went wrong")).Run(func(args mock.Arguments) {
		req, ok := args.Get(0).(*http.Request)
		assert.True(t, ok)
		assert.Equal(t, "/host/pets/id", req.URL.Path)
	})
	petID := "id"
	client.GetPetWithResponse(context.Background(), petID)
	doer.AssertExpectations(t)
}

func TestClient_WhenPathHasIdContainingReservedCharacter_RequestHasCorrectPath(t *testing.T) {
	t.Parallel()
	doer := &HTTPRequestDoerMock{}
	client, _ := NewClientWithResponses(hostname, WithHTTPClient(doer))
	_ = client

	doer.On("Do", mock.Anything).Return(nil, errors.New("something went wrong")).Run(func(args mock.Arguments) {
		req, ok := args.Get(0).(*http.Request)
		assert.True(t, ok)
		assert.Equal(t, "/host/pets/id1%2Fid2", req.URL.Path)
	})
	petID := "id1/id2"
	client.GetPetWithResponse(context.Background(), petID)
	doer.AssertExpectations(t)
}

// HTTPRequestDoerMock mocks the interface HttpRequestDoerMock.
type HTTPRequestDoerMock struct {
	mock.Mock
}

func (m *HTTPRequestDoerMock) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}
