package headersimple

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testServer struct {
	params GetHeaderParams
}

func (s *testServer) reset() {
	s.params = GetHeaderParams{}
}

func (s *testServer) GetHeader(ctx echo.Context, params GetHeaderParams) error {
	s.params = params
	return nil
}

func setup(t *testing.T) (*testServer, *echo.Echo, string) {
	t.Helper()
	ts := &testServer{}
	e := echo.New()
	RegisterHandlers(e, ts)
	server := "http://example.com"
	return ts, e, server
}

func doHeaderRequest(t *testing.T, e *echo.Echo, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/header", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// TestServerSimpleHeaderParams tests server-side deserialization of simple-style header parameters.
func TestServerSimpleHeaderParams(t *testing.T) {
	ts, e, _ := setup(t)

	expectedObject := Object{FirstName: "Alex", Role: "admin"}
	expectedArray := []int32{3, 4, 5}
	var expectedPrimitive int32 = 5

	t.Run("primitive", func(t *testing.T) {
		rec := doHeaderRequest(t, e, map[string]string{
			"X-Primitive": "5",
		})
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.params.XPrimitive)
		assert.EqualValues(t, expectedPrimitive, *ts.params.XPrimitive)
		ts.reset()
	})

	t.Run("primitiveExploded", func(t *testing.T) {
		rec := doHeaderRequest(t, e, map[string]string{
			"X-Primitive-Exploded": "5",
		})
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.params.XPrimitiveExploded)
		assert.EqualValues(t, expectedPrimitive, *ts.params.XPrimitiveExploded)
		ts.reset()
	})

	t.Run("array", func(t *testing.T) {
		rec := doHeaderRequest(t, e, map[string]string{
			"X-Array": "3,4,5",
		})
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.params.XArray)
		assert.EqualValues(t, expectedArray, *ts.params.XArray)
		ts.reset()
	})

	t.Run("arrayExploded", func(t *testing.T) {
		rec := doHeaderRequest(t, e, map[string]string{
			"X-Array-Exploded": "3,4,5",
		})
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.params.XArrayExploded)
		assert.EqualValues(t, expectedArray, *ts.params.XArrayExploded)
		ts.reset()
	})

	t.Run("object", func(t *testing.T) {
		rec := doHeaderRequest(t, e, map[string]string{
			"X-Object": "role,admin,firstName,Alex",
		})
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.params.XObject)
		assert.EqualValues(t, expectedObject, *ts.params.XObject)
		ts.reset()
	})

	t.Run("objectExploded", func(t *testing.T) {
		rec := doHeaderRequest(t, e, map[string]string{
			"X-Object-Exploded": "role=admin,firstName=Alex",
		})
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.params.XObjectExploded)
		assert.EqualValues(t, expectedObject, *ts.params.XObjectExploded)
		ts.reset()
	})
}

// TestClientSimpleHeaderParams tests client serialization → server deserialization round-trip.
func TestClientSimpleHeaderParams(t *testing.T) {
	ts, e, server := setup(t)

	expectedObject := Object{FirstName: "Alex", Role: "admin"}
	expectedArray := []int32{3, 4, 5}
	var expectedPrimitive int32 = 5

	params := &GetHeaderParams{
		XPrimitive:         &expectedPrimitive,
		XPrimitiveExploded: &expectedPrimitive,
		XArray:             &expectedArray,
		XArrayExploded:     &expectedArray,
		XObject:            &expectedObject,
		XObjectExploded:    &expectedObject,
	}

	req, err := NewGetHeaderRequest(server, params)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	require.NotNil(t, ts.params.XPrimitive)
	assert.EqualValues(t, expectedPrimitive, *ts.params.XPrimitive)

	require.NotNil(t, ts.params.XPrimitiveExploded)
	assert.EqualValues(t, expectedPrimitive, *ts.params.XPrimitiveExploded)

	require.NotNil(t, ts.params.XArray)
	assert.EqualValues(t, expectedArray, *ts.params.XArray)

	require.NotNil(t, ts.params.XArrayExploded)
	assert.EqualValues(t, expectedArray, *ts.params.XArrayExploded)

	require.NotNil(t, ts.params.XObject)
	assert.EqualValues(t, expectedObject, *ts.params.XObject)

	require.NotNil(t, ts.params.XObjectExploded)
	assert.EqualValues(t, expectedObject, *ts.params.XObjectExploded)
}
