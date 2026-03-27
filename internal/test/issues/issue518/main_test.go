package issue518

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
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

// hasSecurityScopes returns true if the BearerAuthScopes key was set in context,
// even if the scopes slice is empty (an empty slice means the security scheme is
// defined on the operation with no required scopes, which still requires auth).
func hasSecurityScopes(c *fiber.Ctx) bool {
	_, ok := c.Context().UserValue(BearerAuthScopes).([]string)
	return ok
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
					if hasSecurityScopes(c) && c.Get(fiber.HeaderAuthorization) == "" {
						return c.SendStatus(fiber.StatusUnauthorized)
					}
					return next(c)
				},
			},
		})
	})

	t.Run("secured endpoint requires auth when scopes are present", func(t *testing.T) {
		r := fiber.New()
		RegisterHandlersWithOptions(r, server, FiberServerOptions{
			HandlerMiddlewares: []HandlerMiddlewareFunc{
				func(c *fiber.Ctx, next fiber.Handler) error {
					if hasSecurityScopes(c) && c.Get(fiber.HeaderAuthorization) == "" {
						return c.SendStatus(fiber.StatusUnauthorized)
					}
					return next(c)
				},
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/auth-check", nil)
		resp, err := r.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

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
