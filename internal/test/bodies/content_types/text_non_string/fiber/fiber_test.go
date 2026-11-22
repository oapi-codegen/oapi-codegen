package textnonstringfiber

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oapi-codegen/testutil"
)

type server struct{}

func (server) GetPing(_ context.Context, _ GetPingRequestObject) (GetPingResponseObject, error) {
	return GetPing201TextResponse(201), nil
}

// TestFiberIntegerTextResponse is the fiber-path (ctx.WriteString) regression
// test for issue-1897: an integer text/plain response must serialize as the
// decimal "201", not the single Unicode code-point that string(response)
// previously produced.
func TestFiberIntegerTextResponse(t *testing.T) {
	app := fiber.New()
	RegisterHandlers(app, NewStrictHandler(server{}, nil))

	rr := testutil.NewRequest().Get("/ping").GoWithHTTPHandler(t, adaptor.FiberApp(app)).Recorder
	require.Equal(t, 201, rr.Code)
	assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
	assert.Equal(t, "201", rr.Body.String())
}
