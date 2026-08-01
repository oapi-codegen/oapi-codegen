package serversmiddlewareiris

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kataras/iris/v12"
	"github.com/stretchr/testify/assert"
)

type impl struct{}

// (GET /test)
func (i *impl) Test(ctx iris.Context) {
	ctx.StatusCode(http.StatusOK)
}

// IrisServerOptions.Middlewares used to be silently ignored by
// RegisterHandlersWithOptions; this asserts they actually run.
func TestMiddlewaresAreApplied(t *testing.T) {
	server := &impl{}

	calls := 0
	app := iris.New()
	RegisterHandlersWithOptions(app, server, IrisServerOptions{
		Middlewares: []MiddlewareFunc{
			func(ctx iris.Context) {
				calls++
				ctx.Next()
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, calls)
}
