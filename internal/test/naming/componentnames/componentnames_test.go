package namingcomponentnames

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// server implements the renamed strict server interface. That this compiles
// at all is the point: the interface is PetStoreStrictServer, not
// StrictServerInterface.
type server struct{}

func (server) CreateThing(_ context.Context, request CreateThingRequestObject) (CreateThingResponseObject, error) {
	return CreateThing201JSONResponse(*request.Body), nil
}

func (server) GetThing(_ context.Context, request GetThingRequestObject) (GetThingResponseObject, error) {
	return GetThing200JSONResponse{Id: request.ThingId, Name: &request.Params.Kind}, nil
}

// compile-time assertions that the renamed types have the shapes their
// original names had.
var (
	_ PetStoreStrictServer    = server{}
	_ PetStoreRouter          = http.NewServeMux()
	_ AdminServerInterface    = AdminNotImplementedYet{}
	_ PetStoreAPIInterface    = (*PetStoreAPI)(nil)
	_ PetStoreHttpRequestDoer = (*http.Client)(nil)
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	handler := PetStoreMuxWithOptions(
		PetStoreNewStrictHandler(server{}, nil),
		PetStoreStdHTTPServerOptions{BaseRouter: http.NewServeMux()},
	)
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

// TestRenamedClientAndServerRoundTrip drives the renamed client against the
// renamed strict server, exercising the whole rename surface end to end.
func TestRenamedClientAndServerRoundTrip(t *testing.T) {
	srv := newTestServer(t)

	client, err := NewPetStoreAPIWithResponses(srv.URL, PetStoreWithHTTPClient(srv.Client()))
	require.NoError(t, err)

	created, err := client.CreateThingWithResponse(context.Background(), CreateThingJSONRequestBody{
		Id: "abc",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, created.StatusCode())
	require.NotNil(t, created.JSON201)
	assert.Equal(t, "abc", created.JSON201.Id)

	got, err := client.GetThingWithResponse(context.Background(), "abc", &GetThingParams{
		Kind:       "widget",
		XRequestId: "req-1",
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, got.StatusCode())
	require.NotNil(t, got.JSON200)
	require.NotNil(t, got.JSON200.Name)
	assert.Equal(t, "widget", *got.JSON200.Name)
}

// TestRenamedErrorType checks that the renamed parameter-binding error type
// is the one the wrapper actually hands to the error handler: `kind` is a
// required query parameter, and its absence must surface as
// PetStoreMissingQueryParam (the rename of RequiredParamError).
func TestRenamedErrorType(t *testing.T) {
	var handed error
	handler := PetStoreMuxWithOptions(
		PetStoreNewStrictHandler(server{}, nil),
		PetStoreStdHTTPServerOptions{
			BaseRouter: http.NewServeMux(),
			ErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
				handed = err
				w.WriteHeader(http.StatusBadRequest)
			},
		},
	)
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/things/abc", nil)
	require.NoError(t, err)
	req.Header.Set("X-Request-Id", "req-1")
	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	require.IsType(t, &PetStoreMissingQueryParam{}, handed)
	assert.Equal(t, "kind", handed.(*PetStoreMissingQueryParam).ParamName)
}

// TestTwoEmbeddedSpecsInOnePackage is the multi-spec-one-package case that
// the prefix exists for: both generation runs embed the spec, each with its
// own set of unexported names, and both accessors work.
func TestTwoEmbeddedSpecsInOnePackage(t *testing.T) {
	petStore, err := PetStoreLoadSpec()
	require.NoError(t, err)
	require.NotNil(t, petStore)
	assert.Equal(t, "component-names", petStore.Info.Title)

	// The deprecated accessor was renamed independently of GetSpec.
	viaSwagger, err := PetStoreLoadSwagger()
	require.NoError(t, err)
	assert.Equal(t, petStore.Info.Title, viaSwagger.Info.Title)

	admin, err := AdminGetSpec()
	require.NoError(t, err)
	assert.Equal(t, petStore.Info.Title, admin.Info.Title)

	// Distinct unexported caches, which is what makes the two runs coexist.
	petStoreRaw, err := petStoreRawSpec()
	require.NoError(t, err)
	adminRaw, err := adminRawSpec()
	require.NoError(t, err)
	assert.Equal(t, petStoreRaw, adminRaw)
	assert.NotEmpty(t, petStoreSwaggerSpec)
	assert.NotEmpty(t, adminSwaggerSpec)
}
