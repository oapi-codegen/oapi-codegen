package encoding

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

// HTTPRequestDoerMock mocks the HttpRequestDoer interface.
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

// mockServer implements the ServerInterface for testing.
type mockServer struct {
	getPet       func(ctx echo.Context, petId string) error
	validatePets func(ctx echo.Context) error
}

func (m *mockServer) GetPet(ctx echo.Context, petId string) error {
	if m.getPet != nil {
		return m.getPet(ctx, petId)
	}
	return ctx.NoContent(http.StatusNotImplemented)
}

func (m *mockServer) ValidatePets(ctx echo.Context) error {
	if m.validatePets != nil {
		return m.validatePets(ctx)
	}
	return ctx.NoContent(http.StatusNotImplemented)
}

func TestClient_PathWithColon(t *testing.T) {
	doer := &HTTPRequestDoerMock{}
	client, _ := NewClientWithResponses("http://host", WithHTTPClient(doer))

	doer.On("Do", mock.Anything).Return(nil, errors.New("expected")).Run(func(args mock.Arguments) {
		req := args.Get(0).(*http.Request)
		assert.Equal(t, "http://host/pets:validate", req.URL.String())
	})

	_, _ = client.ValidatePetsWithResponse(context.Background(), ValidatePetsJSONRequestBody{
		Names: []string{"fido"},
	})
	doer.AssertExpectations(t)
}

func TestClient_PathParamWithID(t *testing.T) {
	doer := &HTTPRequestDoerMock{}
	client, _ := NewClientWithResponses("http://host", WithHTTPClient(doer))

	doer.On("Do", mock.Anything).Return(nil, errors.New("expected")).Run(func(args mock.Arguments) {
		req := args.Get(0).(*http.Request)
		assert.Equal(t, "/pets/id", req.URL.Path)
	})

	_, _ = client.GetPetWithResponse(context.Background(), "id")
	doer.AssertExpectations(t)
}

func TestClient_PathParamWithReservedChar(t *testing.T) {
	doer := &HTTPRequestDoerMock{}
	client, _ := NewClientWithResponses("http://host", WithHTTPClient(doer))

	doer.On("Do", mock.Anything).Return(nil, errors.New("expected")).Run(func(args mock.Arguments) {
		req := args.Get(0).(*http.Request)
		assert.Equal(t, "http://host/pets/id1%2Fid2", req.URL.String())
	})

	_, _ = client.GetPetWithResponse(context.Background(), "id1/id2")
	doer.AssertExpectations(t)
}

func TestServer_UnescapesPathParam(t *testing.T) {
	e := echo.New()
	m := &mockServer{}
	RegisterHandlers(e, m)

	receivedPetID := ""
	m.getPet = func(ctx echo.Context, petId string) error {
		receivedPetID = petId
		return ctx.NoContent(http.StatusOK)
	}

	// Use the generated client to build the request, then serve via recorder.
	req, err := NewGetPetRequest("http://example.com", "id1/id2")
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "id1/id2", receivedPetID)
}
