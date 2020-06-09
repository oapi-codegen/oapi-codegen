package parameters

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tidepool-org/oapi-codegen/pkg/testutil"
)

type testServer struct {
	array         []int32
	object        *Object
	complexObject *ComplexObject
	passThrough   *string
	primitive     *int32
	cookieParams  *GetCookieParams
	queryParams   *GetQueryFormParams
	headerParams  *GetHeaderParams
}

func (t *testServer) reset() {
	t.array = nil
	t.object = nil
	t.complexObject = nil
	t.passThrough = nil
	t.primitive = nil
	t.cookieParams = nil
	t.queryParams = nil
	t.headerParams = nil
}

//  (GET /contentObject/{param})
func (t *testServer) GetContentObject(ctx echo.Context, param ComplexObject) error {
	t.complexObject = &param
	return nil
}

//  (GET /labelExplodeArray/{.param*})
func (t *testServer) GetLabelExplodeArray(ctx echo.Context, param []int32) error {
	t.array = param
	return nil
}

//  (GET /labelExplodeObject/{.param*})
func (t *testServer) GetLabelExplodeObject(ctx echo.Context, param Object) error {
	t.object = &param
	return nil
}

//  (GET /labelNoExplodeArray/{.param})
func (t *testServer) GetLabelNoExplodeArray(ctx echo.Context, param []int32) error {
	t.array = param
	return nil
}

//  (GET /labelNoExplodeObject/{.param})
func (t *testServer) GetLabelNoExplodeObject(ctx echo.Context, param Object) error {
	t.object = &param
	return nil
}

//  (GET /matrixExplodeArray/{.param*})
func (t *testServer) GetMatrixExplodeArray(ctx echo.Context, param []int32) error {
	t.array = param
	return nil
}

//  (GET /matrixExplodeObject/{.param*})
func (t *testServer) GetMatrixExplodeObject(ctx echo.Context, param Object) error {
	t.object = &param
	return nil
}

//  (GET /matrixNoExplodeArray/{.param})
func (t *testServer) GetMatrixNoExplodeArray(ctx echo.Context, param []int32) error {
	t.array = param
	return nil
}

//  (GET /matrixNoExplodeObject/{.param})
func (t *testServer) GetMatrixNoExplodeObject(ctx echo.Context, param Object) error {
	t.object = &param
	return nil
}

//  (GET /simpleExplodeArray/{param*})
func (t *testServer) GetSimpleExplodeArray(ctx echo.Context, param []int32) error {
	t.array = param
	return nil
}

//  (GET /simpleExplodeObject/{param*})
func (t *testServer) GetSimpleExplodeObject(ctx echo.Context, param Object) error {
	t.object = &param
	return nil
}

//  (GET /simpleNoExplodeArray/{param})
func (t *testServer) GetSimpleNoExplodeArray(ctx echo.Context, param []int32) error {
	t.array = param
	return nil
}

//  (GET /simpleNoExplodeObject/{param})
func (t *testServer) GetSimpleNoExplodeObject(ctx echo.Context, param Object) error {
	t.object = &param
	return nil
}

//  (GET /passThrough/{param})
func (t *testServer) GetPassThrough(ctx echo.Context, param string) error {
	t.passThrough = &param
	return nil
}

// (GET /queryDeepObject)
func (t *testServer) GetDeepObject(ctx echo.Context, params GetDeepObjectParams) error {
	t.complexObject = &params.DeepObj
	return nil
}

//  (GET /simplePrimitive/{param})
func (t *testServer) GetSimplePrimitive(ctx echo.Context, param int32) error {
	t.primitive = &param
	return nil
}

//  (GET /queryForm)
func (t *testServer) GetQueryForm(ctx echo.Context, params GetQueryFormParams) error {
	t.queryParams = &params
	if params.Ea != nil {
		t.array = *params.Ea
	}
	if params.A != nil {
		t.array = *params.A
	}
	if params.Eo != nil {
		t.object = params.Eo
	}
	if params.O != nil {
		t.object = params.O
	}
	if params.P != nil {
		t.primitive = params.P
	}
	if params.Ep != nil {
		t.primitive = params.Ep
	}
	if params.Co != nil {
		t.complexObject = params.Co
	}
	return nil
}

