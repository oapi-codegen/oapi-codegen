package pathsimple

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testServer struct {
	primitive *int32
	array     []int32
	object    *Object
}

func (s *testServer) reset() {
	s.primitive = nil
	s.array = nil
	s.object = nil
}

func (s *testServer) GetNoExplodePrimitive(ctx echo.Context, param int32) error {
	s.primitive = &param
	return nil
}

func (s *testServer) GetNoExplodeArray(ctx echo.Context, param []int32) error {
	s.array = param
	return nil
}

func (s *testServer) GetNoExplodeObject(ctx echo.Context, param Object) error {
	s.object = &param
	return nil
}

func (s *testServer) GetExplodePrimitive(ctx echo.Context, param int32) error {
	s.primitive = &param
	return nil
}

func (s *testServer) GetExplodeArray(ctx echo.Context, param []int32) error {
	s.array = param
	return nil
}

func (s *testServer) GetExplodeObject(ctx echo.Context, param Object) error {
	s.object = &param
	return nil
}

func doRequest(t *testing.T, e *echo.Echo, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func setup(t *testing.T) (*testServer, *echo.Echo, string) {
	t.Helper()
	ts := &testServer{}
	e := echo.New()
	RegisterHandlers(e, ts)
	server := "http://example.com"
	return ts, e, server
}

// TestServerSimplePathParams tests server-side deserialization of simple-style path parameters.
func TestServerSimplePathParams(t *testing.T) {
	ts, e, _ := setup(t)

	expectedObject := Object{FirstName: "Alex", Role: "admin"}
	expectedArray := []int32{3, 4, 5}
	var expectedPrimitive int32 = 5

	t.Run("noExplode/primitive", func(t *testing.T) {
		rec := doRequest(t, e, http.MethodGet, "/noExplodePrimitive/5")
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.primitive)
		assert.EqualValues(t, expectedPrimitive, *ts.primitive)
		ts.reset()
	})

	t.Run("noExplode/array", func(t *testing.T) {
		rec := doRequest(t, e, http.MethodGet, "/noExplodeArray/3,4,5")
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.EqualValues(t, expectedArray, ts.array)
		ts.reset()
	})

	t.Run("noExplode/object", func(t *testing.T) {
		rec := doRequest(t, e, http.MethodGet, "/noExplodeObject/role,admin,firstName,Alex")
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.object)
		assert.EqualValues(t, expectedObject, *ts.object)
		ts.reset()
	})

	t.Run("explode/primitive", func(t *testing.T) {
		rec := doRequest(t, e, http.MethodGet, "/explodePrimitive/5")
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.primitive)
		assert.EqualValues(t, expectedPrimitive, *ts.primitive)
		ts.reset()
	})

	t.Run("explode/array", func(t *testing.T) {
		rec := doRequest(t, e, http.MethodGet, "/explodeArray/3,4,5")
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.EqualValues(t, expectedArray, ts.array)
		ts.reset()
	})

	t.Run("explode/object", func(t *testing.T) {
		rec := doRequest(t, e, http.MethodGet, "/explodeObject/role=admin,firstName=Alex")
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.object)
		assert.EqualValues(t, expectedObject, *ts.object)
		ts.reset()
	})
}

// TestClientSimplePathParams tests client serialization → server deserialization round-trip.
func TestClientSimplePathParams(t *testing.T) {
	ts, e, server := setup(t)

	expectedObject := Object{FirstName: "Alex", Role: "admin"}
	expectedArray := []int32{3, 4, 5}
	var expectedPrimitive int32 = 5

	t.Run("noExplode/primitive", func(t *testing.T) {
		req, err := NewGetNoExplodePrimitiveRequest(server, expectedPrimitive)
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.primitive)
		assert.EqualValues(t, expectedPrimitive, *ts.primitive)
		ts.reset()
	})

	t.Run("noExplode/array", func(t *testing.T) {
		req, err := NewGetNoExplodeArrayRequest(server, expectedArray)
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.EqualValues(t, expectedArray, ts.array)
		ts.reset()
	})

	t.Run("noExplode/object", func(t *testing.T) {
		req, err := NewGetNoExplodeObjectRequest(server, expectedObject)
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.object)
		assert.EqualValues(t, expectedObject, *ts.object)
		ts.reset()
	})

	t.Run("explode/primitive", func(t *testing.T) {
		req, err := NewGetExplodePrimitiveRequest(server, expectedPrimitive)
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.primitive)
		assert.EqualValues(t, expectedPrimitive, *ts.primitive)
		ts.reset()
	})

	t.Run("explode/array", func(t *testing.T) {
		req, err := NewGetExplodeArrayRequest(server, expectedArray)
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.EqualValues(t, expectedArray, ts.array)
		ts.reset()
	})

	t.Run("explode/object", func(t *testing.T) {
		req, err := NewGetExplodeObjectRequest(server, expectedObject)
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.object)
		assert.EqualValues(t, expectedObject, *ts.object)
		ts.reset()
	})
}
