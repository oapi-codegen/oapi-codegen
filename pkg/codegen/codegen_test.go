package codegen

import (
	_ "embed"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
)

const (
	remoteRefFile = `https://raw.githubusercontent.com/oapi-codegen/oapi-codegen/master/examples/petstore-expanded` +
		`/petstore-expanded.yaml`
	remoteRefImport = `github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded`
)

func TestExampleOpenAPICodeGeneration(t *testing.T) {

	// Input vars for code generation:
	packageName := "testswagger"
	opts := Configuration{
		PackageName: packageName,
		Generate: GenerateOptions{
			EchoServer:   true,
			Client:       true,
			Models:       true,
			EmbeddedSpec: true,
		},
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	// Get a spec from the test definition in this file:
	swagger, err := loader.LoadFromData([]byte(testOpenAPIDefinition))
	assert.NoError(t, err)

	// Run our code generation:
	code, err := Generate(swagger, opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, code)

	// Check that we have valid (formattable) code:
	_, err = format.Source([]byte(code))
	assert.NoError(t, err)

	// Check that we have a package:
	assert.Contains(t, code, "package testswagger")

	// Check that response structs are generated correctly:
	assert.Contains(t, code, "type GetTestByNameResponse struct {")

	// Check that response structs contains fallbacks to interface for invalid types:
	// Here an invalid array with no items.
	assert.Contains(t, code, `
type GetTestByNameResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	// JSON200 the response for an HTTP 200 `+"`application/json`"+` response
	JSON200 *[]Test
	// XML200 the response for an HTTP 200 `+"`application/xml`"+` response
	XML200 *[]Test
	// JSON422 the response for an HTTP 422 `+"`application/json`"+` response
	JSON422 *[]interface{}
	// XML422 the response for an HTTP 422 `+"`application/xml`"+` response
	XML422 *[]interface{}
	// JSONDefault the response for an HTTP default `+"`application/json`"+` response
	JSONDefault *Error
}`)

	// Check that the helper methods are generated correctly:
	assert.Contains(t, code, "func (r GetTestByNameResponse) Status() string {")
	assert.Contains(t, code, "func (r GetTestByNameResponse) StatusCode() int {")
	assert.Contains(t, code, "func ParseGetTestByNameResponse(rsp *http.Response) (*GetTestByNameResponse, error) {")

	// Check the client method signatures:
	assert.Contains(t, code, "type GetTestByNameParams struct {")
	assert.Contains(t, code, "Top *int `form:\"$top,omitempty\" json:\"$top,omitempty\"`")
	assert.Contains(t, code, "func (c *Client) GetTestByName(ctx context.Context, name string, params *GetTestByNameParams, reqEditors ...RequestEditorFn) (*http.Response, error) {")
	assert.Contains(t, code, "func (c *ClientWithResponses) GetTestByNameWithResponse(ctx context.Context, name string, params *GetTestByNameParams, reqEditors ...RequestEditorFn) (*GetTestByNameResponse, error) {")
	assert.Contains(t, code, "FavouriteBirds     *[]*string          `json:\"favourite_birds,omitempty\"`")
	assert.Contains(t, code, "DetestedBirds      *[]string           `json:\"detested_birds,omitempty\"`")
	assert.Contains(t, code, "SlicedBirds        []string            `json:\"sliced_birds\"`")
	assert.Contains(t, code, "ForgettableBirds   *map[string]*string `json:\"forgettable_birds,omitempty\"`")
	assert.Contains(t, code, "MemorableBirds     *map[string]string  `json:\"memorable_birds,omitempty\"`")
	assert.Contains(t, code, "VeryMemorableBirds map[string]string   `json:\"very_memorable_birds\"`")
	assert.Contains(t, code, "DeadSince          *time.Time          `json:\"dead_since,omitempty\" tag1:\"value1\" tag2:\"value2\"`")
	assert.Contains(t, code, "VeryDeadSince      time.Time           `json:\"very_dead_since\"`")
	assert.Contains(t, code, "type EnumTestNumerics int")
	assert.Contains(t, code, "N2 EnumTestNumerics = 2")
	assert.Contains(t, code, "type EnumTestEnumNames int")
	assert.Contains(t, code, "Two  EnumTestEnumNames = 2")
	assert.Contains(t, code, "Double EnumTestEnumVarnames = 2")
}

func TestExtPropGoTypeSkipOptionalPointer(t *testing.T) {
	packageName := "api"
	opts := Configuration{
		PackageName: packageName,
		Generate: GenerateOptions{
			EchoServer:   true,
			Models:       true,
			EmbeddedSpec: true,
			Strict:       true,
		},
	}
	spec := "test_specs/x-go-type-skip-optional-pointer.yaml"
	swagger, err := util.LoadSwagger(spec)
	require.NoError(t, err)

	// Run our code generation:
	code, err := Generate(swagger, opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, code)

	// Check that we have valid (formattable) code:
	_, err = format.Source([]byte(code))
	assert.NoError(t, err)

	// Check that optional pointer fields are skipped if requested
	assert.Contains(t, code, "NullableFieldSkipFalse *string `json:\"nullableFieldSkipFalse,omitempty\"`")
	assert.Contains(t, code, "NullableFieldSkipTrue  string  `json:\"nullableFieldSkipTrue,omitempty\"`")
	assert.Contains(t, code, "OptionalField          *string `json:\"optionalField,omitempty\"`")
	assert.Contains(t, code, "OptionalFieldSkipFalse *string `json:\"optionalFieldSkipFalse,omitempty\"`")
	assert.Contains(t, code, "OptionalFieldSkipTrue  string  `json:\"optionalFieldSkipTrue,omitempty\"`")

	// Check that the extension applies on custom types as well
	assert.Contains(t, code, "CustomTypeWithSkipTrue string  `json:\"customTypeWithSkipTrue,omitempty\"`")

	// Check that the extension has no effect on required fields
	assert.Contains(t, code, "RequiredField          string  `json:\"requiredField\"`")
}

func TestGoTypeImport(t *testing.T) {
	packageName := "api"
	opts := Configuration{
		PackageName: packageName,
		Generate: GenerateOptions{
			EchoServer:   true,
			Models:       true,
			EmbeddedSpec: true,
		},
	}
	spec := "test_specs/x-go-type-import-pet.yaml"
	swagger, err := util.LoadSwagger(spec)
	require.NoError(t, err)

	// Run our code generation:
	code, err := Generate(swagger, opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, code)

	// Check that we have valid (formattable) code:
	_, err = format.Source([]byte(code))
	assert.NoError(t, err)

	imports := []string{
		`github.com/CavernaTechnologies/pgext`, // schemas - direct object
		`myuuid "github.com/google/uuid"`,      // schemas - object
		`github.com/lib/pq`,                    // schemas - array
		`github.com/spf13/viper`,               // responses - direct object
		`golang.org/x/text`,                    // responses - complex object
		`golang.org/x/email`,                   // requestBodies - in components
		`github.com/fatih/color`,               // parameters - query
		`github.com/go-openapi/swag`,           // parameters - path
		`github.com/jackc/pgtype`,              // direct parameters - path
		`github.com/mailru/easyjson`,           // direct parameters - query
		`github.com/subosito/gotenv`,           // direct request body
	}

	// Check import
	for _, imp := range imports {
		assert.Contains(t, code, imp)
	}
}

func TestGoAllofTypeOverride(t *testing.T) {
	packageName := "api"
	opts := Configuration{
		PackageName: packageName,
		Generate: GenerateOptions{
			EchoServer:   true,
			Models:       true,
			EmbeddedSpec: true,
		},
	}
	spec := "test_specs/x-go-type-pet-allof.yaml"
	swagger, err := util.LoadSwagger(spec)
	require.NoError(t, err)

	// Run our code generation:
	code, err := Generate(swagger, opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, code)

	// Check that we have valid (formattable) code:
	_, err = format.Source([]byte(code))
	assert.NoError(t, err)

	for _, expected := range []string{
		"type Cat = cat.Cat",
		"type Dog = dog.Dog",
		"type Pet = pet.Pet",
		"github.com/somepetproject/pkg/cat",
		"github.com/somepetproject/pkg/dog",
		"github.com/somepetproject/pkg/pet",
	} {
		assert.Contains(t, code, expected)
	}
}

func TestRemoteExternalReference(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that interacts with the network")
	}

	packageName := "api"
	opts := Configuration{
		PackageName: packageName,
		Generate: GenerateOptions{
			Models: true,
		},
		ImportMapping: map[string]string{
			remoteRefFile: remoteRefImport,
		},
	}
	spec := "test_specs/remote-external-reference.yaml"
	swagger, err := util.LoadSwagger(spec)
	require.NoError(t, err)

	// Run our code generation:
	code, err := Generate(swagger, opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, code)

	// Check that we have valid (formattable) code:
	_, err = format.Source([]byte(code))
	assert.NoError(t, err)

	// Check that we have a package:
	assert.Contains(t, code, "package api")

	// Check import
	assert.Contains(t, code, `externalRef0 "github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded"`)

	// Check generated oneOf structure:
	assert.Contains(t, code, `
// ExampleSchema_Item defines model for ExampleSchema.Item.
type ExampleSchema_Item struct {
	union json.RawMessage
}
`)

	// Check generated oneOf structure As method:
	assert.Contains(t, code, `
// AsExternalRef0NewPet returns the union data inside the ExampleSchema_Item as a externalRef0.NewPet
func (t ExampleSchema_Item) AsExternalRef0NewPet() (externalRef0.NewPet, error) {
`)

	// Check generated oneOf structure From method:
	assert.Contains(t, code, `
// FromExternalRef0NewPet overwrites any union data inside the ExampleSchema_Item as the provided externalRef0.NewPet
func (t *ExampleSchema_Item) FromExternalRef0NewPet(v externalRef0.NewPet) error {
`)

	// Check generated oneOf structure Merge method:
	assert.Contains(t, code, `
// FromExternalRef0NewPet overwrites any union data inside the ExampleSchema_Item as the provided externalRef0.NewPet
func (t *ExampleSchema_Item) FromExternalRef0NewPet(v externalRef0.NewPet) error {
`)
}

