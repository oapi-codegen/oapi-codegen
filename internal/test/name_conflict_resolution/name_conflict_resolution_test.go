package nameconflictresolution

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCrossSectionCollisions verifies Pattern A: when the same name "Bar"
// appears in schemas, parameters, requestBodies, and responses, the resolver
// keeps the bare name for the component schema and suffixes the others.
//
// Covers issues: #200, #254, #407, #1881, PR #292
func TestCrossSectionCollisions(t *testing.T) {
	// Schema type keeps bare name "Bar"
	bar := Bar{Value: ptr("hello")}
	assert.Equal(t, "hello", *bar.Value)

	// No collision for Bar2
	bar2 := Bar2{Value: ptr(float32(1.5))}
	assert.Equal(t, float32(1.5), *bar2.Value)

	// Parameter type gets "Parameter" suffix
	param := BarParameter("query-value")
	assert.Equal(t, "query-value", string(param))

	// RequestBody type gets "RequestBody" suffix
	reqBody := BarRequestBody{Value: ptr(42)}
	assert.Equal(t, 42, *reqBody.Value)

	// Response type gets "Response" suffix
	resp := BarResponse{
		Value1: &Bar{Value: ptr("v1")},
		Value2: &Bar2{Value: ptr(float32(2.0))},
	}
	assert.Equal(t, "v1", *resp.Value1.Value)
	assert.Equal(t, float32(2.0), *resp.Value2.Value)

	// PostFoo wrapper does not collide (unique name PostFooResponse)
	var wrapper PostFooResponse
	assert.Nil(t, wrapper.JSON200)
	// JSON200 field points to the response type BarResponse (not schema type Bar)
	wrapper.JSON200 = &resp
	assert.Equal(t, "v1", *wrapper.JSON200.Value1.Value)
}

// TestSchemaVsClientWrapper verifies Pattern B: schema "CreateItemResponse"
// collides with the client wrapper for operation "createItem". The schema
// keeps the bare name; the wrapper gets numeric fallback "CreateItemResponse2".
//
// Covers issues: #1474, #1713, #1450
func TestSchemaVsClientWrapper(t *testing.T) {
	// Schema type keeps bare name
	schema := CreateItemResponse{
		Id:   ptr("item-1"),
		Name: ptr("Widget"),
	}
	assert.Equal(t, "item-1", *schema.Id)
	assert.Equal(t, "Widget", *schema.Name)

	// Client wrapper gets numeric fallback
	var wrapper CreateItemResponse2
	assert.Nil(t, wrapper.Body)
	assert.Nil(t, wrapper.HTTPResponse)
	assert.Nil(t, wrapper.JSON200)

	// JSON200 field references the schema type, not the wrapper itself
	wrapper.JSON200 = &schema
	assert.Equal(t, "item-1", *wrapper.JSON200.Id)
}

// TestSchemaAliasVsClientWrapper verifies Pattern C: schema "ListItemsResponse"
// (a string alias) collides with the client wrapper for operation "listItems".
// The schema keeps the bare name; the wrapper gets "ListItemsResponse2".
//
// Covers issue: #1357
func TestSchemaAliasVsClientWrapper(t *testing.T) {
	// Schema type is a string alias
	var schema ListItemsResponse = "item-list"
	assert.Equal(t, "item-list", schema)

	// Client wrapper gets numeric fallback
	var wrapper ListItemsResponse2
	assert.Nil(t, wrapper.Body)
	assert.Nil(t, wrapper.HTTPResponse)
	assert.Nil(t, wrapper.JSON200)

	// JSON200 field references the schema type (string alias)
	wrapper.JSON200 = &schema
	assert.Equal(t, "item-list", *wrapper.JSON200)
}

// TestOperationNameMatchesSchema verifies Pattern D: schema "QueryResponse"
// collides with the client wrapper for operation "query" (which generates
// "QueryResponse"). The schema keeps the bare name; the wrapper gets
// "QueryResponse2".
//
// Covers issue: #255
func TestOperationNameMatchesSchema(t *testing.T) {
	// Schema type keeps bare name
	schema := QueryResponse{
		Results: &[]string{"result1", "result2"},
	}
	assert.Len(t, *schema.Results, 2)

	// Client wrapper gets numeric fallback
	var wrapper QueryResponse2
	assert.Nil(t, wrapper.JSON200)

	// JSON200 field references the schema type
	wrapper.JSON200 = &schema
	assert.Len(t, *wrapper.JSON200.Results, 2)
}

// TestSchemaMatchesOpResponse verifies Pattern E: schema "GetStatusResponse"
// collides with the client wrapper for operation "getStatus" (which generates
// "GetStatusResponse"). The schema keeps the bare name; the wrapper gets
// "GetStatusResponse2".
//
// Covers issues: #2097, #899
func TestSchemaMatchesOpResponse(t *testing.T) {
	// Schema type keeps bare name
	schema := GetStatusResponse{
		Status:    ptr("healthy"),
		Timestamp: ptr("2025-01-01T00:00:00Z"),
	}
	assert.Equal(t, "healthy", *schema.Status)
	assert.Equal(t, "2025-01-01T00:00:00Z", *schema.Timestamp)

	// Client wrapper gets numeric fallback
	var wrapper GetStatusResponse2
	assert.Nil(t, wrapper.JSON200)

	// JSON200 field references the schema type
	wrapper.JSON200 = &schema
	assert.Equal(t, "healthy", *wrapper.JSON200.Status)
}

