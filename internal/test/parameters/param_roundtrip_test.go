package parameters_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gorilla/mux"
	"github.com/kataras/iris/v12"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	chiparams "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/chi"
	chigen "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/chi/gen"
	paramclient "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/client/gen"
	echoparams "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/echo"
	echogen "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/echo/gen"
	fiberparams "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/fiber"
	fibergen "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/fiber/gen"
	ginparams "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/gin"
	gingen "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/gin/gen"
	gorillaparams "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/gorilla"
	gorillgen "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/gorilla/gen"
	irisparams "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/iris"
	irisgen "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/iris/gen"
	stdhttpparams "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/stdhttp"
	stdhttpgen "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/stdhttp/gen"
)

func TestEchoParameterRoundTrip(t *testing.T) {
	var s echoparams.Server
	e := echo.New()
	echogen.RegisterHandlers(e, &s)
	testImpl(t, e)
}

func TestChiParameterRoundTrip(t *testing.T) {
	var s chiparams.Server
	r := chi.NewRouter()
	handler := chigen.HandlerFromMux(&s, r)
	testImpl(t, handler)
}

func TestGinParameterRoundTrip(t *testing.T) {
	var s ginparams.Server
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	gingen.RegisterHandlers(r, &s)
	testImpl(t, r)
}

func TestGorillaParameterRoundTrip(t *testing.T) {
	var s gorillaparams.Server
	r := mux.NewRouter()
	handler := gorillgen.HandlerFromMux(&s, r)
	testImpl(t, handler)
}

func TestIrisParameterRoundTrip(t *testing.T) {
	var s irisparams.Server
	app := iris.New()
	irisgen.RegisterHandlers(app, &s)
	testImpl(t, app)
}

func TestFiberParameterRoundTrip(t *testing.T) {
	var s fiberparams.Server
	app := fiber.New()
	fibergen.RegisterHandlers(app, &s)
	testImpl(t, adaptor.FiberApp(app))
}

func TestStdHttpParameterRoundTrip(t *testing.T) {
	// The OpenAPI spec includes a path parameter named "1param" which starts
	// with a digit. Go's stdlib ServeMux requires wildcard names to be valid
	// Go identifiers, so registering this route panics. This is a known
	// stdhttp panics because net/http.ServeMux rejects wildcard names
	// starting with a digit ("1param"). Skip until codegen sanitizes the name.
	t.Skip("stdhttp panics on path param name starting with digit (1param) — see #2306")
	var s stdhttpparams.Server
	handler := stdhttpgen.Handler(&s)
	testImpl(t, handler)
}

