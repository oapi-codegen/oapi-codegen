package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/kataras/iris/v12"
	"github.com/stretchr/testify/assert"

	clientAPI "github.com/oapi-codegen/oapi-codegen/v2/internal/test/servers/strict/client"
	"github.com/oapi-codegen/testutil"
)

type erroringServer struct{ StrictServer }

func (erroringServer) JSONExample(ctx context.Context, request JSONExampleRequestObject) (JSONExampleResponseObject, error) {
	return nil, errors.New("handler failure")
}

// Errors returned by strict handlers used to be reported via
// ctx.StopWithError(StatusBadRequest, ...); they are server-side
// failures and must be reported as 500.
func TestIrisHandlerErrorIsServerError(t *testing.T) {
	app := iris.New()
	RegisterHandlers(app, NewStrictHandler(erroringServer{}, nil))

	value := "123"
	requestBody := clientAPI.Example{Value: &value}
	rr := testutil.NewRequest().Post("/json").WithJsonBody(requestBody).GoWithHTTPHandler(t, app).Recorder
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