func TestDuplicatePathParameter(t *testing.T) {
	// Regression test for https://github.com/oapi-codegen/oapi-codegen/issues/1574
	// Some real-world specs (e.g. Keycloak) have paths where the same parameter
	// appears more than once: /clients/{client-uuid}/.../clients/{client-uuid}
	spec := `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Duplicate path param test
paths:
  /admin/realms/{realm}/clients/{client-uuid}/roles/{role-name}/composites/clients/{client-uuid}:
    get:
      operationId: getCompositeRoles
      parameters:
        - name: realm
          in: path
          required: true
          schema:
            type: string
        - name: client-uuid
          in: path
          required: true
          schema:
            type: string
        - name: role-name
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Success
`
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	opts := Configuration{
		PackageName: "api",
		Generate: GenerateOptions{
			EchoServer: true,
			Client:     true,
			Models:     true,
		},
	}

	code, err := Generate(swagger, opts)
	require.NoError(t, err)
	assert.NotEmpty(t, code)

	// Verify the generated code is valid Go.
	_, err = format.Source([]byte(code))
	require.NoError(t, err)

	// The path params should appear exactly once in the function signature.
	assert.Contains(t, code, "realm string")
	assert.Contains(t, code, "clientUuid string")
	assert.Contains(t, code, "roleName string")
}

