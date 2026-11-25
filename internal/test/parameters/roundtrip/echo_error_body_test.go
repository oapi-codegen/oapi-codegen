package parametersroundtrip_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	echov5 "github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	echoparams "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/roundtrip/echo"
	echogen "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/roundtrip/echo/gen"
	echov5params "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/roundtrip/echov5"
	echov5gen "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/roundtrip/echov5/gen"
)

// echoBindErrorCases covers every parameter location whose bind error used to be
// echoed to the client verbatim. The response must quote only what the caller
// sent, never the Go error text produced by the runtime or strconv.
var echoBindErrorCases = []struct {
	name   string
	target string
	header http.Header
	cookie *http.Cookie
	want   string
}{
	{name: "path", target: "/simplePrimitive/foo", want: "Invalid format for parameter param: 'foo'"},
	{name: "query", target: "/queryForm?p=foo", want: "Invalid format for parameter p: 'foo'"},
	{name: "query with repeated keys", target: "/queryForm?ea=1&ea=bar", want: "Invalid format for parameter ea: '1,bar'"},
	{name: "header", target: "/header", header: http.Header{"X-Primitive": {"foo"}}, want: "Invalid format for parameter X-Primitive: 'foo'"},
	{name: "cookie", target: "/cookie", cookie: &http.Cookie{Name: "p", Value: "foo"}, want: "Invalid format for parameter p: 'foo'"},
}

func testEchoBindErrorBody(t *testing.T, handler http.Handler) {
	t.Helper()
	for _, tc := range echoBindErrorCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.target, nil)
			for k, v := range tc.header {
				req.Header[k] = v
			}
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			require.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.want)
			assert.NotContains(t, rec.Body.String(), "strconv", "Go error text must not reach the response body")
			assert.NotContains(t, rec.Body.String(), "parsing", "Go error text must not reach the response body")
		})
	}
}

func TestEchoBindErrorBodyHidesGoError(t *testing.T) {
	e := echo.New()
	echogen.RegisterHandlers(e, &echoparams.Server{})
	testEchoBindErrorBody(t, e)
}

func TestEchoV5BindErrorBodyHidesGoError(t *testing.T) {
	e := echov5.New()
	echov5gen.RegisterHandlers(e, &echov5params.Server{})
	testEchoBindErrorBody(t, e)
}
