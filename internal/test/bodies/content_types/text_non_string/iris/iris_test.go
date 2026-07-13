package textnonstringiris

import (
	"context"
	"testing"

	"github.com/kataras/iris/v12"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oapi-codegen/testutil"
)

type server struct{}

func (server) GetPing(_ context.Context, _ GetPingRequestObject) (GetPingResponseObject, error) {
	return GetPing201TextResponse(201), nil
}

// TestIrisIntegerTextResponse is the iris-path (ctx.WriteString) regression
// test for issue-1897: an integer text/plain response must serialize as the
// decimal "201", not the single Unicode code-point that string(response)
// previously produced.
func TestIrisIntegerTextResponse(t *testing.T) {
	app := iris.New()
	RegisterHandlers(app, NewStrictHandler(server{}, nil))

	rr := testutil.NewRequest().Get("/ping").GoWithHTTPHandler(t, app).Recorder
	require.Equal(t, 201, rr.Code)
	assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
	assert.Equal(t, "201", rr.Body.String())
}