// TestEnumConflictDetectionOrderIndependent checks that conflict detection
// doesn't miss overlaps because an enum was already marked for prefixing.
func TestEnumConflictDetectionOrderIndependent(t *testing.T) {
	// AState+BState share "running" (both prefixed), AState+CState share "migrating".
	// The bug: once AState was marked, GetValues() returned prefixed names that
	// no longer matched CState's raw values, so CState's conflict was missed.
	const spec = `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Test Enum Conflict Detection
paths: {}
components:
  schemas:
    AState:
      type: string
      enum:
        - running
        - migrating
    BState:
      type: string
      enum:
        - running
    CState:
      type: string
      enum:
        - migrating
`
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	opts := Configuration{
		PackageName: "api",
		Generate: GenerateOptions{
			Models: true,
		},
		OutputOptions: OutputOptions{
			SkipPrune: true,
		},
	}

	code, err := Generate(swagger, opts)
	require.NoError(t, err)

	_, err = format.Source([]byte(code))
	require.NoError(t, err)

	// All three enums share values with at least one other enum; all must be prefixed.
	assert.Contains(t, code, "AStateRunning")
	assert.Contains(t, code, "AStateMigrating")
	assert.Contains(t, code, "BStateRunning")
	assert.Contains(t, code, "CStateMigrating")
}

