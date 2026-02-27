package queryform

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testServer struct {
	params GetQueryFormParams
}

func (s *testServer) reset() {
	s.params = GetQueryFormParams{}
}

func (s *testServer) GetQueryForm(ctx echo.Context, params GetQueryFormParams) error {
	s.params = params
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

// TestServerQueryFormParams tests server-side deserialization of form-style query parameters.
func TestServerQueryFormParams(t *testing.T) {
	ts, e, _ := setup(t)

	t.Run("unexploded/primitive", func(t *testing.T) {
		rec := doRequest(t, e, http.MethodGet, "/queryForm?p=5")
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.params.P)
		assert.EqualValues(t, int32(5), *ts.params.P)
		ts.reset()
	})

	t.Run("exploded/primitive", func(t *testing.T) {
		rec := doRequest(t, e, http.MethodGet, "/queryForm?ep=5")
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.params.Ep)
		assert.EqualValues(t, int32(5), *ts.params.Ep)
		ts.reset()
	})

	t.Run("unexploded/array", func(t *testing.T) {
		rec := doRequest(t, e, http.MethodGet, "/queryForm?a=3,4,5")
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.params.A)
		assert.EqualValues(t, []int32{3, 4, 5}, *ts.params.A)
		ts.reset()
	})

	t.Run("exploded/array", func(t *testing.T) {
		rec := doRequest(t, e, http.MethodGet, "/queryForm?ea=3&ea=4&ea=5")
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.params.Ea)
		assert.EqualValues(t, []int32{3, 4, 5}, *ts.params.Ea)
		ts.reset()
	})

	t.Run("unexploded/object", func(t *testing.T) {
		rec := doRequest(t, e, http.MethodGet, "/queryForm?o=role,admin,firstName,Alex")
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.params.O)
		assert.EqualValues(t, Object{FirstName: "Alex", Role: "admin"}, *ts.params.O)
		ts.reset()
	})

	t.Run("exploded/object", func(t *testing.T) {
		rec := doRequest(t, e, http.MethodGet, "/queryForm?role=admin&firstName=Alex")
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.params.Eo)
		assert.EqualValues(t, Object{FirstName: "Alex", Role: "admin"}, *ts.params.Eo)
		ts.reset()
	})
}

// TestClientQueryFormParams tests client serialization to server deserialization round-trip.
func TestClientQueryFormParams(t *testing.T) {
	ts, e, server := setup(t)

	var p int32 = 5
	var ep int32 = 5
	a := []int32{3, 4, 5}
	ea := []int32{3, 4, 5}
	o := Object{FirstName: "Alex", Role: "admin"}
	eo := Object{FirstName: "Alex", Role: "admin"}

	params := &GetQueryFormParams{
		P:  &p,
		Ep: &ep,
		A:  &a,
		Ea: &ea,
		O:  &o,
		Eo: &eo,
	}

	t.Run("round-trip/all", func(t *testing.T) {
		req, err := NewGetQueryFormRequest(server, params)
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		require.NotNil(t, ts.params.P)
		assert.EqualValues(t, p, *ts.params.P)

		require.NotNil(t, ts.params.Ep)
		assert.EqualValues(t, ep, *ts.params.Ep)

		require.NotNil(t, ts.params.A)
		assert.EqualValues(t, a, *ts.params.A)

		require.NotNil(t, ts.params.Ea)
		assert.EqualValues(t, ea, *ts.params.Ea)

		require.NotNil(t, ts.params.O)
		assert.EqualValues(t, o, *ts.params.O)

		require.NotNil(t, ts.params.Eo)
		assert.EqualValues(t, eo, *ts.params.Eo)

		ts.reset()
	})
}
