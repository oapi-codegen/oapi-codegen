package cookieform

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testServer struct {
	params GetCookieParams
}

func (s *testServer) reset() {
	s.params = GetCookieParams{}
}

func (s *testServer) GetCookie(ctx echo.Context, params GetCookieParams) error {
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

func doRequestWithCookies(t *testing.T, e *echo.Echo, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/cookie", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// TestServerCookieParams tests server-side deserialization of cookie parameters.
func TestServerCookieParams(t *testing.T) {
	ts, e, _ := setup(t)

	expectedObject := Object{FirstName: "Alex", Role: "admin"}
	expectedArray := []int32{3, 4, 5}
	var expectedPrimitive int32 = 5

	t.Run("unexploded/primitive", func(t *testing.T) {
		rec := doRequestWithCookies(t, e, &http.Cookie{Name: "p", Value: "5"})
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.params.P)
		assert.EqualValues(t, expectedPrimitive, *ts.params.P)
		ts.reset()
	})

	t.Run("exploded/primitive", func(t *testing.T) {
		rec := doRequestWithCookies(t, e, &http.Cookie{Name: "ep", Value: "5"})
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.params.Ep)
		assert.EqualValues(t, expectedPrimitive, *ts.params.Ep)
		ts.reset()
	})

	t.Run("unexploded/array", func(t *testing.T) {
		rec := doRequestWithCookies(t, e, &http.Cookie{Name: "a", Value: "3,4,5"})
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.params.A)
		assert.EqualValues(t, expectedArray, *ts.params.A)
		ts.reset()
	})

	t.Run("exploded/array", func(t *testing.T) {
		rec := doRequestWithCookies(t, e, &http.Cookie{Name: "ea", Value: "3,4,5"})
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.params.Ea)
		assert.EqualValues(t, expectedArray, *ts.params.Ea)
		ts.reset()
	})

	t.Run("unexploded/object", func(t *testing.T) {
		rec := doRequestWithCookies(t, e, &http.Cookie{Name: "o", Value: "role,admin,firstName,Alex"})
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.params.O)
		assert.EqualValues(t, expectedObject, *ts.params.O)
		ts.reset()
	})

	t.Run("exploded/object", func(t *testing.T) {
		rec := doRequestWithCookies(t, e, &http.Cookie{Name: "eo", Value: "role=admin,firstName=Alex"})
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.params.Eo)
		assert.EqualValues(t, expectedObject, *ts.params.Eo)
		ts.reset()
	})
}

// TestClientCookieParams tests client serialization -> server deserialization round-trip.
func TestClientCookieParams(t *testing.T) {
	ts, e, server := setup(t)

	expectedObject := Object{FirstName: "Alex", Role: "admin"}
	expectedArray := []int32{3, 4, 5}
	var expectedPrimitive int32 = 5

	params := &GetCookieParams{
		P:  &expectedPrimitive,
		Ep: &expectedPrimitive,
		A:  &expectedArray,
		Ea: &expectedArray,
		O:  &expectedObject,
		Eo: &expectedObject,
	}

	req, err := NewGetCookieRequest(server, params)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	require.NotNil(t, ts.params.P)
	assert.EqualValues(t, expectedPrimitive, *ts.params.P)

	require.NotNil(t, ts.params.Ep)
	assert.EqualValues(t, expectedPrimitive, *ts.params.Ep)

	require.NotNil(t, ts.params.A)
	assert.EqualValues(t, expectedArray, *ts.params.A)

	require.NotNil(t, ts.params.Ea)
	assert.EqualValues(t, expectedArray, *ts.params.Ea)

	require.NotNil(t, ts.params.O)
	assert.EqualValues(t, expectedObject, *ts.params.O)

	require.NotNil(t, ts.params.Eo)
	assert.EqualValues(t, expectedObject, *ts.params.Eo)
}
