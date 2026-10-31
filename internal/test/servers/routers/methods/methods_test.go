package serversroutersmethods

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	fiberv2 "github.com/gofiber/fiber/v2"
	fiberv3 "github.com/gofiber/fiber/v3"
	"github.com/kataras/iris/v12"
	echov4 "github.com/labstack/echo/v4"
	echov5 "github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	chiapi "github.com/oapi-codegen/oapi-codegen/v2/internal/test/servers/routers/methods/chi"
	clientapi "github.com/oapi-codegen/oapi-codegen/v2/internal/test/servers/routers/methods/client"
	echoapi "github.com/oapi-codegen/oapi-codegen/v2/internal/test/servers/routers/methods/echo"
	echo5api "github.com/oapi-codegen/oapi-codegen/v2/internal/test/servers/routers/methods/echo5"
	fiberapi "github.com/oapi-codegen/oapi-codegen/v2/internal/test/servers/routers/methods/fiber"
	fiberv3api "github.com/oapi-codegen/oapi-codegen/v2/internal/test/servers/routers/methods/fiber-v3"
	ginapi "github.com/oapi-codegen/oapi-codegen/v2/internal/test/servers/routers/methods/gin"
	gorillaapi "github.com/oapi-codegen/oapi-codegen/v2/internal/test/servers/routers/methods/gorilla"
	irisapi "github.com/oapi-codegen/oapi-codegen/v2/internal/test/servers/routers/methods/iris"
	stdhttpapi "github.com/oapi-codegen/oapi-codegen/v2/internal/test/servers/routers/methods/std-http"
)

// calls records what the fake server saw, so the test can check that each
// router dispatched the custom-method request to the right handler with the
// path and query parameters bound.
type calls struct {
	listed   bool
	queried  string // request body seen by QUERY /items
	purgedID string
	force    *bool
}

const queryBody = `{"name":"widget"}`

// drive sends the three requests through the router under test and checks
// the recorded calls. handler is the router as an http.Handler.
func drive(t *testing.T, handler http.Handler, c *calls) {
	t.Helper()

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/items", nil))
	assert.Equal(t, http.StatusOK, rec.Code, "GET /items")
	assert.True(t, c.listed)

	rec = httptest.NewRecorder()
	req := httptest.NewRequest("QUERY", "/items", bytes.NewBufferString(queryBody))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code, "QUERY /items")
	assert.Equal(t, queryBody, c.queried)

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("PURGE", "/items/42?force=true", nil))
	assert.Equal(t, http.StatusNoContent, rec.Code, "PURGE /items/42")
	assert.Equal(t, "42", c.purgedID)
	if assert.NotNil(t, c.force) {
		assert.True(t, *c.force)
	}
}

// --- net/http family: chi, gorilla, std-http share the handler signatures.

type netHTTPServer struct{ c *calls }

func (s *netHTTPServer) ListItems(w http.ResponseWriter, r *http.Request) {
	s.c.listed = true
	w.WriteHeader(http.StatusOK)
}

func (s *netHTTPServer) QueryItems(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	s.c.queried = string(b)
	w.WriteHeader(http.StatusOK)
}

func (s *netHTTPServer) purge(w http.ResponseWriter, id string, force *bool) {
	s.c.purgedID = id
	s.c.force = force
	w.WriteHeader(http.StatusNoContent)
}

type chiServer struct{ netHTTPServer }

func (s *chiServer) PurgeItem(w http.ResponseWriter, r *http.Request, id string, params chiapi.PurgeItemParams) {
	s.purge(w, id, params.Force)
}

type gorillaServer struct{ netHTTPServer }

func (s *gorillaServer) PurgeItem(w http.ResponseWriter, r *http.Request, id string, params gorillaapi.PurgeItemParams) {
	s.purge(w, id, params.Force)
}

type stdHTTPServer struct{ netHTTPServer }

func (s *stdHTTPServer) PurgeItem(w http.ResponseWriter, r *http.Request, id string, params stdhttpapi.PurgeItemParams) {
	s.purge(w, id, params.Force)
}

