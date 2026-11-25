package serversmiddlewareiris

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kataras/iris/v12"
	"github.com/stretchr/testify/assert"
)

type impl struct {
	// securedScopes records what the Secured handler found in the Iris
	// context's value store under the generated scopes key.
	securedScopes any
}

// (GET /test)
func (i *impl) Test(ctx iris.Context) {
	ctx.StatusCode(http.StatusOK)
}

// (GET /secured)
func (i *impl) Secured(ctx iris.Context) {
	i.securedScopes = ctx.Values().Get(string(BearerAuthScopes))
	ctx.StatusCode(http.StatusOK)
}

// From issue-1988: with enable-auth-scopes-on-context, the generated wrapper
// called ctx.Set, which does not exist on iris.Context, so the generated code
// did not compile. The scopes are stored in ctx.Values() instead, where the
// handler (and any handler run after the wrapper) can read them back.
func TestIssue1988ScopesAreStoredInContextValues(t *testing.T) {
	server := &impl{}

	app := iris.New()
	RegisterHandlers(app, server)

	req := httptest.NewRequest(http.MethodGet, "/secured", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, []string{"pet:read"}, server.securedScopes)
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
