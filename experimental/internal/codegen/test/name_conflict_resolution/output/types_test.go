package output

import (
	"encoding/json"
	"testing"
)

// TestCrossSectionCollisions verifies Pattern A: when the same name "Bar"
// appears in schemas, parameters, requestBodies, and responses, the resolver
// keeps the bare name for the component schema and path-level schemas get
// operationId-based names.
//
// In the experimental codegen, libopenapi resolves $refs, so component-level
// requestBodies/parameters/responses appear at their path-level positions
// with operationId-based names. This inherently avoids the cross-section
// collision that plagued V2.
//
// Covers issues: #200, #254, #407, #1881, PR #292
func TestCrossSectionCollisions(t *testing.T) {
	// Schema type keeps bare name "Bar"
	bar := Bar{Value: ptr("hello")}
	assertEqual(t, "hello", *bar.Value)

	// No collision for Bar2
	bar2 := Bar2{Value: ptr(float32(1.5))}
	assertEqual(t, float32(1.5), *bar2.Value)

	// RequestBody type gets operationId-based name "PostFooJSONRequest"
	reqBody := PostFooJSONRequest{Value: ptr(42)}
	assertEqual(t, 42, *reqBody.Value)

	// Response type gets operationId-based name "PostFooJSONResponse"
	resp := PostFooJSONResponse{
		Value1: &Bar{Value: ptr("v1")},
		Value2: &Bar2{Value: ptr(float32(2.0))},
	}
	assertEqual(t, "v1", *resp.Value1.Value)
	assertEqual(t, float32(2.0), *resp.Value2.Value)
}

// TestSchemaVsOperationResponse verifies Pattern B: schema "CreateItemResponse"
// does not collide with the operation response type for "createItem".
//
// In experimental, the SimpleClient returns the schema type directly (no wrapper),
// so no collision occurs. The inline request body gets "CreateItemJSONRequest".
//
// Covers issues: #1474, #1713, #1450
func TestSchemaVsOperationResponse(t *testing.T) {
	// Schema type keeps bare name
	schema := CreateItemResponse{
		ID:   ptr("item-1"),
		Name: ptr("Widget"),
	}
	assertEqual(t, "item-1", *schema.ID)
	assertEqual(t, "Widget", *schema.Name)

	// Inline request body gets operationId-based name
	reqBody := CreateItemJSONRequest{
		Name: ptr("Widget"),
	}
	assertEqual(t, "Widget", *reqBody.Name)
}

// TestSchemaAliasVsOperationResponse verifies Pattern C: schema "ListItemsResponse"
// (a string alias) does not collide with the operation response for "listItems".
//
// Covers issue: #1357
func TestSchemaAliasVsOperationResponse(t *testing.T) {
	// Schema type is a string alias
	var schema ListItemsResponse = "item-list"
	assertEqual(t, "item-list", schema)
}

// TestOperationNameMatchesSchema verifies Pattern D: schema "QueryResponse"
// does not collide with the response type for operation "query".
//
// In experimental, path-level schemas from $ref use the target type directly
// via the SimpleClient, so "QueryResponse" (the component schema name) is used.
//
// Covers issue: #255
func TestOperationNameMatchesSchema(t *testing.T) {
	// Schema type keeps bare name
	schema := QueryResponse{
		Results: []string{"result1", "result2"},
	}
	if len(schema.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(schema.Results))
	}

	// Inline request body gets operationId-based name
	reqBody := QueryJSONRequest{
		Q: ptr("search term"),
	}
	assertEqual(t, "search term", *reqBody.Q)
}

// TestSchemaMatchesOpResponse verifies Pattern E: schema "GetStatusResponse"
// does not collide with the response type for operation "getStatus".
//
// Covers issues: #2097, #899
func TestSchemaMatchesOpResponse(t *testing.T) {
	// Schema type keeps bare name
	schema := GetStatusResponse{
		Status:    ptr("healthy"),
		Timestamp: ptr("2025-01-01T00:00:00Z"),
	}
	assertEqual(t, "healthy", *schema.Status)
	assertEqual(t, "2025-01-01T00:00:00Z", *schema.Timestamp)
}