// chi only routes methods it knows about, and the generated code deliberately
// does not register them (chi.RegisterMethod is process-wide state, and one
// process commonly hosts several generated packages). The application does it
// once at startup, before any routes are registered.
func TestChi(t *testing.T) {
	assert.Panics(t, func() {
		chiapi.Handler(&chiServer{netHTTPServer{&calls{}}})
	}, "chi rejects custom methods until chi.RegisterMethod has been called")

	chi.RegisterMethod("QUERY")
	chi.RegisterMethod("PURGE")

	c := &calls{}
	drive(t, chiapi.Handler(&chiServer{netHTTPServer{c}}), c)
}

func TestGorilla(t *testing.T) {
	c := &calls{}
	drive(t, gorillaapi.Handler(&gorillaServer{netHTTPServer{c}}), c)
}

func TestStdHTTP(t *testing.T) {
	c := &calls{}
	drive(t, stdhttpapi.Handler(&stdHTTPServer{netHTTPServer{c}}), c)
}

// --- echo v4 / v5

type echoServer struct{ c *calls }

func (s *echoServer) ListItems(ctx echov4.Context) error {
	s.c.listed = true
	return ctx.NoContent(http.StatusOK)
}

func (s *echoServer) QueryItems(ctx echov4.Context) error {
	b, _ := io.ReadAll(ctx.Request().Body)
	s.c.queried = string(b)
	return ctx.NoContent(http.StatusOK)
}

func (s *echoServer) PurgeItem(ctx echov4.Context, id string, params echoapi.PurgeItemParams) error {
	s.c.purgedID = id
	s.c.force = params.Force
	return ctx.NoContent(http.StatusNoContent)
}

func TestEcho(t *testing.T) {
	c := &calls{}
	e := echov4.New()
	echoapi.RegisterHandlers(e, &echoServer{c})
	drive(t, e, c)
}

type echo5Server struct{ c *calls }

func (s *echo5Server) ListItems(ctx *echov5.Context) error {
	s.c.listed = true
	return ctx.NoContent(http.StatusOK)
}

func (s *echo5Server) QueryItems(ctx *echov5.Context) error {
	b, _ := io.ReadAll(ctx.Request().Body)
	s.c.queried = string(b)
	return ctx.NoContent(http.StatusOK)
}

func (s *echo5Server) PurgeItem(ctx *echov5.Context, id string, params echo5api.PurgeItemParams) error {
	s.c.purgedID = id
	s.c.force = params.Force
	return ctx.NoContent(http.StatusNoContent)
}

func TestEcho5(t *testing.T) {
	c := &calls{}
	e := echov5.New()
	echo5api.RegisterHandlers(e, &echo5Server{c})
	drive(t, e, c)
}

// --- gin

type ginServer struct{ c *calls }

func (s *ginServer) ListItems(ctx *gin.Context) {
	s.c.listed = true
	ctx.Status(http.StatusOK)
}

func (s *ginServer) QueryItems(ctx *gin.Context) {
	b, _ := io.ReadAll(ctx.Request.Body)
	s.c.queried = string(b)
	ctx.Status(http.StatusOK)
}

func (s *ginServer) PurgeItem(ctx *gin.Context, id string, params ginapi.PurgeItemParams) {
	s.c.purgedID = id
	s.c.force = params.Force
	ctx.Status(http.StatusNoContent)
}

func TestGin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c := &calls{}
	g := gin.New()
	ginapi.RegisterHandlers(g, &ginServer{c})
	drive(t, g, c)
}

// --- iris

type irisServer struct{ c *calls }

func (s *irisServer) ListItems(ctx iris.Context) {
	s.c.listed = true
	ctx.StatusCode(http.StatusOK)
}

func (s *irisServer) QueryItems(ctx iris.Context) {
	b, _ := io.ReadAll(ctx.Request().Body)
	s.c.queried = string(b)
	ctx.StatusCode(http.StatusOK)
}

