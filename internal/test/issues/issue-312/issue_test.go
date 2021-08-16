package issue_312

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const hostname = "http://host"

func TestClient_WhenPathHasColon_RequestHasCorrectPath(t *testing.T) {
	doer := &HTTPRequestDoerMock{}
	client, _ := NewClientWithResponses(hostname, WithHTTPClient(doer))
	_ = client

	doer.On("Do", mock.Anything).Return(nil, errors.New("something went wrong")).Run(func(args mock.Arguments) {
		req, ok := args.Get(0).(*http.Request)
		assert.True(t, ok)
		assert.Equal(t, "http://host/pets:validate", req.URL.String())
	})

	client.ValidatePetsWithResponse(context.Background(), ValidatePetsJSONRequestBody{
		Names: []string{"fido"},
	})
	doer.AssertExpectations(t)
}

func TestClient_WhenPathHasId_RequestHasCorrectPath(t *testing.T) {
	doer := &HTTPRequestDoerMock{}
	client, _ := NewClientWithResponses(hostname, WithHTTPClient(doer))
	_ = client

	doer.On("Do", mock.Anything).Return(nil, errors.New("something went wrong")).Run(func(args mock.Arguments) {
		req, ok := args.Get(0).(*http.Request)
		assert.True(t, ok)
		assert.Equal(t, "/pets/id", req.URL.Path)
	})
	petID := "id"
	client.GetPetWithResponse(context.Background(), petID)
	doer.AssertExpectations(t)
}

func TestClient_WhenPathHasIdContainingReservedCharacter_RequestHasCorrectPath(t *testing.T) {
	doer := &HTTPRequestDoerMock{}
	client, _ := NewClientWithResponses(hostname, WithHTTPClient(doer))
	_ = client

	doer.On("Do", mock.Anything).Return(nil, errors.New("something went wrong")).Run(func(args mock.Arguments) {
		req, ok := args.Get(0).(*http.Request)
		assert.True(t, ok)
		assert.Equal(t, "http://host/pets/id1%2Fid2", req.URL.String())
	})
	petID := "id1/id2"
	client.GetPetWithResponse(context.Background(), petID)
	doer.AssertExpectations(t)
}

func TestClient_ServerUnescapesEscapedArg(t *testing.T) {

	e := echo.New()
	m := &MockClient{}
	RegisterHandlers(e, m)

	svr := httptest.NewServer(e)
	defer svr.Close()

	// We'll make a function in the mock client which records the value of
	// the petId variable
	receivedPetID := ""
	m.getPet = func(ctx echo.Context, petId string) error {
		receivedPetID = petId
		return ctx.NoContent(http.StatusOK)
	}

	client, err := NewClientWithResponses(svr.URL)
	require.NoError(t, err)

	petID := "id1/id2"
	response, err := client.GetPetWithResponse(context.Background(), petID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode())
	assert.Equal(t, petID, receivedPetID)
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

// An implementation of the server interface which helps us check server
// expectations for funky paths and parameters.
type MockClient struct {
	getPet       func(ctx echo.Context, petId string) error
	validatePets func(ctx echo.Context) error
}

func (m *MockClient) GetPet(ctx echo.Context, petId string) error {
	if m.getPet != nil {
		return m.getPet(ctx, petId)
	}
	return ctx.NoContent(http.StatusNotImplemented)
}

func (m *MockClient) ValidatePets(ctx echo.Context) error {
	if m.validatePets != nil {
		return m.validatePets(ctx)
	}
	return ctx.NoContent(http.StatusNotImplemented)
}