// TestEnumConflictDetectionBothOrders verifies that enum conflict detection
// produces identical, fully-prefixed output regardless of the order the
// schemas appear in the spec.
func TestEnumConflictDetectionBothOrders(t *testing.T) {
	specAFirst := `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Test
paths: {}
components:
  schemas:
    AState:
      type: string
      enum: [running, migrating]
    BState:
      type: string
      enum: [running]
    CState:
      type: string
      enum: [migrating]
`
	specCFirst := `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Test
paths: {}
components:
  schemas:
    CState:
      type: string
      enum: [migrating]
    BState:
      type: string
      enum: [running]
    AState:
      type: string
      enum: [running, migrating]
`
	opts := Configuration{
		PackageName: "api",
		Generate:    GenerateOptions{Models: true},
		OutputOptions: OutputOptions{
			SkipPrune: true,
		},
	}

	loader := openapi3.NewLoader()

	swaggerA, err := loader.LoadFromData([]byte(specAFirst))
	require.NoError(t, err)
	codeA, err := Generate(swaggerA, opts)
	require.NoError(t, err)

	swaggerC, err := loader.LoadFromData([]byte(specCFirst))
	require.NoError(t, err)
	codeC, err := Generate(swaggerC, opts)
	require.NoError(t, err)

	// Both orderings must produce fully prefixed constants.
	for _, code := range []string{codeA, codeC} {
		assert.Contains(t, code, "AStateRunning")
		assert.Contains(t, code, "AStateMigrating")
		assert.Contains(t, code, "BStateRunning")
		assert.Contains(t, code, "CStateMigrating")
		// Unprefixed names must not appear as standalone constants.
		assert.NotContains(t, code, "\tRunning ")
		assert.NotContains(t, code, "\tMigrating ")
	}
}