// TestMultipleJsonContentTypes verifies Pattern H: schema "Order" collides with
// requestBody "Order" which has 3 content types that all contain "json":
//   - application/json
//   - application/merge-patch+json
//   - application/json-patch+json
//
// All three map to the same "JSON" short name via the content type namer, so
// the numeric fallback disambiguates them.
//
// Expected types:
//   - Order struct                    (schema keeps bare name)
//   - CreateOrderJSONRequest1 struct  (application/json requestBody, numeric fallback)
//   - CreateOrderJSONRequest2 struct  (application/merge-patch+json, numeric fallback)
//   - CreateOrderJSONRequest3 []PostOrdersRequest (application/json-patch+json, numeric fallback)
//
// Covers: PR #2213 (TMF622 scenario)
func TestMultipleJsonContentTypes(t *testing.T) {
	// Schema type keeps bare name "Order"
	order := Order{
		ID:      ptr("order-1"),
		Product: ptr("Widget"),
	}
	assertEqual(t, "order-1", *order.ID)
	assertEqual(t, "Widget", *order.Product)

	// application/json requestBody (numeric fallback 1)
	jsonBody := CreateOrderJSONRequest1{
		ID:      ptr("order-2"),
		Product: ptr("Gadget"),
	}
	assertEqual(t, "order-2", *jsonBody.ID)

	// application/merge-patch+json requestBody (numeric fallback 2)
	mergePatch := CreateOrderJSONRequest2{
		Product: ptr("Gadget-patched"),
	}
	assertEqual(t, "Gadget-patched", *mergePatch.Product)

	// application/json-patch+json requestBody (numeric fallback 3, array type alias)
	var jsonPatch CreateOrderJSONRequest3
	jsonPatch = append(jsonPatch, PostOrdersRequest{
		Op:    ptr("replace"),
		Path:  ptr("/product"),
		Value: ptr("Gadget-v2"),
	})
	assertEqual(t, "replace", *jsonPatch[0].Op)
	assertEqual(t, "/product", *jsonPatch[0].Path)
	assertEqual(t, "Gadget-v2", *jsonPatch[0].Value)
}

// TestRequestBodyVsSchema verifies that "Pet" in schemas and requestBodies
// resolves correctly: the schema keeps bare name "Pet", the requestBody gets
// "CreatePetJSONRequest" (operationId-based).
//
// Covers issues: #254, #407
func TestRequestBodyVsSchema(t *testing.T) {
	// Schema type keeps bare name
	pet := Pet{
		ID:   ptr(1),
		Name: ptr("Fluffy"),
	}
	assertEqual(t, 1, *pet.ID)
	assertEqual(t, "Fluffy", *pet.Name)

	// RequestBody type gets operationId-based name
	petReqBody := CreatePetJSONRequest{
		Name:    ptr("Fluffy"),
		Species: ptr("cat"),
	}
	assertEqual(t, "Fluffy", *petReqBody.Name)
	assertEqual(t, "cat", *petReqBody.Species)
}

// TestExtTypeNameOverrideWithCollisionResolver verifies that when a component schema
// has x-oapi-codegen-type-name-override: CustomQux and collides with a response "Qux",
// the type name override controls the generated type name.
//
// Expected types:
//   - CustomQux struct     (schema type from x-oapi-codegen-type-name-override)
//   - GetQuxJSONResponse struct  (response gets operationId-based name)
func TestExtTypeNameOverrideWithCollisionResolver(t *testing.T) {
	// CustomQux is the struct created by x-oapi-codegen-type-name-override
	custom := CustomQux{Label: ptr("hello")}
	assertEqual(t, "hello", *custom.Label)

	// GetQuxJSONResponse is the response type (operationId-based name)
	quxResp := GetQuxJSONResponse{Data: ptr("response-data")}
	assertEqual(t, "response-data", *quxResp.Data)
}

// TestExtGoTypeWithCollisionResolver verifies that when a component schema has
// x-go-type: string and collides with a response "Zap", the type override
// controls the generated type name.
//
// Expected types:
//   - Zap = string           (schema keeps bare name, x-go-type controls target)
//   - GetZapJSONResponse struct  (response gets operationId-based name)
func TestExtGoTypeWithCollisionResolver(t *testing.T) {
	// Zap is a string type alias (x-go-type controls the target)
	var zap Zap = "test-value"
	assertEqual(t, "test-value", zap)

	// GetZapJSONResponse is the response type (operationId-based name)
	zapResp := GetZapJSONResponse{Result: ptr("response-result")}
	assertEqual(t, "response-result", *zapResp.Result)
}

// TestJSONRoundTrip verifies that the generated types marshal/unmarshal correctly.
func TestJSONRoundTrip(t *testing.T) {
	// Bar
	bar := Bar{Value: ptr("hello")}
	data, err := json.Marshal(bar)
	if err != nil {
		t.Fatalf("marshal Bar failed: %v", err)
	}
	var decoded Bar
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal Bar failed: %v", err)
	}
	assertEqual(t, "hello", *decoded.Value)

	// Order
	order := Order{ID: ptr("o1"), Product: ptr("Widget")}
	data, err = json.Marshal(order)
	if err != nil {
		t.Fatalf("marshal Order failed: %v", err)
	}
	var decodedOrder Order
	if err := json.Unmarshal(data, &decodedOrder); err != nil {
		t.Fatalf("unmarshal Order failed: %v", err)
	}
	assertEqual(t, "o1", *decodedOrder.ID)
	assertEqual(t, "Widget", *decodedOrder.Product)
}

// TestGetOpenAPISpecJSON verifies the embedded spec can be decoded.
func TestGetOpenAPISpecJSON(t *testing.T) {
	data, err := GetOpenAPISpecJSON()
	if err != nil {
		t.Fatalf("GetOpenAPISpecJSON failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetOpenAPISpecJSON returned empty data")
	}
}

func ptr[T any](v T) *T {
	return &v
}

func assertEqual[T comparable](t *testing.T, expected, actual T) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}
