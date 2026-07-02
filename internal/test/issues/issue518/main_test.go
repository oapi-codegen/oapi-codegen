package issue518

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type impl struct{}

// (GET /auth-check)
func (i *impl) AuthCheck(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

// (GET /test)
func (i *impl) Test(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

func TestIssue518(t *testing.T) {
	server := &impl{}

	assert.NotPanics(t, func() {
		r := fiber.New()
		RegisterHandlers(r, server)
	})

	assert.NotPanics(t, func() {
		r := fiber.New()
		RegisterHandlersWithOptions(r, server, FiberServerOptions{
			Middlewares: []MiddlewareFunc{
				func(c *fiber.Ctx) error {
					return nil
				},
			},
			HandlerMiddlewares: []HandlerMiddlewareFunc{
				func(c *fiber.Ctx, next fiber.Handler) error {
					return next(c)
				},
			},
		})
	})

	t.Run("endpoint with anonymous security alternative allows missing auth", func(t *testing.T) {
		r := fiber.New()
		RegisterHandlersWithOptions(r, server, FiberServerOptions{
			HandlerMiddlewares: []HandlerMiddlewareFunc{
				func(c *fiber.Ctx, next fiber.Handler) error {
					return next(c)
				},
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/auth-check", nil)
		resp, err := r.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		req = httptest.NewRequest(http.MethodGet, "/auth-check", nil)
		req.Header.Set(fiber.HeaderAuthorization, "Bearer token")
		resp, err = r.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		req = httptest.NewRequest(http.MethodGet, "/test", nil)
		resp, err = r.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

func TestIssue518AnonymousSecurityAlternativeDoesNotEmitScopes(t *testing.T) {
	_, testFile, _, ok := runtime.Caller(0)
	assert.True(t, ok)

	generatedPath := filepath.Join(filepath.Dir(testFile), "main.gen.go")
	generated, err := os.ReadFile(generatedPath)
	require.NoError(t, err)
	assert.False(t, strings.Contains(string(generated), "SetUserValue"))
}
