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

// TestMultipleJsonContentTypes verifies Pattern H: schema "Order" collides with
// requestBody "Order" which has 3 content types that all contain "json":
//   - application/json
//   - application/merge-patch+json
//   - application/json-patch+json
//
// All three map to the same "JSON" short name via contentTypeSuffix(), which
// would trigger an infinite oscillation between context suffix ("RequestBody")
// and content type suffix ("JSON") strategies if applied per-group. The global
// phase approach lets numeric fallback break the cycle.
//
// Expected types:
//   - Order struct                    (schema keeps bare name)
//   - OrderRequestBodyJSON struct     (application/json requestBody)
//   - OrderRequestBodyJSON2 []struct  (application/json-patch+json, numeric fallback)
//   - OrderRequestBodyJSON3 struct    (application/merge-patch+json, numeric fallback)
//
// Covers: PR #2213 (multiple JSON content types in requestBody)
func TestMultipleJsonContentTypes(t *testing.T) {
	// Schema type keeps bare name "Order"
	order := Order{
		Id:      ptr("order-1"),
		Product: ptr("Widget"),
	}
	assert.Equal(t, "order-1", *order.Id)
	assert.Equal(t, "Widget", *order.Product)

	// The 3 requestBody content types should each get a unique name.
	// They all collide on "OrderRequestBodyJSON", so numeric fallback kicks in.
	// The type names below are compile-time assertions that all 3 exist and are distinct.

	// application/json requestBody
	jsonBody := OrderRequestBodyJSON{
		Id:      ptr("order-2"),
		Product: ptr("Gadget"),
	}
	assert.Equal(t, "order-2", *jsonBody.Id)

	// application/json-patch+json requestBody (numeric fallback, array type alias)
	var jsonPatch OrderRequestBodyJSON2
	assert.Nil(t, jsonPatch)

	// application/merge-patch+json requestBody (numeric fallback)
	mergePatch := OrderRequestBodyJSON3{
		Product: ptr("Gadget-patched"),
	}
	assert.Equal(t, "Gadget-patched", *mergePatch.Product)

	// CreateOrder wrapper should not collide
	var wrapper CreateOrderResponse
	assert.Nil(t, wrapper.JSON200)
	wrapper.JSON200 = &order
	assert.Equal(t, "order-1", *wrapper.JSON200.Id)
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
	var qux Qux = custom //nolint:staticcheck // explicit type needed to prove Qux aliases CustomQux
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

// TestInlineResponseWithRefProperties verifies Pattern I (oapi-codegen-exp#14):
// when a response has an inline object whose properties contain $refs to component
// schemas with x-go-type, the property-level refs must NOT produce duplicate type
// declarations. The component schemas keep their type aliases (Widget = string,
// Metadata = string), and the inline response object gets its own struct type.
//
// Covers: oapi-codegen-exp#14
func TestInlineResponseWithRefProperties(t *testing.T) {
	// Component schemas with x-go-type: string produce type aliases
	var widget Widget = "widget-value"
	assert.Equal(t, "widget-value", string(widget))

	var metadata Metadata = "metadata-value"
	assert.Equal(t, "metadata-value", string(metadata))

	// The inline response object should have fields typed by the component aliases.
	// The client wrapper for listEntities should exist and have a JSON200 field
	// pointing to the inline response type.
	var wrapper ListEntitiesResponse
	assert.Nil(t, wrapper.JSON200)
}

// TestDuplicateOneOfMembersAcrossContentTypes verifies Pattern J:
// when a response has multiple JSON content types (e.g., application/json-patch+json
// and application/json-patch-query+json) that share an identical oneOf schema with
// inline (non-$ref) members, the codegen must not emit duplicate type declarations
// for those inline members.
//
// Additionally, when a requestBody shares its name with a component schema and its
// content schemas $ref the component schema (plus one $refs a different schema),
// the collision resolver must assign unique names.
//
// Expected types:
//   - ResourceMVO struct                          (schema keeps bare name)
//   - Resource struct                             (no collision)
//   - JsonPatch []struct                          (no collision)
//   - ResourceMVORequestBodyJSON = ResourceMVO    (requestBody application/json)
//   - ResourceMVORequestBodyJSON2 = JsonPatch     (requestBody application/json-patch+json)
//   - ResourceMVORequestBodyJSON3 = ResourceMVO   (requestBody application/merge-patch+json)
//   - PatchResourceResponse struct                (client response wrapper)
//   - inline oneOf member types                   (must not be duplicated)
func TestDuplicateOneOfMembersAcrossContentTypes(t *testing.T) {
	// Schema types keep bare names
	resource := Resource{
		Id:     ptr("res-1"),
		Name:   ptr("MyResource"),
		Status: ptr("active"),
	}
	assert.Equal(t, "res-1", *resource.Id)

	resourceMVO := ResourceMVO{
		Name:   ptr("MyResource"),
		Status: ptr("active"),
	}
	assert.Equal(t, "MyResource", *resourceMVO.Name)

	// RequestBody collision resolution: schema "Resource_MVO" keeps bare name,
	// requestBody content types get suffixed.
	var reqBodyJSON ResourceMVORequestBodyJSON
	reqBodyJSON.Name = ptr("test")
	assert.Equal(t, "test", *reqBodyJSON.Name)

	var reqBodyPatch ResourceMVORequestBodyJSON2
	assert.Nil(t, reqBodyPatch) // JsonPatch alias (slice type)

	var reqBodyMerge ResourceMVORequestBodyJSON3
	reqBodyMerge.Name = ptr("merge")
	assert.Equal(t, "merge", *reqBodyMerge.Name)

	// Client response wrapper should exist. The primary assertion here
	// is that the package compiles — no duplicate oneOf member types and
	// no undefined response type names.
	var wrapper PatchResourceResponse
	assert.Nil(t, wrapper.Body)
	assert.Nil(t, wrapper.HTTPResponse)
}

// TestXGoNameOnSchemaPreserved verifies Pattern K: when a component schema
// has x-go-name, the collision resolver must use the x-go-name value as the
// schema's type name (pinned), not the original spec name.
//
// Schema "Renamer" has x-go-name: "SpecialName" and shares a name with
// response "Renamer". With correct x-go-name handling the schema becomes
// "SpecialName", so no collision exists and the response keeps "Renamer".
//
// Expected types:
//   - SpecialName struct   (schema "Renamer" pinned by x-go-name)
//   - Renamer struct       (response "Renamer" — no collision)
//
// Covers: PR #2213 review finding (x-go-name not respected by resolver)
func TestXGoNameOnSchemaPreserved(t *testing.T) {
	// Schema "Renamer" should use its x-go-name "SpecialName"
	schema := SpecialName{Label: ptr("test-label")}
	assert.Equal(t, "test-label", *schema.Label)

	// Response "Renamer" should keep its bare name (no collision with schema)
	resp := Renamer{Data: ptr("response-data")}
	assert.Equal(t, "response-data", *resp.Data)

	// Client wrapper for getRenamedSchema should reference the response type
	var wrapper GetRenamedSchemaResponse
	assert.Nil(t, wrapper.JSON200)
	wrapper.JSON200 = &resp
	assert.Equal(t, "response-data", *wrapper.JSON200.Data)
}

// TestXGoNameOnResponsePreserved verifies Pattern L: when a component response
// has x-go-name, the collision resolver must use the x-go-name value as the
// response's type name (pinned), not the original spec name.
//
// Response "Outcome" has x-go-name: "OutcomeResult" and shares a name with
// schema "Outcome". With correct x-go-name handling the response becomes
// "OutcomeResult", so no collision exists and the schema keeps "Outcome".
//
// Expected types:
//   - Outcome struct        (schema keeps bare name — no collision)
//   - OutcomeResult struct  (response "Outcome" pinned by x-go-name)
//
// Covers: PR #2213 review finding (x-go-name not respected by resolver)
func TestXGoNameOnResponsePreserved(t *testing.T) {
	// Schema "Outcome" should keep its bare name
	schema := Outcome{Value: ptr("some-value")}
	assert.Equal(t, "some-value", *schema.Value)

	// Response "Outcome" should use its x-go-name "OutcomeResult"
	resp := OutcomeResult{Result: ptr("outcome-data")}
	assert.Equal(t, "outcome-data", *resp.Result)

	// Client wrapper for getOutcome should reference the response type
	var wrapper GetOutcomeResponse
	assert.Nil(t, wrapper.JSON200)
	wrapper.JSON200 = &resp
	assert.Equal(t, "outcome-data", *wrapper.JSON200.Result)
}

// TestXGoNameOnRequestBodyPreserved verifies Pattern M: when a component
// requestBody has x-go-name, the collision resolver must use the x-go-name
// value as the requestBody's type name (pinned), not the original spec name.
//
// RequestBody "Payload" has x-go-name: "PayloadBody" and shares a name with
// schema "Payload". With correct x-go-name handling the requestBody becomes
// "PayloadBody", so no collision exists and the schema keeps "Payload".
//
// Expected types:
//   - Payload struct      (schema keeps bare name — no collision)
//   - PayloadBody struct  (requestBody "Payload" pinned by x-go-name)
//
// Covers: PR #2213 review finding (x-go-name not respected by resolver)
func TestXGoNameOnRequestBodyPreserved(t *testing.T) {
	// Schema "Payload" should keep its bare name
	schema := Payload{Content: ptr("payload-content")}
	assert.Equal(t, "payload-content", *schema.Content)

	// RequestBody "Payload" should use its x-go-name "PayloadBody"
	reqBody := PayloadBody{Data: ptr("body-data")}
	assert.Equal(t, "body-data", *reqBody.Data)

	// Client wrapper for sendPayload should reference the schema type
	var wrapper SendPayloadResponse
	assert.Nil(t, wrapper.JSON200)
	wrapper.JSON200 = &schema
	assert.Equal(t, "payload-content", *wrapper.JSON200.Content)
}

func ptr[T any](v T) *T {
	return &v
}
