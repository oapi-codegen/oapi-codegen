package parametersroundtrip_test

import (
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	fiberv3 "github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	fiberparams "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/roundtrip/fiber"
	fibergen "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/roundtrip/fiber/gen"
	fiberv3params "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/roundtrip/fiberv3"
	fiberv3gen "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/roundtrip/fiberv3/gen"
)

// Fiber hands out parameter strings that alias buffers it reuses for the next
// request. The generated wrappers copy them, so a handler may keep its
// arguments after it returns. These tests keep the arguments of a first round
// of requests, send a second round with different bytes over the same
// connection, and check the kept values are unchanged.
// See https://github.com/oapi-codegen/oapi-codegen/issues/1932.

// retained collects the strings a handler was given, without copying them.
type retained map[string]*string

func (r retained) keep(name string, v *string) {
	if _, seen := r[name]; !seen && v != nil {
		r[name] = v
	}
}

type fiberRetainingServer struct {
	fiberparams.Server
	got retained
}

func (s *fiberRetainingServer) GetSimpleString(c *fiber.Ctx, param string) error {
	s.got.keep("path styled", &param)
	return c.SendStatus(http.StatusNoContent)
}

func (s *fiberRetainingServer) GetPassThrough(c *fiber.Ctx, param string) error {
	s.got.keep("path passthrough", &param)
	return c.SendStatus(http.StatusNoContent)
}

func (s *fiberRetainingServer) GetQueryForm(c *fiber.Ctx, params fibergen.GetQueryFormParams) error {
	s.got.keep("query styled", params.Ps)
	return c.SendStatus(http.StatusNoContent)
}

func (s *fiberRetainingServer) GetHeader(c *fiber.Ctx, params fibergen.GetHeaderParams) error {
	s.got.keep("header passthrough", params.XPassThrough)
	s.got.keep("header styled", params.N1StartingWithNumber)
	return c.SendStatus(http.StatusNoContent)
}

func (s *fiberRetainingServer) GetCookie(c *fiber.Ctx, params fibergen.GetCookieParams) error {
	s.got.keep("cookie styled", params.N1s)
	return c.SendStatus(http.StatusNoContent)
}

type fiberV3RetainingServer struct {
	fiberv3params.Server
	got retained
}

func (s *fiberV3RetainingServer) GetSimpleString(c fiberv3.Ctx, param string) error {
	s.got.keep("path styled", &param)
	return c.SendStatus(http.StatusNoContent)
}

func (s *fiberV3RetainingServer) GetPassThrough(c fiberv3.Ctx, param string) error {
	s.got.keep("path passthrough", &param)
	return c.SendStatus(http.StatusNoContent)
}

func (s *fiberV3RetainingServer) GetQueryForm(c fiberv3.Ctx, params fiberv3gen.GetQueryFormParams) error {
	s.got.keep("query styled", params.Ps)
	return c.SendStatus(http.StatusNoContent)
}

func (s *fiberV3RetainingServer) GetHeader(c fiberv3.Ctx, params fiberv3gen.GetHeaderParams) error {
	s.got.keep("header passthrough", params.XPassThrough)
	s.got.keep("header styled", params.N1StartingWithNumber)
	return c.SendStatus(http.StatusNoContent)
}

func (s *fiberV3RetainingServer) GetCookie(c fiberv3.Ctx, params fiberv3gen.GetCookieParams) error {
	s.got.keep("cookie styled", params.N1s)
	return c.SendStatus(http.StatusNoContent)
}

func TestFiberParametersSurviveHandlerReturn(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	s := &fiberRetainingServer{got: retained{}}
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	fibergen.RegisterHandlers(app, s)
	go func() { _ = app.Listener(ln) }()
	t.Cleanup(func() { _ = app.Shutdown() })

	testFiberRetention(t, "http://"+ln.Addr().String(), s.got)
}

func TestFiberV3ParametersSurviveHandlerReturn(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	s := &fiberV3RetainingServer{got: retained{}}
	app := fiberv3.New()
	fiberv3gen.RegisterHandlers(app, s)
	go func() { _ = app.Listener(ln, fiberv3.ListenConfig{DisableStartupMessage: true}) }()
	t.Cleanup(func() { _ = app.Shutdown() })

	testFiberRetention(t, "http://"+ln.Addr().String(), s.got)
}

func testFiberRetention(t *testing.T, base string, got retained) {
	t.Helper()

	// One client, so both rounds reuse a single keep-alive connection and
	// therefore the same Fiber buffers.
	client := &http.Client{}
	round := func(path, query, passThrough, header, cookie string) {
		for _, target := range []string{
			"/simpleString/" + path,
			"/passThrough/" + path,
			"/queryForm?ps=" + query,
			"/header",
			"/cookie",
		} {
			req, err := http.NewRequest(http.MethodGet, base+target, nil)
			require.NoError(t, err)
			req.Header.Set("X-Pass-Through", passThrough)
			req.Header.Set("1-Starting-With-Number", header)
			req.Header.Set("Cookie", "1s="+cookie)
			resp, err := client.Do(req)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			require.Equal(t, http.StatusNoContent, resp.StatusCode, target)
		}
	}

	round("alpha", "query1", "rawhdr1", "hdrone", "cookie1")

	// The second round has the same shape as the first, so its bytes land where
	// the first round's did.
	for range 3 {
		round("OMEGA", "QUERY2", "RAWHDR2", "HDRTWO", "COOKIE2")
	}

	want := map[string]string{
		"path styled":        "alpha",
		"path passthrough":   "alpha",
		"query styled":       "query1",
		"header passthrough": "rawhdr1",
		"header styled":      "hdrone",
		"cookie styled":      "cookie1",
	}
	for name, expected := range want {
		if assert.Contains(t, got, name) {
			// Clone before comparing so a failure message shows the value as it is now.
			assert.Equal(t, expected, strings.Clone(*got[name]), name)
		}
	}
}
