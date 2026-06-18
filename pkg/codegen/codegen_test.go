package codegen

import (
	_ "embed"
	"go/format"
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

func TestJSONEncodingOptions(t *testing.T) {
	// Spec with a oneOf type (triggers union.tmpl MarshalJSON), a type with
	// additionalProperties (triggers additional-properties.tmpl MarshalJSON),
	// and a path with a JSON request body and response (exercises client.tmpl
	// and strict-interface.tmpl).
	const spec = `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: JSON encoding test
components:
  schemas:
    Cat:
      type: object
      properties:
        name:
          type: string
    Dog:
      type: object
      properties:
        breed:
          type: string
    Pet:
      oneOf:
        - $ref: '#/components/schemas/Cat'
        - $ref: '#/components/schemas/Dog'
    Metadata:
      type: object
      properties:
        name:
          type: string
      additionalProperties:
        type: string
paths:
  /pets:
    post:
      operationId: createPet
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Pet'
      responses:
        '200':
          description: created pet
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Pet'
`
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	boolPtr := func(b bool) *bool { return &b }

	tests := []struct {
		name        string
		generate    GenerateOptions
		encoding    JSONEncodingOptions
		wantContain []string
		wantAbsent  []string
	}{
		// --- models (union / additionalProperties MarshalJSON) ---
		{
			name:        "models/default: uses json.Marshal",
			generate:    GenerateOptions{Models: true},
			encoding:    JSONEncodingOptions{},
			wantContain: []string{"json.Marshal(v)"},
			wantAbsent:  []string{"SetEscapeHTML", "SetIndent"},
		},
		{
			name:        "models/escape-html false: uses encoder with SetEscapeHTML(false)",
			generate:    GenerateOptions{Models: true},
			encoding:    JSONEncodingOptions{EscapeHTML: boolPtr(false)},
			wantContain: []string{"json.NewEncoder", "SetEscapeHTML(false)"},
			wantAbsent:  []string{"json.Marshal(v)", "SetIndent"},
		},
		{
			name:        "models/indent set: uses encoder with SetIndent",
			generate:    GenerateOptions{Models: true},
			encoding:    JSONEncodingOptions{Indent: "\t"},
			wantContain: []string{"json.NewEncoder", `SetIndent("", "\t")`},
			wantAbsent:  []string{"json.Marshal(v)", "SetEscapeHTML"},
		},
		{
			name:        "models/both options: encoder with SetEscapeHTML and SetIndent",
			generate:    GenerateOptions{Models: true},
			encoding:    JSONEncodingOptions{EscapeHTML: boolPtr(false), Indent: "  "},
			wantContain: []string{"json.NewEncoder", "SetEscapeHTML(false)", `SetIndent("", "  ")`},
			wantAbsent:  []string{"json.Marshal(v)"},
		},
		{
			name:        "models/escape-html true explicit: same as default",
			generate:    GenerateOptions{Models: true},
			encoding:    JSONEncodingOptions{EscapeHTML: boolPtr(true)},
			wantContain: []string{"json.Marshal(v)"},
			wantAbsent:  []string{"SetEscapeHTML", "SetIndent"},
		},
		// --- client (JSON request body path in client.tmpl) ---
		{
			name:        "client/default: uses json.Marshal(body)",
			generate:    GenerateOptions{Client: true},
			encoding:    JSONEncodingOptions{},
			wantContain: []string{"json.Marshal(body)"},
			wantAbsent:  []string{"func() ([]byte, error) {"},
		},
		{
			name:        "client/escape-html false: uses IIFE encoder for body",
			generate:    GenerateOptions{Client: true},
			encoding:    JSONEncodingOptions{EscapeHTML: boolPtr(false)},
			wantContain: []string{"func() ([]byte, error) {", "SetEscapeHTML(false)"},
			wantAbsent:  []string{"json.Marshal(body)"},
		},
		{
			name:        "client/indent set: uses IIFE encoder with SetIndent for body",
			generate:    GenerateOptions{Client: true},
			encoding:    JSONEncodingOptions{Indent: "\t"},
			wantContain: []string{"func() ([]byte, error) {", `SetIndent("", "\t")`},
			wantAbsent:  []string{"json.Marshal(body)", "SetEscapeHTML"},
		},
		// --- strict server (JSON response path in strict-interface.tmpl) ---
		{
			name:        "strict-server/default: uses json.NewEncoder(&buf) directly",
			generate:    GenerateOptions{StdHTTPServer: true, Strict: true, Models: true},
			encoding:    JSONEncodingOptions{},
			wantContain: []string{"json.NewEncoder(&buf)"},
			wantAbsent:  []string{"func() *json.Encoder {"},
		},
		{
			name:        "strict-server/escape-html false: uses IIFE encoder",
			generate:    GenerateOptions{StdHTTPServer: true, Strict: true, Models: true},
			encoding:    JSONEncodingOptions{EscapeHTML: boolPtr(false)},
			wantContain: []string{"func() *json.Encoder {", "SetEscapeHTML(false)"},
			wantAbsent:  []string{},
		},
		{
			name:        "strict-server/indent set: uses IIFE encoder with SetIndent",
			generate:    GenerateOptions{StdHTTPServer: true, Strict: true, Models: true},
			encoding:    JSONEncodingOptions{Indent: "  "},
			wantContain: []string{"func() *json.Encoder {", `SetIndent("", "  ")`},
			wantAbsent:  []string{"SetEscapeHTML"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Configuration{
				PackageName: "testpkg",
				Generate:    tt.generate,
				OutputOptions: OutputOptions{
					SkipPrune:    true,
					JSONEncoding: tt.encoding,
				},
			}
			code, err := Generate(swagger, opts)
			require.NoError(t, err)
			assert.NotEmpty(t, code)

			_, err = format.Source([]byte(code))
			require.NoError(t, err, "generated code must be valid Go")

			for _, want := range tt.wantContain {
				assert.Contains(t, code, want)
			}
			for _, absent := range tt.wantAbsent {
				assert.NotContains(t, code, absent)
			}
		})
	}
}

//go:embed test_spec.yaml
var testOpenAPIDefinition string