// TestRequestBodyVsSchema verifies that "Pet" in schemas and requestBodies
// resolves correctly: the schema keeps bare name "Pet", the requestBody
// gets "PetRequestBody".
//
// Covers issues: #254, #407
func TestRequestBodyVsSchema(t *testing.T) {
	// Schema type keeps bare name
	pet := Pet{
		Id:   ptr(1),
		Name: ptr("Fluffy"),
	}
	assert.Equal(t, 1, *pet.Id)
	assert.Equal(t, "Fluffy", *pet.Name)

	// RequestBody type gets "RequestBody" suffix
	petReqBody := PetRequestBody{
		Name:    ptr("Fluffy"),
		Species: ptr("cat"),
	}
	assert.Equal(t, "Fluffy", *petReqBody.Name)
	assert.Equal(t, "cat", *petReqBody.Species)

	// CreatePet wrapper doesn't collide (unique name CreatePetResponse)
	var wrapper CreatePetResponse
	assert.Nil(t, wrapper.JSON200)

	// JSON200 field references the schema type Pet
	wrapper.JSON200 = &pet
	assert.Equal(t, "Fluffy", *wrapper.JSON200.Name)
}

// TestRefTargetPicksUpRename verifies that when an operation references a
// renamed component via $ref, the generated wrapper type uses the resolved
// (renamed) type, not the original spec name.
func TestRefTargetPicksUpRename(t *testing.T) {
	// When postFoo references $ref: '#/components/responses/Bar',
	// and response Bar is renamed to BarResponse, the wrapper's
	// JSON200 field must use BarResponse (not Bar).
	barResp := BarResponse{
		Value1: &Bar{Value: ptr("v1")},
		Value2: &Bar2{Value: ptr(float32(2.0))},
	}
	var wrapper PostFooResponse
	wrapper.JSON200 = &barResp // compile-time: JSON200 must be *BarResponse
	assert.Equal(t, "v1", *wrapper.JSON200.Value1.Value)
	assert.Equal(t, float32(2.0), *wrapper.JSON200.Value2.Value)
}

// TestExtGoTypeNameWithCollisionResolver verifies that when a component schema
// has x-go-type-name: CustomQux and collides with a response "Qux", the
// collision resolver controls the top-level Go type names while x-go-type-name
// controls the underlying type definition.
//
// Expected types:
//   - CustomQux struct   (underlying type from x-go-type-name)
//   - Qux = CustomQux    (schema keeps bare name, aliased)
//   - QuxResponse struct  (response gets suffixed)
func TestExtGoTypeNameWithCollisionResolver(t *testing.T) {
	// CustomQux is the underlying struct created by x-go-type-name
	custom := CustomQux{Label: ptr("hello")}
	assert.Equal(t, "hello", *custom.Label)

	// Qux is a type alias for CustomQux (schema keeps bare name)
	var qux Qux = custom
	assert.Equal(t, "hello", *qux.Label)

	// QuxResponse is the response type (response gets suffixed)
	quxResp := QuxResponse{Data: ptr("response-data")}
	assert.Equal(t, "response-data", *quxResp.Data)

	// GetQuxResponse client wrapper's JSON200 field uses *QuxResponse
	var wrapper GetQuxResponse
	assert.Nil(t, wrapper.JSON200)
	wrapper.JSON200 = &quxResp
	assert.Equal(t, "response-data", *wrapper.JSON200.Data)
}

// TestExtGoTypeWithCollisionResolver verifies that when a component schema has
// x-go-type: string and collides with a response "Zap", the collision resolver
// controls the top-level Go type names while x-go-type controls the target type.
//
// Expected types:
//   - Zap = string        (schema keeps bare name, x-go-type controls target)
//   - ZapResponse struct   (response gets suffixed)
func TestExtGoTypeWithCollisionResolver(t *testing.T) {
	// Zap is a string type alias (x-go-type controls the target)
	var zap Zap = "test-value"
	assert.Equal(t, "test-value", string(zap))

	// ZapResponse is the response type (response gets suffixed)
	zapResp := ZapResponse{Result: ptr("response-result")}
	assert.Equal(t, "response-result", *zapResp.Result)

	// GetZapResponse client wrapper's JSON200 field uses *ZapResponse
	var wrapper GetZapResponse
	assert.Nil(t, wrapper.JSON200)
	wrapper.JSON200 = &zapResp
	assert.Equal(t, "response-result", *wrapper.JSON200.Result)
}

func ptr[T any](v T) *T {
	return &v
}