//  (GET /header)
func (t *testServer) GetHeader(ctx echo.Context, params GetHeaderParams) error {
	t.headerParams = &params
	if params.XPrimitive != nil {
		t.primitive = params.XPrimitive
	}
	if params.XPrimitiveExploded != nil {
		t.primitive = params.XPrimitiveExploded
	}
	if params.XArray != nil {
		t.array = *params.XArray
	}
	if params.XArrayExploded != nil {
		t.array = *params.XArrayExploded
	}
	if params.XObject != nil {
		t.object = params.XObject
	}
	if params.XObjectExploded != nil {
		t.object = params.XObjectExploded
	}
	if params.XComplexObject != nil {
		t.complexObject = params.XComplexObject
	}
	return nil
}

//  (GET /cookie)
func (t *testServer) GetCookie(ctx echo.Context, params GetCookieParams) error {
	t.cookieParams = &params
	if params.Ea != nil {
		t.array = *params.Ea
	}
	if params.A != nil {
		t.array = *params.A
	}
	if params.Eo != nil {
		t.object = params.Eo
	}
	if params.O != nil {
		t.object = params.O
	}
	if params.P != nil {
		t.primitive = params.P
	}
	if params.Ep != nil {
		t.primitive = params.Ep
	}
	if params.Co != nil {
		t.complexObject = params.Co
	}
	return nil
}

