package parameters

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deepmap/oapi-codegen/pkg/testutil"
)

type testServer struct {
	array         []int32
	object        *Object
	complexObject *ComplexObject
	passThrough   *string
	primitive     *int32
}

func (t *testServer) reset() {
	t.array = nil
	t.object = nil
	t.complexObject = nil
	t.passThrough = nil
	t.primitive = nil
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

//  (GET /simplePrimitive/{param})
func (t *testServer) GetSimplePrimitive(ctx echo.Context, param int32) error {
	t.primitive = &param
	return nil
}

//  (GET /queryForm)
func (t *testServer) GetQueryForm(ctx echo.Context, params GetQueryFormParams) error {
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
		Object: expectedObject,
		Id:     "12345",
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
}
