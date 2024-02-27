package codegen

import (
	_ "embed"
	"go/format"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/lint"

	"github.com/deepmap/oapi-codegen/v2/pkg/util"
)

const (
	remoteRefFile = `https://raw.githubusercontent.com/deepmap/oapi-codegen/master/examples/petstore-expanded` +
		`/petstore-expanded.yaml`
	remoteRefImport = `github.com/deepmap/oapi-codegen/v2/examples/petstore-expanded`
)

func checkLint(t *testing.T, filename string, code []byte) {
	linter := new(lint.Linter)
	problems, err := linter.Lint(filename, code)
	assert.NoError(t, err)
	assert.Len(t, problems, 0)
}

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
	JSON200      *[]Test
	XML200       *[]Test
	JSON422      *[]interface{}
	XML422       *[]interface{}
	JSONDefault  *Error
}`)

	// Check that the helper methods are generated correctly:
	assert.Contains(t, code, "func (r GetTestByNameResponse) Status() string {")
	assert.Contains(t, code, "func (r GetTestByNameResponse) StatusCode() int {")
	assert.Contains(t, code, "func ParseGetTestByNameResponse(rsp *http.Response) (*GetTestByNameResponse, error) {")

	// Check the client method signatures:
	assert.Contains(t, code, "type GetTestByNameParams struct {")
	assert.Contains(t, code, "Top *int `form:\"$top,omitempty\" json:\"$top,omitempty\" query:\"$top,omitempty\"`")
	assert.Contains(t, code, "func (c *Client) GetTestByName(ctx context.Context, name string, params *GetTestByNameParams, reqEditors ...RequestEditorFn) (*http.Response, error) {")
	assert.Contains(t, code, "func (c *ClientWithResponses) GetTestByNameWithResponse(ctx context.Context, name string, params *GetTestByNameParams, reqEditors ...RequestEditorFn) (*GetTestByNameResponse, error) {")
	assert.Contains(t, code, "DeadSince *time.Time    `json:\"dead_since,omitempty\" tag1:\"value1\" tag2:\"value2\"`")
	assert.Contains(t, code, "type EnumTestNumerics int")
	assert.Contains(t, code, "N2 EnumTestNumerics = 2")
	assert.Contains(t, code, "type EnumTestEnumNames int")
	assert.Contains(t, code, "Two  EnumTestEnumNames = 2")
	assert.Contains(t, code, "Double EnumTestEnumVarnames = 2")

	// Make sure the generated code is valid:
	checkLint(t, "test.gen.go", []byte(code))
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
	assert.Contains(t, code, "NullableFieldSkipFalse *string `json:\"nullableFieldSkipFalse\"`")
	assert.Contains(t, code, "NullableFieldSkipTrue  string  `json:\"nullableFieldSkipTrue\"`")
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

	// Make sure the generated code is valid:
	checkLint(t, "test.gen.go", []byte(code))

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
	assert.Contains(t, code, `externalRef0 "github.com/deepmap/oapi-codegen/v2/examples/petstore-expanded"`)

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

	// Make sure the generated code is valid:
	checkLint(t, "test.gen.go", []byte(code))

}

//go:embed test_spec.yaml
var testOpenAPIDefinition string
