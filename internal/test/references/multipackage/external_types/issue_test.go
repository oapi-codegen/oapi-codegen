package externaltypes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oapi-codegen/oapi-codegen/v2/internal/test/references/multipackage/external_types/gen/api"
	"github.com/oapi-codegen/oapi-codegen/v2/internal/test/references/multipackage/external_types/gen/common"
)

// Compiling this block is the regression check: each body type is spelled the
// way the imported package declares it, with only its model names qualified.
var (
	_ = api.GetUntypeddefaultJSONResponse{Body: map[string]any{"any": "value"}, StatusCode: http.StatusTeapot}
	_ = api.GetPointerdefaultJSONResponse{Body: new(string), StatusCode: http.StatusAccepted}
	_ = api.ListThingsdefaultJSONResponse{Body: []common.Thing{{Id: "a"}}, StatusCode: http.StatusOK}
	_ = api.UpdateThingdefaultJSONResponse{Body: "untyped", StatusCode: http.StatusBadRequest}
	_ = api.UpdateThing200ResponseHeaders{XThingID: common.ThingID("a")}
	_ = api.UpdateThingParams{Limit: new(int)}
)

type strictServer struct{}

func (strictServer) UpdateThing(_ context.Context, request api.UpdateThingRequestObject) (api.UpdateThingResponseObject, error) {
	if request.Params.Limit != nil && *request.Params.Limit == 0 {
		return api.UpdateThingdefaultJSONResponse{Body: map[string]any{"error": "limit"}, StatusCode: http.StatusBadRequest}, nil
	}
	return api.UpdateThing200JSONResponse{
		Body:    *request.Body,
		Headers: api.UpdateThing200ResponseHeaders{XThingID: request.Id},
	}, nil
}

func (strictServer) GetUntyped(context.Context, api.GetUntypedRequestObject) (api.GetUntypedResponseObject, error) {
	return api.GetUntypeddefaultJSONResponse{Body: []any{"untyped"}, StatusCode: http.StatusTeapot}, nil
}

func (strictServer) GetPointer(context.Context, api.GetPointerRequestObject) (api.GetPointerResponseObject, error) {
	body := "pointer"
	return api.GetPointerdefaultJSONResponse{Body: &body, StatusCode: http.StatusAccepted}, nil
}

func (strictServer) ListThings(context.Context, api.ListThingsRequestObject) (api.ListThingsResponseObject, error) {
	return api.ListThingsdefaultJSONResponse{Body: []common.Thing{{Id: "a"}, {Id: "b"}}, StatusCode: http.StatusOK}, nil
}

// TestExternalTypesRoundTrip sends each response through the generated strict
// server and decodes it with the generated client.
func TestExternalTypesRoundTrip(t *testing.T) {
	srv := httptest.NewServer(api.Handler(api.NewStrictHandler(strictServer{}, nil)))
	defer srv.Close()

	client, err := api.NewClientWithResponses(srv.URL)
	require.NoError(t, err)
	ctx := context.Background()

	t.Run("reused path item", func(t *testing.T) {
		resp, err := client.UpdateThingWithResponse(ctx, "t1", &api.UpdateThingParams{}, api.UpdateThingJSONRequestBody{{Id: "t1"}})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		assert.Equal(t, "t1", resp.HTTPResponse.Header.Get("X-Thing-ID"))
		require.NotNil(t, resp.JSON200)
		assert.Equal(t, []common.Thing{{Id: "t1"}}, *resp.JSON200)
	})

	t.Run("reused path item default", func(t *testing.T) {
		limit := 0
		resp, err := client.UpdateThingWithResponse(ctx, "t1", &api.UpdateThingParams{Limit: &limit}, api.UpdateThingJSONRequestBody{})
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode())
		require.NotNil(t, resp.JSONDefault)
		assert.Equal(t, map[string]any{"error": "limit"}, *resp.JSONDefault)
	})

	t.Run("untyped", func(t *testing.T) {
		resp, err := client.GetUntypedWithResponse(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusTeapot, resp.StatusCode())
		require.NotNil(t, resp.JSONDefault)
		assert.Equal(t, []any{"untyped"}, *resp.JSONDefault)
	})

	t.Run("pointer", func(t *testing.T) {
		resp, err := client.GetPointerWithResponse(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, resp.StatusCode())
		require.NotNil(t, resp.JSONDefault)
		require.NotNil(t, *resp.JSONDefault)
		assert.Equal(t, "pointer", **resp.JSONDefault)
	})

	t.Run("things", func(t *testing.T) {
		resp, err := client.ListThingsWithResponse(ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.NotNil(t, resp.JSONDefault)
		assert.Equal(t, common.Things{{Id: "a"}, {Id: "b"}}, *resp.JSONDefault)
	})
}