func TestParameterBinding(t *testing.T) {
	var ts testServer
	e := echo.New()
	e.Use(middleware.Logger())
	RegisterHandlers(e, &ts)

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

	// Check the passthrough case
	//  (GET /passThrough/{param})
	result := testutil.NewRequest().Get("/passThrough/some%20string").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	require.NotNil(t, ts.passThrough)
	assert.EqualValues(t, "some string", *ts.passThrough)
	ts.reset()

	// Check JSON marshaling of Content based parameter
	//  (GET /contentObject/{param})
	marshaledComplexObject, err := json.Marshal(expectedComplexObject)
	assert.NoError(t, err)
	q := fmt.Sprintf("/contentObject/%s", string(marshaledComplexObject))
	result = testutil.NewRequest().Get(q).Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedComplexObject, ts.complexObject)
	ts.reset()

	//  (GET /labelExplodeArray/{.param*})
	result = testutil.NewRequest().Get("/labelExplodeArray/.3.4.5").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, expectedArray, ts.array)
	ts.reset()

	//  (GET /labelExplodeObject/{.param*})
	result = testutil.NewRequest().Get("/labelExplodeObject/.role=admin.firstName=Alex").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedObject, ts.object)
	ts.reset()

	//  (GET /labelNoExplodeArray/{.param})
	result = testutil.NewRequest().Get("/labelNoExplodeArray/.3,4,5").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, expectedArray, ts.array)
	ts.reset()

	//  (GET /labelNoExplodeObject/{.param})
	result = testutil.NewRequest().Get("/labelNoExplodeObject/.role,admin,firstName,Alex").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedObject, ts.object)
	ts.reset()

	//  (GET /matrixExplodeArray/{.param*})
	uri := "/matrixExplodeArray/;id=3;id=4;id=5"
	result = testutil.NewRequest().Get(uri).Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, expectedArray, ts.array)
	ts.reset()

	//  (GET /matrixExplodeObject/{.param*})
	uri = "/matrixExplodeObject/;role=admin;firstName=Alex"
	result = testutil.NewRequest().Get(uri).Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedObject, ts.object)
	ts.reset()

	//  (GET /matrixNoExplodeArray/{.param})
	result = testutil.NewRequest().Get("/matrixNoExplodeArray/;id=3,4,5").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, expectedArray, ts.array)
	ts.reset()

	//  (GET /matrixNoExplodeObject/{.param})
	result = testutil.NewRequest().Get("/matrixNoExplodeObject/;id=role,admin,firstName,Alex").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedObject, ts.object)
	ts.reset()

	//  (GET /simpleExplodeArray/{param*})
	result = testutil.NewRequest().Get("/simpleExplodeArray/3,4,5").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, expectedArray, ts.array)
	ts.reset()

	//  (GET /simpleExplodeObject/{param*})
	result = testutil.NewRequest().Get("/simpleExplodeObject/role=admin,firstName=Alex").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedObject, ts.object)
	ts.reset()

	//  (GET /simpleNoExplodeArray/{param})
	result = testutil.NewRequest().Get("/simpleNoExplodeArray/3,4,5").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, expectedArray, ts.array)
	ts.reset()

	//  (GET /simpleNoExplodeObject/{param})
	result = testutil.NewRequest().Get("/simpleNoExplodeObject/role,admin,firstName,Alex").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedObject, ts.object)
	ts.reset()

	//  (GET /simplePrimitive/{param})
	result = testutil.NewRequest().Get("/simplePrimitive/5").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedPrimitive, ts.primitive)
	ts.reset()

	// ---------------------- Test Form Query Parameters ----------------------
	//  (GET /queryForm)

	// unexploded array
	result = testutil.NewRequest().Get("/queryForm?a=3,4,5").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, expectedArray, ts.array)
	ts.reset()

	// exploded array
	result = testutil.NewRequest().Get("/queryForm?ea=3&ea=4&ea=5").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, expectedArray, ts.array)
	ts.reset()

	// unexploded object
	result = testutil.NewRequest().Get("/queryForm?o=role,admin,firstName,Alex").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedObject, ts.object)
	ts.reset()

	// exploded object
	result = testutil.NewRequest().Get("/queryForm?role=admin&firstName=Alex").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedObject, ts.object)
	ts.reset()

	// exploded primitive
	result = testutil.NewRequest().Get("/queryForm?ep=5").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedPrimitive, ts.primitive)
	ts.reset()

	// unexploded primitive
	result = testutil.NewRequest().Get("/queryForm?p=5").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedPrimitive, ts.primitive)
	ts.reset()

	// complex object
	q = fmt.Sprintf("/queryForm?co=%s", string(marshaledComplexObject))
	result = testutil.NewRequest().Get(q).Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedComplexObject, ts.complexObject)
	ts.reset()

	// complex object via deepObject
	do := `deepObj[Id]=12345&deepObj[IsAdmin]=true&deepObj[Object][firstName]=Alex&deepObj[Object][role]=admin`
	q = "/queryDeepObject?" + do
	result = testutil.NewRequest().Get(q).Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedComplexObject, ts.complexObject)
	ts.reset()


	// ---------------------- Test Header Query Parameters --------------------

	// unexploded header primitive.
	result = testutil.NewRequest().WithHeader("X-Primitive", "5").Get("/header").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedPrimitive, ts.primitive)
	ts.reset()

	// exploded header primitive.
	result = testutil.NewRequest().WithHeader("X-Primitive-Exploded", "5").Get("/header").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedPrimitive, ts.primitive)
	ts.reset()

	// unexploded header array
	result = testutil.NewRequest().WithHeader("X-Array", "3,4,5").Get("/header").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, expectedArray, ts.array)
	ts.reset()

	// exploded header array
	result = testutil.NewRequest().WithHeader("X-Array-Exploded", "3,4,5").Get("/header").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, expectedArray, ts.array)
	ts.reset()

	// unexploded header object
	result = testutil.NewRequest().WithHeader("X-Object",
		"role,admin,firstName,Alex").Get("/header").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedObject, ts.object)
	ts.reset()

	// exploded header object
	result = testutil.NewRequest().WithHeader("X-Object-Exploded",
		"role=admin,firstName=Alex").Get("/header").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedObject, ts.object)
	ts.reset()

	// complex object
	result = testutil.NewRequest().WithHeader("X-Complex-Object",
		string(marshaledComplexObject)).Get("/header").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedComplexObject, ts.complexObject)
	ts.reset()

	// ------------------------- Test Cookie Parameters ------------------------
	result = testutil.NewRequest().WithCookieNameValue("p", "5").Get("/cookie").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedPrimitive, ts.primitive)
	ts.reset()

	result = testutil.NewRequest().WithCookieNameValue("ep", "5").Get("/cookie").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedPrimitive, ts.primitive)
	ts.reset()

	result = testutil.NewRequest().WithCookieNameValue("a", "3,4,5").Get("/cookie").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, expectedArray, ts.array)
	ts.reset()

	result = testutil.NewRequest().WithCookieNameValue(
		"o", "role,admin,firstName,Alex").Get("/cookie").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.EqualValues(t, &expectedObject, ts.object)
	ts.reset()
}

