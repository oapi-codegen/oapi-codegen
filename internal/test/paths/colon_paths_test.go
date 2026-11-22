package paths

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	fiberv2 "github.com/gofiber/fiber/v2"
	fiberv3 "github.com/gofiber/fiber/v3"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	colonecho "github.com/oapi-codegen/oapi-codegen/v2/internal/test/paths/echo"
	colonfiber "github.com/oapi-codegen/oapi-codegen/v2/internal/test/paths/fiber"
	colonfiberv3 "github.com/oapi-codegen/oapi-codegen/v2/internal/test/paths/fiberv3"
	colongin "github.com/oapi-codegen/oapi-codegen/v2/internal/test/paths/gin"
)

// Each handler writes its own operation name so the test can prove that
// /pets:validate and /pets:generate dispatch to distinct handlers rather than
// colliding on a single ":"-parameter route (issue #1726).

type echoServer struct{}

func (echoServer) ValidatePets(ctx echo.Context) error { return ctx.String(http.StatusOK, "validate") }
func (echoServer) GeneratePets(ctx echo.Context) error { return ctx.String(http.StatusOK, "generate") }

type ginServer struct{}

func (ginServer) ValidatePets(c *gin.Context) { c.String(http.StatusOK, "validate") }
func (ginServer) GeneratePets(c *gin.Context) { c.String(http.StatusOK, "generate") }

type fiberServer struct{}

func (fiberServer) ValidatePets(c *fiberv2.Ctx) error { return c.SendString("validate") }
func (fiberServer) GeneratePets(c *fiberv2.Ctx) error { return c.SendString("generate") }

type fiberv3Server struct{}

func (fiberv3Server) ValidatePets(c fiberv3.Ctx) error { return c.SendString("validate") }
func (fiberv3Server) GeneratePets(c fiberv3.Ctx) error { return c.SendString("generate") }

// wantBody maps each colon path to the body its dedicated handler returns.
var wantBody = map[string]string{
	"/pets:validate": "validate",
	"/pets:generate": "generate",
}

// assertServeHTTP drives an http.Handler-compatible router (echo, gin).
func assertServeHTTP(t *testing.T, name string, h http.Handler) {
	t.Helper()
	for path, want := range wantBody {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, path, nil)
		h.ServeHTTP(rec, req)
		assert.Equalf(t, http.StatusOK, rec.Code, "%s POST %s status", name, path)
		assert.Equalf(t, want, rec.Body.String(), "%s POST %s dispatched to wrong handler", name, path)
	}
}

func TestEchoColonPathsRouteIndependently(t *testing.T) {
	e := echo.New()
	colonecho.RegisterHandlers(e, echoServer{})
	assertServeHTTP(t, "echo", e)
}

func TestGinColonPathsRouteIndependently(t *testing.T) {
	gin.SetMode(gin.TestMode)
	g := gin.New()
	colongin.RegisterHandlers(g, ginServer{})
	assertServeHTTP(t, "gin", g)
}

func TestFiberColonPathsRouteIndependently(t *testing.T) {
	app := fiberv2.New()
	colonfiber.RegisterHandlers(app, fiberServer{})
	for path, want := range wantBody {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		resp, err := app.Test(req)
		require.NoErrorf(t, err, "fiber POST %s", path)
		assert.Equalf(t, http.StatusOK, resp.StatusCode, "fiber POST %s status", path)
		body, _ := io.ReadAll(resp.Body)
		assert.Equalf(t, want, string(body), "fiber POST %s dispatched to wrong handler", path)
	}
}

func TestFiberV3ColonPathsRouteIndependently(t *testing.T) {
	app := fiberv3.New()
	colonfiberv3.RegisterHandlers(app, fiberv3Server{})
	for path, want := range wantBody {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		resp, err := app.Test(req)
		require.NoErrorf(t, err, "fiberv3 POST %s", path)
		assert.Equalf(t, http.StatusOK, resp.StatusCode, "fiberv3 POST %s status", path)
		body, _ := io.ReadAll(resp.Body)
		assert.Equalf(t, want, string(body), "fiberv3 POST %s dispatched to wrong handler", path)
	}
}