func TestBodylessResponseWithDefaultCatchAll(t *testing.T) {
	// Regression test: when a spec declares a response without content (e.g. 204)
	// alongside a "default" error response with JSON content, the generated parser
	// must emit a case clause for the bodyless response that short-circuits before
	// the default catch-all ("&& true") attempts to unmarshal an empty body.
	opts := Configuration{
		PackageName: "api",
		Generate: GenerateOptions{
			Client: true,
			Models: true,
		},
	}

	swagger, err := util.LoadSwagger("test_specs/bodyless-response-default.yaml")
	require.NoError(t, err)

	code, err := Generate(swagger, opts)
	require.NoError(t, err)

	// Check that we have valid (formattable) code:
	_, err = format.Source([]byte(code))
	require.NoError(t, err)

	// The 204 case must appear before the default catch-all so it short-circuits.
	assert.Contains(t, code, "case rsp.StatusCode == 204:\n\t\tbreak // No content-type")

	// The default catch-all must still be present for non-matching status codes.
	assert.Contains(t, code, `strings.Contains(rsp.Header.Get("Content-Type"), "json") && true`)

	// For the range wildcard case (2XX without content alongside 200 with content),
	// the explicit 200 JSON case must appear before the 2XX break to avoid shadowing.
	assert.Contains(t, code, "case rsp.StatusCode/100 == 2:\n\t\tbreak // No content-type")

	// Verify ordering: 200 JSON case appears before the 2XX no-content case.
	json200Idx := indexOf(code, `rsp.StatusCode == 200`)
	range2XXIdx := indexOf(code, `rsp.StatusCode/100 == 2`)
	assert.Greater(t, range2XXIdx, json200Idx, "explicit 200 content case must appear before bodyless 2XX range")

	// Within ParseCreateWidgetResponse: the bodyless 2XX guard must not shadow
	// the explicit 200 XML case, and must itself short-circuit before the
	// strict (exact Content-Type match) default JSON catch-alls emitted when
	// the default response declares multiple JSON content types.
	widgetIdx := indexOf(code, "func ParseCreateWidgetResponse")
	require.NotEqual(t, -1, widgetIdx)
	widgetCode := code[widgetIdx:]
	xml200Idx := indexOf(widgetCode, `strings.Contains(rsp.Header.Get("Content-Type"), "xml") && rsp.StatusCode == 200`)
	widgetRangeIdx := indexOf(widgetCode, `rsp.StatusCode/100 == 2`)
	strictDefaultIdx := indexOf(widgetCode, `rsp.Header.Get("Content-Type") == "application/json" && true`)
	require.NotEqual(t, -1, xml200Idx)
	require.NotEqual(t, -1, widgetRangeIdx)
	require.NotEqual(t, -1, strictDefaultIdx)
	assert.Greater(t, widgetRangeIdx, xml200Idx, "explicit 200 XML case must appear before bodyless 2XX guard")
	assert.Greater(t, strictDefaultIdx, widgetRangeIdx, "bodyless 2XX guard must appear before strict default catch-all")
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// TestEnableAuthScopesOnContext verifies that generated server code embeds
// security scheme scopes into the request context only when the deprecated
// enable-auth-scopes-on-context compatibility option is set.
// Please see https://github.com/oapi-codegen/oapi-codegen/issues/1524
func TestEnableAuthScopesOnContext(t *testing.T) {
	const spec = `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Test Auth Scopes Emission
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
paths:
  /secured:
    get:
      operationId: secured
      security:
        - bearerAuth: ["read"]
      responses:
        '200':
          description: ok
`
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	opts := Configuration{
		PackageName: "api",
		Generate: GenerateOptions{
			StdHTTPServer: true,
			Models:        true,
		},
	}

	// By default, no context key types, scope constants, or context values
	// are generated.
	code, err := Generate(swagger, opts)
	require.NoError(t, err)
	assert.NotContains(t, code, "BearerAuthScopes")
	assert.NotContains(t, code, "bearerAuthContextKey")
	assert.NotContains(t, code, "context.WithValue")

	// With the compatibility option set, the legacy scope emission returns.
	opts.Compatibility.EnableAuthScopesOnContext = true
	code, err = Generate(swagger, opts)
	require.NoError(t, err)
	assert.Contains(t, code, `BearerAuthScopes bearerAuthContextKey = "bearerAuth.Scopes"`)
	assert.Contains(t, code, `ctx = context.WithValue(ctx, BearerAuthScopes, []string{"read"})`)
}

// TestEnumValueEscaping verifies that a string enum value containing a
// backslash is emitted as a valid, properly escaped Go string literal
// (the constants go through strconv.Quote).
// Please see https://github.com/oapi-codegen/oapi-codegen/issues/2180
func TestEnumValueEscaping(t *testing.T) {
	const spec = `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Enum escaping
paths: {}
components:
  schemas:
    AutoAccident:
      type: object
      properties:
        significant_injury:
          type: string
          enum: ["YES", "NO", "N\\A"]
`
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	opts := Configuration{
		PackageName:   "api",
		Generate:      GenerateOptions{Models: true},
		OutputOptions: OutputOptions{SkipPrune: true},
	}

	code, err := Generate(swagger, opts)
	require.NoError(t, err)

	// The spec-level value is the three characters N, \, A; the generated
	// constant must escape the backslash.
	assert.Contains(t, code, `= "N\\A"`)

	_, err = format.Source([]byte(code))
	require.NoError(t, err)
}

const securityScopesSharedSpec = `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Common
paths: {}
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
`

const securityScopesUserSpec = `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: User API
paths:
  /user:
    get:
      operationId: getUser
      security:
        - BearerAuth: ["read"]
        - LocalAuth: []
      responses:
        '200':
          description: ok
components:
  securitySchemes:
    BearerAuth:
      $ref: './common.yml#/components/securitySchemes/BearerAuth'
    LocalAuth:
      type: http
      scheme: basic
`

// loadSecurityScopesUserSpec writes the shared/user spec pair to disk and
// loads the user spec, resolving the cross-file security scheme $ref.
func loadSecurityScopesUserSpec(t *testing.T) *openapi3.T {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "common.yml"), []byte(securityScopesSharedSpec), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "user.yml"), []byte(securityScopesUserSpec), 0o600))

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadFromFile(filepath.Join(dir, "user.yml"))
	require.NoError(t, err)
	return swagger
}

