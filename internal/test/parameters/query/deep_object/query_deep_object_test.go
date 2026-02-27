package querydeepobject

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testServer struct {
	object        *Object
	complexObject *ComplexObject
}

func (s *testServer) reset() {
	s.object = nil
	s.complexObject = nil
}

func (s *testServer) GetDeepObject(ctx echo.Context, params GetDeepObjectParams) error {
	s.object = &params.Obj
	if params.Complex != nil {
		s.complexObject = params.Complex
	}
	return nil
}

func setup(t *testing.T) (*testServer, *echo.Echo, string) {
	t.Helper()
	ts := &testServer{}
	e := echo.New()
	RegisterHandlers(e, ts)
	return ts, e, "http://example.com"
}

func TestServerDeepObjectParams(t *testing.T) {
	ts, e, _ := setup(t)

	expectedObject := Object{FirstName: "Alex", Role: "admin"}
	expectedComplex := ComplexObject{
		Object:  expectedObject,
		Id:      12345,
		IsAdmin: true,
	}

	t.Run("simple_object", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet,
			"/deepObject?obj[role]=admin&obj[firstName]=Alex", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.object)
		assert.EqualValues(t, expectedObject, *ts.object)
		ts.reset()
	})

	t.Run("complex_nested_object", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet,
			"/deepObject?obj[role]=admin&obj[firstName]=Alex&complex[Id]=12345&complex[IsAdmin]=true&complex[Object][role]=admin&complex[Object][firstName]=Alex", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.object)
		assert.EqualValues(t, expectedObject, *ts.object)
		require.NotNil(t, ts.complexObject)
		assert.EqualValues(t, expectedComplex, *ts.complexObject)
		ts.reset()
	})
}

func TestClientDeepObjectParams(t *testing.T) {
	ts, e, server := setup(t)

	expectedObject := Object{FirstName: "Alex", Role: "admin"}
	expectedComplex := ComplexObject{
		Object:  expectedObject,
		Id:      12345,
		IsAdmin: true,
	}

	t.Run("simple_and_complex", func(t *testing.T) {
		params := GetDeepObjectParams{
			Obj:     expectedObject,
			Complex: &expectedComplex,
		}
		req, err := NewGetDeepObjectRequest(server, &params)
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.object)
		assert.EqualValues(t, expectedObject, *ts.object)
		require.NotNil(t, ts.complexObject)
		assert.EqualValues(t, expectedComplex, *ts.complexObject)
		ts.reset()
	})
}
