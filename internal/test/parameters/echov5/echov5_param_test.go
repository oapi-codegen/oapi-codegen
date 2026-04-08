package echov5params

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEchoV5ParameterRoundTrip(t *testing.T) {
	var s Server
	e := echo.New()
	RegisterHandlers(e, &s)
	testImpl(t, e)
}

func testImpl(t *testing.T, handler http.Handler) {
	t.Helper()

	server := "http://example.com"

	expectedObject := Object{
		FirstName: "Alex",
		Role:      "admin",
	}

	expectedComplexObject := ComplexObject{
		Object:  expectedObject,
		Id:      12345,
		IsAdmin: true,
	}

	expectedArray := []int32{3, 4, 5}

	var expectedPrimitive int32 = 5

	doRoundTrip := func(t *testing.T, req *http.Request, target interface{}) {
		t.Helper()
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

	t.Run("path", func(t *testing.T) {
		t.Run("simple", func(t *testing.T) {
			t.Run("primitive", func(t *testing.T) {
				req, err := NewGetSimplePrimitiveRequest(server, expectedPrimitive)
				require.NoError(t, err)
				var got int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedPrimitive, got)
			})
			t.Run("primitive explode", func(t *testing.T) {
				req, err := NewGetSimpleExplodePrimitiveRequest(server, expectedPrimitive)
				require.NoError(t, err)
				var got int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedPrimitive, got)
			})
			t.Run("array noExplode", func(t *testing.T) {
				req, err := NewGetSimpleNoExplodeArrayRequest(server, expectedArray)
				require.NoError(t, err)
				var got []int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedArray, got)
			})
			t.Run("array explode", func(t *testing.T) {
				req, err := NewGetSimpleExplodeArrayRequest(server, expectedArray)
				require.NoError(t, err)
				var got []int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedArray, got)
			})
			t.Run("object noExplode", func(t *testing.T) {
				req, err := NewGetSimpleNoExplodeObjectRequest(server, expectedObject)
				require.NoError(t, err)
				var got Object
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedObject, got)
			})
			t.Run("object explode", func(t *testing.T) {
				req, err := NewGetSimpleExplodeObjectRequest(server, expectedObject)
				require.NoError(t, err)
				var got Object
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedObject, got)
			})
		})
		t.Run("label", func(t *testing.T) {
			t.Run("primitive", func(t *testing.T) {
				req, err := NewGetLabelPrimitiveRequest(server, expectedPrimitive)
				require.NoError(t, err)
				var got int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedPrimitive, got)
			})
			t.Run("primitive explode", func(t *testing.T) {
				req, err := NewGetLabelExplodePrimitiveRequest(server, expectedPrimitive)
				require.NoError(t, err)
				var got int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedPrimitive, got)
			})
			t.Run("array noExplode", func(t *testing.T) {
				req, err := NewGetLabelNoExplodeArrayRequest(server, expectedArray)
				require.NoError(t, err)
				var got []int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedArray, got)
			})
			t.Run("array explode", func(t *testing.T) {
				req, err := NewGetLabelExplodeArrayRequest(server, expectedArray)
				require.NoError(t, err)
				var got []int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedArray, got)
			})
			t.Run("object noExplode", func(t *testing.T) {
				req, err := NewGetLabelNoExplodeObjectRequest(server, expectedObject)
				require.NoError(t, err)
				var got Object
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedObject, got)
			})
			t.Run("object explode", func(t *testing.T) {
				req, err := NewGetLabelExplodeObjectRequest(server, expectedObject)
				require.NoError(t, err)
				var got Object
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedObject, got)
			})
		})
		t.Run("matrix", func(t *testing.T) {
			t.Run("primitive", func(t *testing.T) {
				req, err := NewGetMatrixPrimitiveRequest(server, expectedPrimitive)
				require.NoError(t, err)
				var got int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedPrimitive, got)
			})
			t.Run("primitive explode", func(t *testing.T) {
				req, err := NewGetMatrixExplodePrimitiveRequest(server, expectedPrimitive)
				require.NoError(t, err)
				var got int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedPrimitive, got)
			})
			t.Run("array noExplode", func(t *testing.T) {
				req, err := NewGetMatrixNoExplodeArrayRequest(server, expectedArray)
				require.NoError(t, err)
				var got []int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedArray, got)
			})
			t.Run("array explode", func(t *testing.T) {
				req, err := NewGetMatrixExplodeArrayRequest(server, expectedArray)
				require.NoError(t, err)
				var got []int32
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedArray, got)
			})
			t.Run("object noExplode", func(t *testing.T) {
				req, err := NewGetMatrixNoExplodeObjectRequest(server, expectedObject)
				require.NoError(t, err)
				var got Object
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedObject, got)
			})
			t.Run("object explode", func(t *testing.T) {
				req, err := NewGetMatrixExplodeObjectRequest(server, expectedObject)
				require.NoError(t, err)
				var got Object
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedObject, got)
			})
		})
		t.Run("content-based", func(t *testing.T) {
			t.Run("json complex object", func(t *testing.T) {
				req, err := NewGetContentObjectRequest(server, expectedComplexObject)
				require.NoError(t, err)
				var got ComplexObject
				doRoundTrip(t, req, &got)
				assert.Equal(t, expectedComplexObject, got)
			})
			t.Run("passthrough string", func(t *testing.T) {
				req, err := NewGetPassThroughRequest(server, "hello world")
				require.NoError(t, err)
				var got string
				doRoundTrip(t, req, &got)
				assert.Equal(t, "hello world", got)
			})
		})
	})

	t.Run("query", func(t *testing.T) {
		t.Run("form", func(t *testing.T) {
			expectedArray2 := []int32{6, 7, 8}
			expectedObject2 := Object{FirstName: "Marcin", Role: "annoyed_at_swagger"}
			var expectedPrimitive2 int32 = 100
			var expectedPrimitiveString = "123;456"
			var expectedN1s = "111"

			t.Run("all params at once", func(t *testing.T) {
				params := GetQueryFormParams{
					Ea: &expectedArray, A: &expectedArray2,
					Eo: &expectedObject, O: &expectedObject2,
					Ep: &expectedPrimitive, P: &expectedPrimitive2,
					Ps: &expectedPrimitiveString, Co: &expectedComplexObject, N1s: &expectedN1s,
				}
				req, err := NewGetQueryFormRequest(server, &params)
				require.NoError(t, err)
				var got GetQueryFormParams
				doRoundTrip(t, req, &got)
				assert.EqualValues(t, params, got)
			})
			t.Run("exploded array only", func(t *testing.T) {
				params := GetQueryFormParams{Ea: &expectedArray}
				req, err := NewGetQueryFormRequest(server, &params)
				require.NoError(t, err)
				var got GetQueryFormParams
				doRoundTrip(t, req, &got)
				require.NotNil(t, got.Ea)
				assert.Equal(t, expectedArray, *got.Ea)
			})
			t.Run("unexploded array only", func(t *testing.T) {
				params := GetQueryFormParams{A: &expectedArray}
				req, err := NewGetQueryFormRequest(server, &params)
				require.NoError(t, err)
				var got GetQueryFormParams
				doRoundTrip(t, req, &got)
				require.NotNil(t, got.A)
				assert.Equal(t, expectedArray, *got.A)
			})
			t.Run("exploded object only", func(t *testing.T) {
				params := GetQueryFormParams{Eo: &expectedObject}
				req, err := NewGetQueryFormRequest(server, &params)
				require.NoError(t, err)
				var got GetQueryFormParams
				doRoundTrip(t, req, &got)
				require.NotNil(t, got.Eo)
				assert.Equal(t, expectedObject, *got.Eo)
			})
			t.Run("unexploded object only", func(t *testing.T) {
				params := GetQueryFormParams{O: &expectedObject}
				req, err := NewGetQueryFormRequest(server, &params)
				require.NoError(t, err)
				var got GetQueryFormParams
				doRoundTrip(t, req, &got)
				require.NotNil(t, got.O)
				assert.Equal(t, expectedObject, *got.O)
			})
			t.Run("primitive with semicolon", func(t *testing.T) {
				params := GetQueryFormParams{Ps: &expectedPrimitiveString}
				req, err := NewGetQueryFormRequest(server, &params)
				require.NoError(t, err)
				var got GetQueryFormParams
				doRoundTrip(t, req, &got)
				require.NotNil(t, got.Ps)
				assert.Equal(t, expectedPrimitiveString, *got.Ps)
			})
		})
		t.Run("deepObject", func(t *testing.T) {
			params := GetDeepObjectParams{DeepObj: expectedComplexObject}
			req, err := NewGetDeepObjectRequest(server, &params)
			require.NoError(t, err)
			var got GetDeepObjectParams
			doRoundTrip(t, req, &got)
			assert.Equal(t, expectedComplexObject, got.DeepObj)
		})
		t.Run("spaceDelimited", func(t *testing.T) {
		})
		t.Run("pipeDelimited", func(t *testing.T) {
		})
	})

	t.Run("header", func(t *testing.T) {
		expectedArray2 := []int32{6, 7, 8}
		expectedObject2 := Object{FirstName: "Marcin", Role: "annoyed_at_swagger"}
		var expectedPrimitive2 int32 = 100
		var expectedN1s = "111"

		t.Run("all params at once", func(t *testing.T) {
			params := GetHeaderParams{
				XPrimitive: &expectedPrimitive2, XPrimitiveExploded: &expectedPrimitive,
				XArrayExploded: &expectedArray, XArray: &expectedArray2,
				XObjectExploded: &expectedObject, XObject: &expectedObject2,
				XComplexObject: &expectedComplexObject, N1StartingWithNumber: &expectedN1s,
			}
			req, err := NewGetHeaderRequest(server, &params)
			require.NoError(t, err)
			var got GetHeaderParams
			doRoundTrip(t, req, &got)
			assert.EqualValues(t, params, got)
		})
		t.Run("primitive only", func(t *testing.T) {
			params := GetHeaderParams{XPrimitive: &expectedPrimitive}
			req, err := NewGetHeaderRequest(server, &params)
			require.NoError(t, err)
			var got GetHeaderParams
			doRoundTrip(t, req, &got)
			require.NotNil(t, got.XPrimitive)
			assert.Equal(t, expectedPrimitive, *got.XPrimitive)
		})
		t.Run("array only", func(t *testing.T) {
			params := GetHeaderParams{XArray: &expectedArray}
			req, err := NewGetHeaderRequest(server, &params)
			require.NoError(t, err)
			var got GetHeaderParams
			doRoundTrip(t, req, &got)
			require.NotNil(t, got.XArray)
			assert.Equal(t, expectedArray, *got.XArray)
		})
		t.Run("object only", func(t *testing.T) {
			params := GetHeaderParams{XObject: &expectedObject}
			req, err := NewGetHeaderRequest(server, &params)
			require.NoError(t, err)
			var got GetHeaderParams
			doRoundTrip(t, req, &got)
			require.NotNil(t, got.XObject)
			assert.Equal(t, expectedObject, *got.XObject)
		})
	})

	t.Run("cookie", func(t *testing.T) {
		expectedArray2 := []int32{6, 7, 8}
		expectedObject2 := Object{FirstName: "Marcin", Role: "annoyed_at_swagger"}
		var expectedPrimitive2 int32 = 100
		var expectedN1s = "111"

		t.Run("all params at once", func(t *testing.T) {
			params := GetCookieParams{
				P: &expectedPrimitive2, Ep: &expectedPrimitive,
				Ea: &expectedArray, A: &expectedArray2,
				Eo: &expectedObject, O: &expectedObject2,
				Co: &expectedComplexObject, N1s: &expectedN1s,
			}
			req, err := NewGetCookieRequest(server, &params)
			require.NoError(t, err)
			var got GetCookieParams
			doRoundTrip(t, req, &got)
			assert.EqualValues(t, params, got)
		})
		t.Run("primitive only", func(t *testing.T) {
			params := GetCookieParams{P: &expectedPrimitive}
			req, err := NewGetCookieRequest(server, &params)
			require.NoError(t, err)
			var got GetCookieParams
			doRoundTrip(t, req, &got)
			require.NotNil(t, got.P)
			assert.Equal(t, expectedPrimitive, *got.P)
		})
		t.Run("array only", func(t *testing.T) {
			params := GetCookieParams{A: &expectedArray}
			req, err := NewGetCookieRequest(server, &params)
			require.NoError(t, err)
			var got GetCookieParams
			doRoundTrip(t, req, &got)
			require.NotNil(t, got.A)
			assert.Equal(t, expectedArray, *got.A)
		})
		t.Run("object only", func(t *testing.T) {
			params := GetCookieParams{O: &expectedObject}
			req, err := NewGetCookieRequest(server, &params)
			require.NoError(t, err)
			var got GetCookieParams
			doRoundTrip(t, req, &got)
			require.NotNil(t, got.O)
			assert.Equal(t, expectedObject, *got.O)
		})
	})
}
