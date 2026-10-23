package api

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oapi-codegen/testutil"
)

// newStrictApp builds the strict server, optionally behind a middleware that
// stores validated under validatedBodyLocalsKey the way fiber-middleware does
// after a successful validation.
func newStrictApp(validated any) http.Handler {
	app := fiber.New()
	if validated != nil {
		app.Use(func(c *fiber.Ctx) error {
			c.Locals(validatedBodyLocalsKey, validated)
			return c.Next()
		})
	}
	RegisterHandlers(app, NewStrictHandler(StrictServer{}, nil))
	return adaptor.FiberApp(app)
}

// TestFiberServer_BindsValidatedBody checks that the strict handler binds the
// value the validation middleware left in the Fiber context, with exact member
// matching, and otherwise decodes the body as before: Fiber's body parser would
// match "VALUE" to the `value` field case-insensitively and bind the
// unvalidated member instead.
func TestFiberServer_BindsValidatedBody(t *testing.T) {
	const body = `{"value":"123","VALUE":"456"}`
	post := func(t *testing.T, handler http.Handler) *testutil.CompletedRequest {
		t.Helper()
		return testutil.NewRequest().Post("/json").WithContentType("application/json").WithBody([]byte(body)).GoWithHTTPHandler(t, handler)
	}

	t.Run("with a validated value in the context", func(t *testing.T) {
		// What the validator saw: it keeps the last of duplicate members only
		// when they are spelled identically; here both are distinct keys.
		rr := post(t, newStrictApp(map[string]any{"value": "123", "VALUE": "456"})).Recorder
		require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
		assert.JSONEq(t, `{"value":"123"}`, rr.Body.String())
	})

	t.Run("without a validated value the body is decoded as before", func(t *testing.T) {
		rr := post(t, newStrictApp(nil)).Recorder
		require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
		// Fiber's case-insensitive matching lets the later "VALUE" member win;
		// this is the pre-existing behaviour the fallback keeps.
		assert.JSONEq(t, `{"value":"456"}`, rr.Body.String())
	})

	t.Run("a value that does not fit falls back to decoding the body", func(t *testing.T) {
		rr := post(t, newStrictApp(map[string]any{"value": 123})).Recorder
		require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
		assert.JSONEq(t, `{"value":"456"}`, rr.Body.String())
	})
}
