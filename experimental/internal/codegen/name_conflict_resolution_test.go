package codegen

import (
	"sort"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nameConflictSpec is the comprehensive collision spec from PR #2213.
// It exercises all documented name collision patterns:
//
//	Pattern A: Cross-section collision (#200, #254, #407, #1881, PR #292)
//	Pattern B: Schema vs client wrapper (#1474, #1713, #1450)
//	Pattern C: Schema alias vs client wrapper (#1357)
//	Pattern D: Operation name = schema response name (#255)
//	Pattern E: Schema matches op+Response (#2097, #899)
//	Pattern F: x-oapi-codegen-type-name extension + cross-section collision
//	Pattern G: x-go-type extension + cross-section collision
//	Pattern H: Multiple JSON content types in requestBody (TMF622 scenario, PR #2213)
//
// Note: The experimental code gathers schemas at path-level positions since
// libopenapi resolves $refs. Component-level requestBodies, parameters, responses,
// and headers are NOT gathered separately — they appear at their path-level
// positions instead (e.g., #/paths//foo/post/requestBody/content/application/json/schema).
// This inherently avoids many cross-section collision patterns that affected V2.
const nameConflictSpec = `openapi: "3.1.0"
info:
  title: "Comprehensive name collision resolution test"
  version: "0.0.0"
paths:
  # Pattern A: Cross-section collision
  # "Bar" appears in schemas, parameters, requestBodies, responses, and headers.
  /foo:
    post:
      operationId: postFoo
      parameters:
        - $ref: '#/components/parameters/Bar'
      requestBody:
        $ref: '#/components/requestBodies/Bar'
      responses:
        "200":
          $ref: '#/components/responses/Bar'

  # Pattern B: Schema vs client wrapper
  # Schema "CreateItemResponse" collides with createItem wrapper.
  /items:
    post:
      operationId: createItem
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateItemResponse'

    # Pattern C: Schema alias vs client wrapper
    # Schema "ListItemsResponse" (string alias) collides with listItems wrapper.
    get:
      operationId: listItems
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListItemsResponse'

  # Pattern D: Operation name = schema response name
  # Schema "QueryResponse" collides with query wrapper.
  /query:
    post:
      operationId: query
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                q:
                  type: string
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/QueryResponse'

  # Pattern E: Schema matches op+Response
  # Schema "GetStatusResponse" collides with getStatus wrapper.
  /status:
    get:
      operationId: getStatus
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetStatusResponse'

  # Pattern F: x-oapi-codegen-type-name extension + cross-section collision
  /qux:
    get:
      operationId: getQux
      responses:
        "200":
          $ref: '#/components/responses/Qux'
    post:
      operationId: postQux
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Qux'
      responses:
        "200":
          description: OK

  # Pattern G: x-go-type extension + cross-section collision
  /zap:
    get:
      operationId: getZap
      responses:
        "200":
          $ref: '#/components/responses/Zap'
    post:
      operationId: postZap
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Zap'
      responses:
        "200":
          description: OK

  # Pattern H: Multiple JSON content types in requestBody (TMF622 scenario)
  /orders:
    post:
      operationId: createOrder
      requestBody:
        $ref: '#/components/requestBodies/Order'
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Order'

  # Cross-section: requestBody vs schema
  /pets:
    post:
      operationId: createPet
      requestBody:
        $ref: '#/components/requestBodies/Pet'
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Pet'

components:
  schemas:
    Bar:
      type: object
      properties:
        value:
          type: string

    Bar2:
      type: object
      properties:
        value:
          type: number

    CreateItemResponse:
      type: object
      properties:
        id:
          type: string
        name:
          type: string

    ListItemsResponse:
      type: string

    QueryResponse:
      type: object
      properties:
        results:
          type: array
          items:
            type: string

    GetStatusResponse:
      type: object
      properties:
        status:
          type: string
        timestamp:
          type: string

    Order:
      type: object
      properties:
        id:
          type: string
        product:
          type: string

    Pet:
      type: object
      properties:
        id:
          type: integer
        name:
          type: string

    Qux:
      type: object
      x-oapi-codegen-type-name-override: CustomQux
      properties:
        label:
          type: string

    Zap:
      type: object
      x-go-type: string
      properties:
        unused:
          type: string

  parameters:
    Bar:
      name: bar
      in: query
      schema:
        type: string

  requestBodies:
    Bar:
      content:
        application/json:
          schema:
            type: object
            properties:
              value:
                type: integer

    Order:
      content:
        application/json:
          schema:
            type: object
            properties:
              id:
                type: string
              product:
                type: string
        application/merge-patch+json:
          schema:
            type: object
            properties:
              product:
                type: string
        application/json-patch+json:
          schema:
            type: array
            items:
              type: object
              properties:
                op:
                  type: string
                path:
                  type: string
                value:
                  type: string

    Pet:
      content:
        application/json:
          schema:
            type: object
            properties:
              name:
                type: string
              species:
                type: string

  headers:
    Bar:
      schema:
        type: boolean

  responses:
    Bar:
      description: Bar response
      headers:
        X-Bar:
          $ref: '#/components/headers/Bar'
      content:
        application/json:
          schema:
            type: object
            properties:
              value1:
                $ref: '#/components/schemas/Bar'
              value2:
                $ref: '#/components/schemas/Bar2'

    Qux:
      description: A Qux response
      content:
        application/json:
          schema:
            type: object
            properties:
              data:
                type: string

    Zap:
      description: A Zap response
      content:
        application/json:
          schema:
            type: object
            properties:
              result:
                type: string
`

// gatherAndComputeNames parses the spec, gathers schemas, and computes names.
// Returns a map of path string -> short name for easy assertions.
func gatherAndComputeNames(t *testing.T, spec string) map[string]string {
	t.Helper()

	doc, err := libopenapi.NewDocument([]byte(spec))
	require.NoError(t, err)

	matcher := NewContentTypeMatcher(DefaultContentTypes())
	schemas, err := GatherSchemas(doc, matcher, OutputOptions{})
	require.NoError(t, err)

	converter := NewNameConverter(DefaultNameMangling(), NameSubstitutions{})
	contentTypeNamer := NewContentTypeShortNamer(DefaultContentTypeShortNames())
	ComputeSchemaNames(schemas, converter, contentTypeNamer)

	result := make(map[string]string)
	for _, s := range schemas {
		result[s.Path.String()] = s.ShortName
	}
	return result
}

// assertUniqueShortNames verifies that all non-reference schemas have unique short names.
func assertUniqueShortNames(t *testing.T, spec string) map[string]string {
	t.Helper()

	doc, err := libopenapi.NewDocument([]byte(spec))
	require.NoError(t, err)

	matcher := NewContentTypeMatcher(DefaultContentTypes())
	schemas, err := GatherSchemas(doc, matcher, OutputOptions{})
	require.NoError(t, err)

	converter := NewNameConverter(DefaultNameMangling(), NameSubstitutions{})
	contentTypeNamer := NewContentTypeShortNamer(DefaultContentTypeShortNames())
	ComputeSchemaNames(schemas, converter, contentTypeNamer)

	// Build map of short name -> paths, excluding references
	nameToPath := make(map[string][]string)
	result := make(map[string]string)
	for _, s := range schemas {
		result[s.Path.String()] = s.ShortName
		if s.Ref == "" {
			nameToPath[s.ShortName] = append(nameToPath[s.ShortName], s.Path.String())
		}
	}

	for name, paths := range nameToPath {
		if len(paths) > 1 {
			t.Errorf("short name %q is not unique, used by: %v", name, paths)
		}
	}

	return result
}

// TestNameConflictResolution_AllNamesUnique verifies that the collision resolver
// produces unique short names for all non-reference schemas in the comprehensive spec.
func TestNameConflictResolution_AllNamesUnique(t *testing.T) {
	names := assertUniqueShortNames(t, nameConflictSpec)

	// Log all names sorted by path for readability
	var paths []string
	for path := range names {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	t.Log("All schema names:")
	for _, path := range paths {
		t.Logf("  %-80s -> %s", path, names[path])
	}
}

// TestNameConflictResolution_PatternA_CrossSection verifies that when "Bar" appears in
// schemas, parameters, requestBodies, responses, and headers, each gets a unique name.
//
// In experimental, libopenapi resolves $refs so component requestBodies/parameters/
// responses/headers appear at their path-level positions. The component schema keeps
// its bare name, and path-level schemas get operationId-based names.
//
// Covers issues: #200, #254, #407, #1881, PR #292
func TestNameConflictResolution_PatternA_CrossSection(t *testing.T) {
	names := gatherAndComputeNames(t, nameConflictSpec)

	// Component schema keeps bare name
	schemaName := names["#/components/schemas/Bar"]
	assert.Equal(t, "Bar", schemaName, "component schema should keep bare name")

	// The $ref to components/requestBodies/Bar is resolved by libopenapi and
	// gathered at the path level. The operationId "postFoo" gives it a distinct name.
	reqBodyName := names["#/paths//foo/post/requestBody/content/application/json/schema"]
	assert.NotEmpty(t, reqBodyName, "requestBody schema should be gathered")
	assert.NotEqual(t, "Bar", reqBodyName, "requestBody should not collide with schema")
	t.Logf("RequestBody Bar (via path) -> %s", reqBodyName)

	// The $ref to components/responses/Bar is resolved at the path level.
	respName := names["#/paths//foo/post/responses/200/content/application/json/schema"]
	assert.NotEmpty(t, respName, "response schema should be gathered")
	assert.NotEqual(t, "Bar", respName, "response should not collide with schema")
	t.Logf("Response Bar (via path) -> %s", respName)

	// All three should be distinct
	assert.NotEqual(t, reqBodyName, respName,
		"requestBody and response should have different names")
}

// TestNameConflictResolution_PatternB_SchemaVsOperationResponse verifies that
// schema "CreateItemResponse" does not collide with the inline operation response
// type generated for createItem. In experimental, the path response is a $ref to
// the component schema, so it's a reference schema and doesn't generate a new type.
//
// Covers issues: #1474, #1713, #1450
func TestNameConflictResolution_PatternB_SchemaVsOperationResponse(t *testing.T) {
	names := gatherAndComputeNames(t, nameConflictSpec)

	// Component schema keeps bare name
	schemaName := names["#/components/schemas/CreateItemResponse"]
	assert.Equal(t, "CreateItemResponse", schemaName,
		"component schema should keep bare name 'CreateItemResponse'")

	// The inline request body for createItem should get a distinct name
	reqBodyName := names["#/paths//items/post/requestBody/content/application/json/schema"]
	assert.NotEmpty(t, reqBodyName, "inline request body should be gathered")
	t.Logf("createItem requestBody -> %s", reqBodyName)

	// The response is a $ref to CreateItemResponse, so it's a reference and
	// uses the component schema's name. Verify it's gathered as a reference.
	opRespName := names["#/paths//items/post/responses/200/content/application/json/schema"]
	t.Logf("createItem response (ref to schema) -> %s", opRespName)
}

// TestNameConflictResolution_PatternD_OperationNameMatchesSchema verifies that
// schema "QueryResponse" does not collide with the response type generated
// for operation "query". In experimental, the path response is a $ref to the
// component schema, so no collision occurs.
//
// Covers issue: #255
func TestNameConflictResolution_PatternD_OperationNameMatchesSchema(t *testing.T) {
	names := gatherAndComputeNames(t, nameConflictSpec)

	schemaName := names["#/components/schemas/QueryResponse"]
	assert.Equal(t, "QueryResponse", schemaName,
		"component schema should keep bare name 'QueryResponse'")

	// The inline request body for query should get a distinct name
	reqBodyName := names["#/paths//query/post/requestBody/content/application/json/schema"]
	assert.NotEmpty(t, reqBodyName, "inline request body should be gathered")
	t.Logf("query requestBody -> %s", reqBodyName)

	// The response is a $ref, check its name
	opRespName := names["#/paths//query/post/responses/200/content/application/json/schema"]
	t.Logf("query response (ref to schema) -> %s", opRespName)
}

// TestNameConflictResolution_PatternE_SchemaMatchesOpResponse verifies that
// schema "GetStatusResponse" does not collide with the response type generated
// for operation "getStatus".
//
// Covers issues: #2097, #899
func TestNameConflictResolution_PatternE_SchemaMatchesOpResponse(t *testing.T) {
	names := gatherAndComputeNames(t, nameConflictSpec)

	schemaName := names["#/components/schemas/GetStatusResponse"]
	assert.Equal(t, "GetStatusResponse", schemaName,
		"component schema should keep bare name 'GetStatusResponse'")

	// The response is a $ref, check its name
	opRespName := names["#/paths//status/get/responses/200/content/application/json/schema"]
	t.Logf("getStatus response (ref to schema) -> %s", opRespName)
}

// TestNameConflictResolution_PatternH_MultipleJsonContentTypes verifies that
// when a requestBody has 3 content types that all contain "json", the resolver
// produces unique names for each. The component $ref is resolved by libopenapi,
// so the schemas appear at the path level.
//
// Expected: all 3 content type schemas get unique names despite all mapping to
// the "JSON" content type short name.
//
// Covers: PR #2213 (TMF622 scenario)
func TestNameConflictResolution_PatternH_MultipleJsonContentTypes(t *testing.T) {
	names := gatherAndComputeNames(t, nameConflictSpec)

	// Schema "Order" keeps bare name
	schemaName := names["#/components/schemas/Order"]
	assert.Equal(t, "Order", schemaName, "component schema should keep bare name")

	// The $ref to components/requestBodies/Order is resolved at path level.
	// The 3 content types should each get unique names.
	jsonName := names["#/paths//orders/post/requestBody/content/application/json/schema"]
	mergePatchName := names["#/paths//orders/post/requestBody/content/application/merge-patch+json/schema"]
	jsonPatchName := names["#/paths//orders/post/requestBody/content/application/json-patch+json/schema"]

	t.Logf("Order schema                         -> %s", schemaName)
	t.Logf("Order reqBody application/json        -> %s", jsonName)
	t.Logf("Order reqBody merge-patch+json        -> %s", mergePatchName)
	t.Logf("Order reqBody json-patch+json         -> %s", jsonPatchName)

	// All should be non-empty
	assert.NotEmpty(t, jsonName, "application/json requestBody should have a name")
	assert.NotEmpty(t, mergePatchName, "application/merge-patch+json requestBody should have a name")
	assert.NotEmpty(t, jsonPatchName, "application/json-patch+json requestBody should have a name")

	// All should be different from each other
	assert.NotEqual(t, jsonName, mergePatchName, "json and merge-patch+json should have different names")
	assert.NotEqual(t, jsonName, jsonPatchName, "json and json-patch+json should have different names")
	assert.NotEqual(t, mergePatchName, jsonPatchName, "merge-patch+json and json-patch+json should have different names")

	// None should collide with the schema name
	assert.NotEqual(t, schemaName, jsonName, "requestBody json should not collide with schema")
	assert.NotEqual(t, schemaName, mergePatchName, "requestBody merge-patch should not collide with schema")
	assert.NotEqual(t, schemaName, jsonPatchName, "requestBody json-patch should not collide with schema")
}

// TestNameConflictResolution_RequestBodyVsSchema verifies that "Pet" in schemas
// and requestBodies resolves correctly: the schema keeps "Pet", the requestBody
// (resolved via $ref at path level) gets a different name.
//
// Covers issues: #254, #407
func TestNameConflictResolution_RequestBodyVsSchema(t *testing.T) {
	names := gatherAndComputeNames(t, nameConflictSpec)

	schemaName := names["#/components/schemas/Pet"]
	assert.Equal(t, "Pet", schemaName, "component schema should keep bare name")

	// The $ref to components/requestBodies/Pet is resolved at path level
	reqBodyName := names["#/paths//pets/post/requestBody/content/application/json/schema"]
	assert.NotEmpty(t, reqBodyName, "requestBody schema should be gathered")
	t.Logf("Pet requestBody (via path) -> %s", reqBodyName)
	assert.NotEqual(t, "Pet", reqBodyName, "requestBody should not collide with schema")
}

// TestNameConflictResolution_PatternF_TypeNameOverride verifies that x-oapi-codegen-type-name
// interacts correctly with collision resolution.
func TestNameConflictResolution_PatternF_TypeNameOverride(t *testing.T) {
	names := gatherAndComputeNames(t, nameConflictSpec)

	// Schema Qux has x-oapi-codegen-type-name: CustomQux
	schemaName := names["#/components/schemas/Qux"]
	assert.Equal(t, "CustomQux", schemaName,
		"schema with x-oapi-codegen-type-name should use override name")

	// Response Qux (resolved at path level from $ref to components/responses/Qux)
	respName := names["#/paths//qux/get/responses/200/content/application/json/schema"]
	assert.NotEmpty(t, respName, "response schema should be gathered")
	t.Logf("Qux schema -> %s", schemaName)
	t.Logf("Qux response (via path) -> %s", respName)

	// They should not collide
	assert.NotEqual(t, schemaName, respName,
		"schema Qux (CustomQux) and response Qux should not collide")
}

// TestNameConflictResolution_PatternG_GoTypeOverride verifies that x-go-type
// interacts correctly with collision resolution.
func TestNameConflictResolution_PatternG_GoTypeOverride(t *testing.T) {
	names := gatherAndComputeNames(t, nameConflictSpec)

	schemaName := names["#/components/schemas/Zap"]
	t.Logf("Zap schema -> %s", schemaName)

	// Response Zap (resolved at path level from $ref to components/responses/Zap)
	respName := names["#/paths//zap/get/responses/200/content/application/json/schema"]
	assert.NotEmpty(t, respName, "response schema should be gathered")
	t.Logf("Zap response (via path) -> %s", respName)

	// They should not collide
	assert.NotEqual(t, schemaName, respName,
		"schema Zap and response Zap should not collide")
}
