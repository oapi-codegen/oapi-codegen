package content

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testServer struct {
	complexObject *ComplexObject
	passThrough   *string
}

func (s *testServer) reset() {
	s.complexObject = nil
	s.passThrough = nil
}

func (s *testServer) GetPathJson(ctx echo.Context, param ComplexObject) error {
	s.complexObject = &param
	return nil
}

func (s *testServer) GetPathText(ctx echo.Context, param string) error {
	s.passThrough = &param
	return nil
}

func (s *testServer) GetQueryJson(ctx echo.Context, params GetQueryJsonParams) error {
	if params.Obj != nil {
		s.complexObject = params.Obj
	}
	return nil
}

func (s *testServer) GetHeaderJson(ctx echo.Context, params GetHeaderJsonParams) error {
	if params.XObject != nil {
		s.complexObject = params.XObject
	}
	return nil
}

func (s *testServer) GetCookieJson(ctx echo.Context, params GetCookieJsonParams) error {
	if params.Obj != nil {
		s.complexObject = params.Obj
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

func TestServerContentParams(t *testing.T) {
	ts, e, _ := setup(t)

	expectedObject := Object{FirstName: "Alex", Role: "admin"}
	expectedComplex := ComplexObject{
		Object:  expectedObject,
		Id:      12345,
		IsAdmin: true,
	}
	marshaledComplex, err := json.Marshal(expectedComplex)
	require.NoError(t, err)

	t.Run("path/json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet,
			fmt.Sprintf("/pathJson/%s", string(marshaledComplex)), nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.complexObject)
		assert.EqualValues(t, expectedComplex, *ts.complexObject)
		ts.reset()
	})

	t.Run("path/text", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/pathText/some%20string", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.passThrough)
		assert.EqualValues(t, "some string", *ts.passThrough)
		ts.reset()
	})

	t.Run("query/json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet,
			fmt.Sprintf("/queryJson?obj=%s", string(marshaledComplex)), nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.complexObject)
		assert.EqualValues(t, expectedComplex, *ts.complexObject)
		ts.reset()
	})

	t.Run("header/json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/headerJson", nil)
		req.Header.Set("X-Object", string(marshaledComplex))
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.complexObject)
		assert.EqualValues(t, expectedComplex, *ts.complexObject)
		ts.reset()
	})

	t.Run("cookie/json", func(t *testing.T) {
		// Cookie values cannot contain raw JSON (quotes are invalid).
		// The runtime URL-encodes JSON cookie values, so we must do the same for the server-side test.
		encoded := url.QueryEscape(string(marshaledComplex))
		req := httptest.NewRequest(http.MethodGet, "/cookieJson", nil)
		req.AddCookie(&http.Cookie{Name: "obj", Value: encoded})
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.complexObject)
		assert.EqualValues(t, expectedComplex, *ts.complexObject)
		ts.reset()
	})
}

func TestClientContentParams(t *testing.T) {
	ts, e, server := setup(t)

	expectedObject := Object{FirstName: "Alex", Role: "admin"}
	expectedComplex := ComplexObject{
		Object:  expectedObject,
		Id:      12345,
		IsAdmin: true,
	}

	t.Run("path/json", func(t *testing.T) {
		req, err := NewGetPathJsonRequest(server, expectedComplex)
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.complexObject)
		assert.EqualValues(t, expectedComplex, *ts.complexObject)
		ts.reset()
	})

	t.Run("path/text", func(t *testing.T) {
		req, err := NewGetPathTextRequest(server, "some string")
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.passThrough)
		assert.Equal(t, "some string", *ts.passThrough)
		ts.reset()
	})

	t.Run("query/json", func(t *testing.T) {
		params := GetQueryJsonParams{Obj: &expectedComplex}
		req, err := NewGetQueryJsonRequest(server, &params)
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.complexObject)
		assert.EqualValues(t, expectedComplex, *ts.complexObject)
		ts.reset()
	})

	t.Run("header/json", func(t *testing.T) {
		params := GetHeaderJsonParams{XObject: &expectedComplex}
		req, err := NewGetHeaderJsonRequest(server, &params)
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.complexObject)
		assert.EqualValues(t, expectedComplex, *ts.complexObject)
		ts.reset()
	})

	t.Run("cookie/json", func(t *testing.T) {
		params := GetCookieJsonParams{Obj: &expectedComplex}
		req, err := NewGetCookieJsonRequest(server, &params)
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotNil(t, ts.complexObject)
		assert.EqualValues(t, expectedComplex, *ts.complexObject)
		ts.reset()
	})
}