// TestSecuritySchemeScopesWithImportMapping verifies that a security scheme
// $ref'd from an import-mapped spec aliases the scopes constant declared by
// the mapped package instead of declaring its own context key type, so
// shared middleware sees a single context key across generated packages.
// Please see https://github.com/oapi-codegen/oapi-codegen/issues/2383
func TestSecuritySchemeScopesWithImportMapping(t *testing.T) {
	swagger := loadSecurityScopesUserSpec(t)

	opts := Configuration{
		PackageName: "user",
		Generate: GenerateOptions{
			StdHTTPServer: true,
			Models:        true,
		},
		Compatibility: CompatibilityOptions{EnableAuthScopesOnContext: true},
		ImportMapping: map[string]string{"./common.yml": "example.com/common"},
	}

	code, err := Generate(swagger, opts)
	require.NoError(t, err)

	// The const block is column-aligned by gofmt, so collapse runs of
	// whitespace before matching the declarations.
	normalized := strings.Join(strings.Fields(code), " ")

	// The imported scheme aliases the shared constant — carrying the shared
	// package's context key type — and declares no local type; middleware
	// still references the constant by its local name.
	assert.Contains(t, normalized, "BearerAuthScopes = externalRef0.BearerAuthScopes")
	assert.NotContains(t, code, "bearerAuthContextKey")
	assert.Contains(t, code, `ctx = context.WithValue(ctx, BearerAuthScopes, []string{"read"})`)

	// The locally declared scheme keeps its own typed constant.
	assert.Contains(t, normalized, `LocalAuthScopes localAuthContextKey = "LocalAuth.Scopes"`)
}

// TestSecuritySchemeScopesWithoutImportMapping verifies that an external
// scheme $ref without an import-mapping entry keeps the historical behavior
// of declaring the context key type and constant locally.
func TestSecuritySchemeScopesWithoutImportMapping(t *testing.T) {
	swagger := loadSecurityScopesUserSpec(t)

	opts := Configuration{
		PackageName: "user",
		Generate: GenerateOptions{
			StdHTTPServer: true,
			Models:        true,
		},
		Compatibility: CompatibilityOptions{EnableAuthScopesOnContext: true},
	}

	code, err := Generate(swagger, opts)
	require.NoError(t, err)

	assert.Contains(t, code, `BearerAuthScopes bearerAuthContextKey = "BearerAuth.Scopes"`)
}