// testImpl runs the full parameter roundtrip test suite against any http.Handler.
// The generated client serializes Go values into an HTTP request, the server
// deserializes them and echoes them back as JSON, and we compare the response
// body against the original values.
func testImpl(t *testing.T, handler http.Handler) {
	t.Helper()

	server := "http://example.com"

	expectedObject := paramclient.Object{
		FirstName: "Alex",
		Role:      "admin",
	}

	expectedComplexObject := paramclient.ComplexObject{
		Object:  expectedObject,
		Id:      12345,
		IsAdmin: true,
	}

	expectedArray := []int32{3, 4, 5}

	var expectedPrimitive int32 = 5

	// doRoundTrip sends a request to the handler, asserts 200, and decodes the JSON response.
	doRoundTrip := func(t *testing.T, req *http.Request, target interface{}) {
		t.Helper()
		// The generated client produces requests via http.NewRequest which
		// leaves RequestURI empty. Some adapters (notably Fiber) need it
		// set to route correctly.
		req.RequestURI = req.URL.RequestURI()
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if !assert.Equal(t, http.StatusOK, rec.Code, "server returned %d; body: %s", rec.Code, rec.Body.String()) {
			return
		}
		if target != nil {
			require.NoError(t, json.NewDecoder(rec.Body).Decode(target), "failed to decode response body")
		}
	}

	// =========================================================================
	// Path Parameters
	// =========================================================================
	t.Run("path", func(t *testing.T) {
		t.Run("simple", func(t *testing.T) {
			t.Run("primitive", func(t *testing.T) {
				req, err := paramclient.NewGetSimplePrimitiveRequest(server, expectedPrimitive)
				require.NoError(t, err)
				var got int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedPrimitive, got)
			})

			t.Run("primitive explode", func(t *testing.T) {
				req, err := paramclient.NewGetSimpleExplodePrimitiveRequest(server, expectedPrimitive)
				require.NoError(t, err)
				var got int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedPrimitive, got)
			})

			t.Run("array noExplode", func(t *testing.T) {
				req, err := paramclient.NewGetSimpleNoExplodeArrayRequest(server, expectedArray)
				require.NoError(t, err)
				var got []int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedArray, got)
			})

			t.Run("array explode", func(t *testing.T) {
				req, err := paramclient.NewGetSimpleExplodeArrayRequest(server, expectedArray)
				require.NoError(t, err)
				var got []int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedArray, got)
			})

			t.Run("object noExplode", func(t *testing.T) {
				req, err := paramclient.NewGetSimpleNoExplodeObjectRequest(server, expectedObject)
				require.NoError(t, err)
				var got paramclient.Object
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedObject, got)
			})

			t.Run("object explode", func(t *testing.T) {
				req, err := paramclient.NewGetSimpleExplodeObjectRequest(server, expectedObject)
				require.NoError(t, err)
				var got paramclient.Object
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedObject, got)
			})
		})

		t.Run("label", func(t *testing.T) {
			t.Run("primitive", func(t *testing.T) {
				req, err := paramclient.NewGetLabelPrimitiveRequest(server, expectedPrimitive)
				require.NoError(t, err)
				var got int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedPrimitive, got)
			})

			t.Run("primitive explode", func(t *testing.T) {
				req, err := paramclient.NewGetLabelExplodePrimitiveRequest(server, expectedPrimitive)
				require.NoError(t, err)
				var got int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedPrimitive, got)
			})

			t.Run("array noExplode", func(t *testing.T) {
				req, err := paramclient.NewGetLabelNoExplodeArrayRequest(server, expectedArray)
				require.NoError(t, err)
				var got []int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedArray, got)
			})

			t.Run("array explode", func(t *testing.T) {
				req, err := paramclient.NewGetLabelExplodeArrayRequest(server, expectedArray)
				require.NoError(t, err)
				var got []int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedArray, got)
			})

			t.Run("object noExplode", func(t *testing.T) {
				req, err := paramclient.NewGetLabelNoExplodeObjectRequest(server, expectedObject)
				require.NoError(t, err)
				var got paramclient.Object
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedObject, got)
			})

			t.Run("object explode", func(t *testing.T) {
				req, err := paramclient.NewGetLabelExplodeObjectRequest(server, expectedObject)
				require.NoError(t, err)
				var got paramclient.Object
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedObject, got)
			})
		})

		t.Run("matrix", func(t *testing.T) {
			t.Run("primitive", func(t *testing.T) {
				req, err := paramclient.NewGetMatrixPrimitiveRequest(server, expectedPrimitive)
				require.NoError(t, err)
				var got int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedPrimitive, got)
			})

			t.Run("primitive explode", func(t *testing.T) {
				req, err := paramclient.NewGetMatrixExplodePrimitiveRequest(server, expectedPrimitive)
				require.NoError(t, err)
				var got int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedPrimitive, got)
			})

			t.Run("array noExplode", func(t *testing.T) {
				req, err := paramclient.NewGetMatrixNoExplodeArrayRequest(server, expectedArray)
				require.NoError(t, err)
				var got []int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedArray, got)
			})

			t.Run("array explode", func(t *testing.T) {
				req, err := paramclient.NewGetMatrixExplodeArrayRequest(server, expectedArray)
				require.NoError(t, err)
				var got []int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedArray, got)
			})

			t.Run("object noExplode", func(t *testing.T) {
				req, err := paramclient.NewGetMatrixNoExplodeObjectRequest(server, expectedObject)
				require.NoError(t, err)
				var got paramclient.Object
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedObject, got)
			})

			t.Run("object explode", func(t *testing.T) {
				req, err := paramclient.NewGetMatrixExplodeObjectRequest(server, expectedObject)
				require.NoError(t, err)
				var got paramclient.Object
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedObject, got)
			})
		})

		t.Run("content-based", func(t *testing.T) {
			t.Run("json complex object", func(t *testing.T) {
				req, err := paramclient.NewGetContentObjectRequest(server, expectedComplexObject)
				require.NoError(t, err)
				var got paramclient.ComplexObject
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedComplexObject, got)
			})

			t.Run("passthrough string", func(t *testing.T) {
				req, err := paramclient.NewGetPassThroughRequest(server, "hello world")
				require.NoError(t, err)
				var got string
				doRoundTrip(t, req, &got)
				assert.Equal(t, "hello world", got)
			})
		})
	})

	// =========================================================================
	// Query Parameters
	// =========================================================================
	t.Run("query", func(t *testing.T) {
		t.Run("form", func(t *testing.T) {
			expectedArray2 := []int32{6, 7, 8}
			expectedObject2 := paramclient.Object{FirstName: "Marcin", Role: "annoyed_at_swagger"}
			var expectedPrimitive2 int32 = 100
			var expectedPrimitiveString = "123;456"
			var expectedN1s = "111"

			t.Run("all params at once", func(t *testing.T) {
				params := paramclient.GetQueryFormParams{
					Ea:  &expectedArray,
					A:   &expectedArray2,
					Eo:  &expectedObject,
					O:   &expectedObject2,
					Ep:  &expectedPrimitive,
					P:   &expectedPrimitive2,
					Ps:  &expectedPrimitiveString,
					Co:  &expectedComplexObject,
					N1s: &expectedN1s,
				}
				req, err := paramclient.NewGetQueryFormRequest(server, &params)
				require.NoError(t, err)
				var got paramclient.GetQueryFormParams
				doRoundTrip(t, req, &got)
				assert.EqualValues(t, params, got)
			})

			t.Run("exploded array only", func(t *testing.T) {
				params := paramclient.GetQueryFormParams{Ea: &expectedArray}
				req, err := paramclient.NewGetQueryFormRequest(server, &params)
				require.NoError(t, err)
				var got paramclient.GetQueryFormParams
				doRoundTrip(t, req, &got)
				require.NotNil(t, got.Ea)
				assert.Equal(t, expectedArray, *got.Ea)
			})

			t.Run("unexploded array only", func(t *testing.T) {
				params := paramclient.GetQueryFormParams{A: &expectedArray}
				req, err := paramclient.NewGetQueryFormRequest(server, &params)
				require.NoError(t, err)
				var got paramclient.GetQueryFormParams
				doRoundTrip(t, req, &got)
				require.NotNil(t, got.A)
				assert.Equal(t, expectedArray, *got.A)
			})

			t.Run("exploded object only", func(t *testing.T) {
				params := paramclient.GetQueryFormParams{Eo: &expectedObject}
				req, err := paramclient.NewGetQueryFormRequest(server, &params)
				require.NoError(t, err)
				var got paramclient.GetQueryFormParams
				doRoundTrip(t, req, &got)
				require.NotNil(t, got.Eo)
				assert.Equal(t, expectedObject, *got.Eo)
			})

			t.Run("unexploded object only", func(t *testing.T) {
				params := paramclient.GetQueryFormParams{O: &expectedObject}
				req, err := paramclient.NewGetQueryFormRequest(server, &params)
				require.NoError(t, err)
				var got paramclient.GetQueryFormParams
				doRoundTrip(t, req, &got)
				require.NotNil(t, got.O)
				assert.Equal(t, expectedObject, *got.O)
			})

			t.Run("primitive with semicolon", func(t *testing.T) {
				params := paramclient.GetQueryFormParams{Ps: &expectedPrimitiveString}
				req, err := paramclient.NewGetQueryFormRequest(server, &params)
				require.NoError(t, err)
				var got paramclient.GetQueryFormParams
				doRoundTrip(t, req, &got)
				require.NotNil(t, got.Ps)
				assert.Equal(t, expectedPrimitiveString, *got.Ps)
			})
		})

		t.Run("deepObject", func(t *testing.T) {
			params := paramclient.GetDeepObjectParams{DeepObj: expectedComplexObject}
			req, err := paramclient.NewGetDeepObjectRequest(server, &params)
			require.NoError(t, err)
			var got paramclient.GetDeepObjectParams
			doRoundTrip(t, req, &got)
			assert.Equal(t, expectedComplexObject, got.DeepObj)
		})

		t.Run("spaceDelimited", func(t *testing.T) {
		})

		t.Run("pipeDelimited", func(t *testing.T) {
		})
	})

	// =========================================================================
	// Header Parameters
	// =========================================================================
	t.Run("header", func(t *testing.T) {
		expectedArray2 := []int32{6, 7, 8}
		expectedObject2 := paramclient.Object{FirstName: "Marcin", Role: "annoyed_at_swagger"}
		var expectedPrimitive2 int32 = 100
		var expectedN1s = "111"

		t.Run("all params at once", func(t *testing.T) {
			params := paramclient.GetHeaderParams{
				XPrimitive:           &expectedPrimitive2,
				XPrimitiveExploded:   &expectedPrimitive,
				XArrayExploded:       &expectedArray,
				XArray:               &expectedArray2,
				XObjectExploded:      &expectedObject,
				XObject:              &expectedObject2,
				XComplexObject:       &expectedComplexObject,
				N1StartingWithNumber: &expectedN1s,
			}
			req, err := paramclient.NewGetHeaderRequest(server, &params)
			require.NoError(t, err)
			var got paramclient.GetHeaderParams
			doRoundTrip(t, req, &got)
			assert.EqualValues(t, params, got)
		})

		t.Run("primitive only", func(t *testing.T) {
			params := paramclient.GetHeaderParams{XPrimitive: &expectedPrimitive}
			req, err := paramclient.NewGetHeaderRequest(server, &params)
			require.NoError(t, err)
			var got paramclient.GetHeaderParams
			doRoundTrip(t, req, &got)
			require.NotNil(t, got.XPrimitive)
			assert.Equal(t, expectedPrimitive, *got.XPrimitive)
		})

		t.Run("array only", func(t *testing.T) {
			params := paramclient.GetHeaderParams{XArray: &expectedArray}
			req, err := paramclient.NewGetHeaderRequest(server, &params)
			require.NoError(t, err)
			var got paramclient.GetHeaderParams
			doRoundTrip(t, req, &got)
			require.NotNil(t, got.XArray)
			assert.Equal(t, expectedArray, *got.XArray)
		})

		t.Run("object only", func(t *testing.T) {
			params := paramclient.GetHeaderParams{XObject: &expectedObject}
			req, err := paramclient.NewGetHeaderRequest(server, &params)
			require.NoError(t, err)
			var got paramclient.GetHeaderParams
			doRoundTrip(t, req, &got)
			require.NotNil(t, got.XObject)
			assert.Equal(t, expectedObject, *got.XObject)
		})
	})

	// =========================================================================
	// Cookie Parameters
	// =========================================================================
	t.Run("cookie", func(t *testing.T) {
		expectedArray2 := []int32{6, 7, 8}
		expectedObject2 := paramclient.Object{FirstName: "Marcin", Role: "annoyed_at_swagger"}
		var expectedPrimitive2 int32 = 100
		var expectedN1s = "111"

		t.Run("all params at once", func(t *testing.T) {
			params := paramclient.GetCookieParams{
				P:   &expectedPrimitive2,
				Ep:  &expectedPrimitive,
				Ea:  &expectedArray,
				A:   &expectedArray2,
				Eo:  &expectedObject,
				O:   &expectedObject2,
				Co:  &expectedComplexObject,
				N1s: &expectedN1s,
			}
			req, err := paramclient.NewGetCookieRequest(server, &params)
			require.NoError(t, err)
			var got paramclient.GetCookieParams
			doRoundTrip(t, req, &got)
			assert.EqualValues(t, params, got)
		})

		t.Run("primitive only", func(t *testing.T) {
			params := paramclient.GetCookieParams{P: &expectedPrimitive}
			req, err := paramclient.NewGetCookieRequest(server, &params)
			require.NoError(t, err)
			var got paramclient.GetCookieParams
			doRoundTrip(t, req, &got)
			require.NotNil(t, got.P)
			assert.Equal(t, expectedPrimitive, *got.P)
		})

		t.Run("array only", func(t *testing.T) {
			params := paramclient.GetCookieParams{A: &expectedArray}
			req, err := paramclient.NewGetCookieRequest(server, &params)
			require.NoError(t, err)
			var got paramclient.GetCookieParams
			doRoundTrip(t, req, &got)
			require.NotNil(t, got.A)
			assert.Equal(t, expectedArray, *got.A)
		})

		t.Run("object only", func(t *testing.T) {
			params := paramclient.GetCookieParams{O: &expectedObject}
			req, err := paramclient.NewGetCookieRequest(server, &params)
			require.NoError(t, err)
			var got paramclient.GetCookieParams
			doRoundTrip(t, req, &got)
			require.NotNil(t, got.O)
			assert.Equal(t, expectedObject, *got.O)
		})
	})
}