func (s *irisServer) PurgeItem(ctx iris.Context, id string, params irisapi.PurgeItemParams) {
	s.c.purgedID = id
	s.c.force = params.Force
	ctx.StatusCode(http.StatusNoContent)
}

func TestIris(t *testing.T) {
	c := &calls{}
	app := iris.New()
	irisapi.RegisterHandlers(app, &irisServer{c})
	drive(t, app, c)
}

// --- fiber v2 / v3
//
// fiber validates methods against its app config, and the generated code
// registers routes on a router it is handed, so the application has to be
// created with the custom methods in fiber.Config.RequestMethods. Without
// that, route registration panics.

type fiberServer struct{ c *calls }

func (s *fiberServer) ListItems(ctx *fiberv2.Ctx) error {
	s.c.listed = true
	return ctx.SendStatus(http.StatusOK)
}

func (s *fiberServer) QueryItems(ctx *fiberv2.Ctx) error {
	s.c.queried = string(ctx.Body())
	return ctx.SendStatus(http.StatusOK)
}

func (s *fiberServer) PurgeItem(ctx *fiberv2.Ctx, id string, params fiberapi.PurgeItemParams) error {
	s.c.purgedID = id
	s.c.force = params.Force
	return ctx.SendStatus(http.StatusNoContent)
}

// fiberHandler adapts fiber's in-process App.Test to http.Handler so drive
// can treat it like the other routers.
type fiberHandler func(*http.Request) (*http.Response, error)

func (f fiberHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp, err := f(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func TestFiber(t *testing.T) {
	c := &calls{}
	app := fiberv2.New(fiberv2.Config{
		RequestMethods: append(slices.Clone(fiberv2.DefaultMethods), "QUERY", "PURGE"),
	})
	fiberapi.RegisterHandlers(app, &fiberServer{c})
	drive(t, fiberHandler(func(r *http.Request) (*http.Response, error) { return app.Test(r) }), c)

	assert.Panics(t, func() {
		fiberapi.RegisterHandlers(fiberv2.New(), &fiberServer{&calls{}})
	}, "fiber rejects custom methods unless they are listed in Config.RequestMethods")
}

type fiberV3Server struct{ c *calls }

func (s *fiberV3Server) ListItems(ctx fiberv3.Ctx) error {
	s.c.listed = true
	return ctx.SendStatus(http.StatusOK)
}

func (s *fiberV3Server) QueryItems(ctx fiberv3.Ctx) error {
	s.c.queried = string(ctx.Body())
	return ctx.SendStatus(http.StatusOK)
}

func (s *fiberV3Server) PurgeItem(ctx fiberv3.Ctx, id string, params fiberv3api.PurgeItemParams) error {
	s.c.purgedID = id
	s.c.force = params.Force
	return ctx.SendStatus(http.StatusNoContent)
}

func TestFiberV3(t *testing.T) {
	c := &calls{}
	app := fiberv3.New(fiberv3.Config{
		RequestMethods: append(slices.Clone(fiberv3.DefaultMethods), "QUERY", "PURGE"),
	})
	fiberv3api.RegisterHandlers(app, &fiberV3Server{c})
	drive(t, fiberHandler(func(r *http.Request) (*http.Response, error) { return app.Test(r) }), c)

	assert.Panics(t, func() {
		fiberv3api.RegisterHandlers(fiberv3.New(), &fiberV3Server{&calls{}})
	}, "fiber rejects custom methods unless they are listed in Config.RequestMethods")
}

// --- client

func TestClientRequests(t *testing.T) {
	req, err := clientapi.NewQueryItemsRequest("http://example.com", clientapi.QueryItemsJSONRequestBody{})
	require.NoError(t, err)
	assert.Equal(t, "QUERY", req.Method)

	force := true
	req, err = clientapi.NewPurgeItemRequest("http://example.com", "42", &clientapi.PurgeItemParams{Force: &force})
	require.NoError(t, err)
	assert.Equal(t, "PURGE", req.Method)
	assert.Equal(t, "/items/42", req.URL.Path)
	assert.Equal(t, "force=true", req.URL.RawQuery)
}