// TestSecuritySchemeScopesCurrentPackageMapping verifies that a scheme $ref'd
// from a spec mapped to the current package ("-") emits neither the type nor
// the constant: the sibling config generating that spec into the same package
// declares both.
func TestSecuritySchemeScopesCurrentPackageMapping(t *testing.T) {
	swagger := loadSecurityScopesUserSpec(t)

	opts := Configuration{
		PackageName: "user",
		Generate: GenerateOptions{
			StdHTTPServer: true,
			Models:        true,
		},
		Compatibility: CompatibilityOptions{EnableAuthScopesOnContext: true},
		ImportMapping: map[string]string{"./common.yml": "-"},
	}

	code, err := Generate(swagger, opts)
	require.NoError(t, err)

	// No local declarations for the shared scheme...
	assert.NotContains(t, code, "bearerAuthContextKey")
	assert.NotContains(t, code, `= "BearerAuth.Scopes"`)
	// ...but middleware references the sibling-declared constant.
	assert.Contains(t, code, `ctx = context.WithValue(ctx, BearerAuthScopes, []string{"read"})`)
}

// TestSecuritySchemeScopesWithoutOperations verifies that a spec holding only
// shared definitions (no paths) still exports the scopes constants, so that
// other packages can alias them via import-mapping.
func TestSecuritySchemeScopesWithoutOperations(t *testing.T) {
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData([]byte(securityScopesSharedSpec))
	require.NoError(t, err)

	opts := Configuration{
		PackageName: "common",
		Generate: GenerateOptions{
			Models: true,
		},
		Compatibility: CompatibilityOptions{EnableAuthScopesOnContext: true},
	}

	code, err := Generate(swagger, opts)
	require.NoError(t, err)

	assert.Contains(t, code, `BearerAuthScopes bearerAuthContextKey = "BearerAuth.Scopes"`)
}

// TestSortHandlerRegistrations verifies the sort-handler-registrations
// compatibility flag: by default handlers are registered in spec-declaration
// order (issue #1887), and setting the flag restores the historical
// lexicographic (by path) registration order.
func TestSortHandlerRegistrations(t *testing.T) {
	// Paths are declared in non-lexicographic order: zebra before apple.
	const spec = `
openapi: 3.0.0
info: { title: t, version: "1.0" }
paths:
  /zebra:
    get:
      operationId: getZebra
      responses: { '200': { description: ok } }
  /apple:
    get:
      operationId: getApple
      responses: { '200': { description: ok } }
`
	load := func() *openapi3.T {
		loader := openapi3.NewLoader()
		loader.IncludeOrigin = true // recover spec declaration order for SpecOrder
		swagger, err := loader.LoadFromData([]byte(spec))
		require.NoError(t, err)
		return swagger
	}

	base := Configuration{
		PackageName: "api",
		Generate:    GenerateOptions{FiberServer: true, Models: true},
	}

	regOrder := func(code string) (zebra, apple int) {
		return strings.Index(code, `options.BaseURL+"/zebra"`),
			strings.Index(code, `options.BaseURL+"/apple"`)
	}

	// Default: registration follows spec order — zebra before apple.
	code, err := Generate(load(), base)
	require.NoError(t, err)
	zebra, apple := regOrder(code)
	require.NotEqual(t, -1, zebra)
	require.NotEqual(t, -1, apple)
	assert.Less(t, zebra, apple, "default registration should follow spec order (zebra before apple)")

	// Flag set: registration restored to lexicographic order — apple before zebra.
	sorted := base
	sorted.Compatibility.SortHandlerRegistrations = true
	code, err = Generate(load(), sorted)
	require.NoError(t, err)
	zebra, apple = regOrder(code)
	require.NotEqual(t, -1, zebra)
	require.NotEqual(t, -1, apple)
	assert.Less(t, apple, zebra, "sort-handler-registrations should restore lexicographic order (apple before zebra)")
}

//go:embed test_spec.yaml
var testOpenAPIDefinition string