func doRequest(t *testing.T, e *echo.Echo, code int, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, code, rec.Code)
	return rec
}

func TestClientPathParams(t *testing.T) {
	var ts testServer
	e := echo.New()
	e.Use(middleware.Logger())
	RegisterHandlers(e, &ts)
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

	req, err := NewGetPassThroughRequest(server, "some string")
	assert.NoError(t, err)
	doRequest(t, e, http.StatusOK, req)
	require.NotNil(t, ts.passThrough)
	assert.Equal(t, "some string", *ts.passThrough)
	ts.reset()

	req, err = NewGetContentObjectRequest(server, expectedComplexObject)
	assert.NoError(t, err)
	doRequest(t, e, http.StatusOK, req)
	assert.EqualValues(t, &expectedComplexObject, ts.complexObject)
	ts.reset()

	// Label style
	req, err = NewGetLabelExplodeArrayRequest(server, expectedArray)
	assert.NoError(t, err)
	doRequest(t, e, http.StatusOK, req)
	assert.EqualValues(t, expectedArray, ts.array)
	ts.reset()

	req, err = NewGetLabelNoExplodeArrayRequest(server, expectedArray)
	assert.NoError(t, err)
	doRequest(t, e, http.StatusOK, req)
	assert.EqualValues(t, expectedArray, ts.array)
	ts.reset()

	req, err = NewGetLabelExplodeObjectRequest(server, expectedObject)
	assert.NoError(t, err)
	doRequest(t, e, http.StatusOK, req)
	assert.EqualValues(t, &expectedObject, ts.object)
	ts.reset()

	req, err = NewGetLabelNoExplodeObjectRequest(server, expectedObject)
	assert.NoError(t, err)
	doRequest(t, e, http.StatusOK, req)
	assert.EqualValues(t, &expectedObject, ts.object)
	ts.reset()

	// Matrix style
	req, err = NewGetMatrixExplodeArrayRequest(server, expectedArray)
	assert.NoError(t, err)
	doRequest(t, e, http.StatusOK, req)
	assert.EqualValues(t, expectedArray, ts.array)
	ts.reset()

	req, err = NewGetMatrixNoExplodeArrayRequest(server, expectedArray)
	assert.NoError(t, err)
	doRequest(t, e, http.StatusOK, req)
	assert.EqualValues(t, expectedArray, ts.array)
	ts.reset()

	req, err = NewGetMatrixExplodeObjectRequest(server, expectedObject)
	assert.NoError(t, err)
	doRequest(t, e, http.StatusOK, req)
	assert.EqualValues(t, &expectedObject, ts.object)
	ts.reset()

	req, err = NewGetMatrixNoExplodeObjectRequest(server, expectedObject)
	assert.NoError(t, err)
	doRequest(t, e, http.StatusOK, req)
	assert.EqualValues(t, &expectedObject, ts.object)
	ts.reset()

	// Simple style
	req, err = NewGetSimpleExplodeArrayRequest(server, expectedArray)
	assert.NoError(t, err)
	doRequest(t, e, http.StatusOK, req)
	assert.EqualValues(t, expectedArray, ts.array)
	ts.reset()

	req, err = NewGetSimpleNoExplodeArrayRequest(server, expectedArray)
	assert.NoError(t, err)
	doRequest(t, e, http.StatusOK, req)
	assert.EqualValues(t, expectedArray, ts.array)
	ts.reset()

	req, err = NewGetSimpleExplodeObjectRequest(server, expectedObject)
	assert.NoError(t, err)
	doRequest(t, e, http.StatusOK, req)
	assert.EqualValues(t, &expectedObject, ts.object)
	ts.reset()

	req, err = NewGetSimpleNoExplodeObjectRequest(server, expectedObject)
	assert.NoError(t, err)
	doRequest(t, e, http.StatusOK, req)
	assert.EqualValues(t, &expectedObject, ts.object)
	ts.reset()

	req, err = NewGetSimplePrimitiveRequest(server, expectedPrimitive)
	assert.NoError(t, err)
	doRequest(t, e, http.StatusOK, req)
	assert.EqualValues(t, &expectedPrimitive, ts.primitive)
	ts.reset()
}

