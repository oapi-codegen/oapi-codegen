package fibermid

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{"empty", "", ""},
		{"no version param", "application/json", ""},
		{"version param", "application/json;version=2", "v2"},
		{"spaces around param", "application/json; version=3 ", "v3"},
		{"first of several", "application/json; version=1; charset=utf-8", "v1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ExtractVersion(tt.header))
		})
	}
}

func TestStripVersionPrefix(t *testing.T) {
	tests := []struct {
		in     string
		want   string
		wantOK bool
	}{
		{"/v1/events/:id", "/events/:id", true},
		{"/v1/events", "/events", true},
		{"/latest/events", "/events", true},
		// No second segment to strip down to, so there is no bare path.
		{"/v1", "", false},
		{"events", "", false},
		{"", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, ok := stripVersionPrefix(tt.in)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAppendToHandlers(t *testing.T) {
	h1 := func(*fiber.Ctx) error { return nil }
	h2 := func(*fiber.Ctx) error { return nil }

	apiHandlers := map[string]map[string]fiber.Handler{
		"/v1/events": {"GET": h1},
	}
	AppendToHandlers(apiHandlers, map[string]map[string]fiber.Handler{
		"/v1/events": {"POST": h2},
		"/v2/events": {"GET": h2},
	})

	// The existing per-method entry survives a merge onto the same path.
	assert.Len(t, apiHandlers, 2)
	assert.Contains(t, apiHandlers["/v1/events"], "GET")
	assert.Contains(t, apiHandlers["/v1/events"], "POST")
	assert.Contains(t, apiHandlers["/v2/events"], "GET")
}

// handlerReturning builds a handler that writes body, so a test can tell which
// version's handler actually ran.
func handlerReturning(body string) fiber.Handler {
	return func(c *fiber.Ctx) error { return c.SendString(body) }
}

// echoParam proves that path parameters resolve through the dispatcher: the
// bare-path route does the matching and the same ctx is forwarded on.
func echoParam(prefix string) fiber.Handler {
	return func(c *fiber.Ctx) error { return c.SendString(prefix + ":" + c.Params("id")) }
}

func TestMountVersionedRoutes(t *testing.T) {
	app := fiber.New()
	apiHandlers := map[string]map[string]fiber.Handler{
		"/v1/events":         {"GET": handlerReturning("v1-list")},
		"/v2/events":         {"GET": handlerReturning("v2-list")},
		"/latest/events":     {"GET": handlerReturning("v2-list")},
		"/v1/events/:id":     {"GET": echoParam("v1")},
		"/latest/events/:id": {"GET": echoParam("v1")},
	}
	MountVersionedRoutes(app, apiHandlers, "v1")

	do := func(t *testing.T, method, target, version string) (int, string) {
		t.Helper()
		req := httptest.NewRequest(method, target, nil)
		if version != "" {
			req.Header.Set("API-Version", version)
		}
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close() //nolint:errcheck // test cleanup
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		return resp.StatusCode, string(body)
	}

	t.Run("explicit version selects that version's handler", func(t *testing.T) {
		code, body := do(t, http.MethodGet, "/events", "v2")
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, "v2-list", body)
	})

	t.Run("bare version number is normalized", func(t *testing.T) {
		code, body := do(t, http.MethodGet, "/events", "2")
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, "v2-list", body)
	})

	t.Run("missing header falls back to the default version", func(t *testing.T) {
		code, body := do(t, http.MethodGet, "/events", "")
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, "v1-list", body)
	})

	t.Run("unknown version falls back to latest", func(t *testing.T) {
		code, body := do(t, http.MethodGet, "/events", "v9")
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, "v2-list", body)
	})

	t.Run("path parameters resolve through the dispatcher", func(t *testing.T) {
		code, body := do(t, http.MethodGet, "/events/42", "v1")
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, "v1:42", body)
	})

	t.Run("versioned paths are not registered", func(t *testing.T) {
		code, _ := do(t, http.MethodGet, "/v1/events", "")
		assert.Equal(t, http.StatusNotFound, code)
	})
}

func TestGetErrorCode(t *testing.T) {
	err := io.EOF
	tests := []struct {
		name string
		in   error
		want int
	}{
		{"invalid input", InvalidInput{Err: err}, http.StatusBadRequest},
		{"unauthorized", Unauthorized{Err: err}, http.StatusUnauthorized},
		{"not found", NotFound{Err: err}, http.StatusNotFound},
		{"transient", TransientError{Err: err}, http.StatusServiceUnavailable},
		{"internal", InternalServer{Err: err}, http.StatusInternalServerError},
		{"unknown defaults to 500", err, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetErrorCode(tt.in))
		})
	}
}