func TestClientQueryParams(t *testing.T) {
	var ts testServer
	e := echo.New()
	e.Use(middleware.Logger())
	RegisterHandlers(e, &ts)
	server := "http://example.com"

	expectedObject1 := Object{
		FirstName: "Alex",
		Role:      "admin",
	}
	expectedObject2 := Object{
		FirstName: "Marcin",
		Role:      "annoyed_at_swagger",
	}

	expectedComplexObject := ComplexObject{
		Object:  expectedObject2,
		Id:      12345,
		IsAdmin: true,
	}

	expectedArray1 := []int32{3, 4, 5}
	expectedArray2 := []int32{6, 7, 8}

	var expectedPrimitive1 int32 = 5
	var expectedPrimitive2 int32 = 100

	// Check query params
	qParams := GetQueryFormParams{
		Ea: &expectedArray1,
		A:  &expectedArray2,
		Eo: &expectedObject1,
		O:  &expectedObject2,
		Ep: &expectedPrimitive1,
		P:  &expectedPrimitive2,
		Co: &expectedComplexObject,
	}

	req, err := NewGetQueryFormRequest(server, &qParams)
	assert.NoError(t, err)
	doRequest(t, e, http.StatusOK, req)
	require.NotNil(t, ts.queryParams)
	assert.EqualValues(t, qParams, *ts.queryParams)
	ts.reset()

	// Check cookie params
	cParams := GetCookieParams{
		Ea: &expectedArray1,
		A:  &expectedArray2,
		Eo: &expectedObject1,
		O:  &expectedObject2,
		Ep: &expectedPrimitive1,
		P:  &expectedPrimitive2,
		Co: &expectedComplexObject,
	}
	req, err = NewGetCookieRequest(server, &cParams)
	assert.NoError(t, err)
	doRequest(t, e, http.StatusOK, req)
	require.NotNil(t, ts.cookieParams)
	assert.EqualValues(t, cParams, *ts.cookieParams)
	ts.reset()

	// Check Header parameters
	hParams := GetHeaderParams{
		XArrayExploded:     &expectedArray1,
		XArray:             &expectedArray2,
		XObjectExploded:    &expectedObject1,
		XObject:            &expectedObject2,
		XPrimitiveExploded: &expectedPrimitive1,
		XPrimitive:         &expectedPrimitive2,
		XComplexObject:     &expectedComplexObject,
	}
	req, err = NewGetHeaderRequest(server, &hParams)
	assert.NoError(t, err)
	doRequest(t, e, http.StatusOK, req)
	require.NotNil(t, ts.headerParams)
	assert.EqualValues(t, hParams, *ts.headerParams)
	ts.reset()
}
