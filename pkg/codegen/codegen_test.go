package codegen

import (
	"bytes"
	_ "embed"
	"go/format"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/golangci/lint-1"
	"github.com/stretchr/testify/assert"

	examplePetstoreClient "github.com/deepmap/oapi-codegen/examples/petstore-expanded"
	examplePetstore "github.com/deepmap/oapi-codegen/examples/petstore-expanded/echo/api"
)

func TestExamplePetStoreCodeGeneration(t *testing.T) {

	// Input vars for code generation:
	packageName := "api"
	opts := Configuration{
		PackageName: packageName,
		Generate: GenerateOptions{
			EchoServer:   true,
			Client:       true,
			Models:       true,
			EmbeddedSpec: true,
		},
	}

	// Get a spec from the example PetStore definition:
	swagger, err := examplePetstore.GetSwagger()
	assert.NoError(t, err)

	// Run our code generation:
	code, err := Generate(swagger, opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, code)

	// Check that we have valid (formattable) code:
	_, err = format.Source([]byte(code))
	assert.NoError(t, err)

	// Check that we have a package:
	assert.Contains(t, code, "package api")

	// Check that the client method signatures return response structs:
	assert.Contains(t, code, "func (c *Client) FindPetByID(ctx context.Context, id int64, reqEditors ...RequestEditorFn) (*http.Response, error) {")

	// Check that the property comments were generated
	assert.Contains(t, code, "// Unique id of the pet")

	// Check that the summary comment contains newlines
	assert.Contains(t, code, `// Deletes a pet by ID
	// (DELETE /pets/{id})
`)

	// Make sure the generated code is valid:
	linter := new(lint.Linter)
	problems, err := linter.Lint("test.gen.go", []byte(code))
	assert.NoError(t, err)
	assert.Len(t, problems, 0)
}

func TestExamplePetStoreCodeGenerationWithUserTemplates(t *testing.T) {

	userTemplates := map[string]string{"typedef.tmpl": "//blah"}

	// Input vars for code generation:
	packageName := "api"
	opts := Configuration{
		PackageName: packageName,
		Generate: GenerateOptions{
			Models: true,
		},
		OutputOptions: OutputOptions{
			UserTemplates: userTemplates,
		},
	}

	// Get a spec from the example PetStore definition:
	swagger, err := examplePetstore.GetSwagger()
	assert.NoError(t, err)

	// Run our code generation:
	code, err := Generate(swagger, opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, code)

	// Check that we have valid (formattable) code:
	_, err = format.Source([]byte(code))
	assert.NoError(t, err)

	// Check that we have a package:
	assert.Contains(t, code, "package api")

	// Check that the built-in template has been overriden
	assert.Contains(t, code, "//blah")
}

func TestExamplePetStoreParseFunction(t *testing.T) {

	bodyBytes := []byte(`{"id": 5, "name": "testpet", "tag": "cat"}`)

	cannedResponse := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(bodyBytes)),
		Header:     http.Header{},
	}
	cannedResponse.Header.Add("Content-type", "application/json")

	findPetByIDResponse, err := examplePetstoreClient.ParseFindPetByIDResponse(cannedResponse)
	assert.NoError(t, err)
	assert.NotNil(t, findPetByIDResponse.JSON200)
	assert.Equal(t, int64(5), findPetByIDResponse.JSON200.Id)
	assert.Equal(t, "testpet", findPetByIDResponse.JSON200.Name)
	assert.NotNil(t, findPetByIDResponse.JSON200.Tag)
	assert.Equal(t, "cat", *findPetByIDResponse.JSON200.Tag)
}

var codegenTestCases = []struct {
	label       string
	typeMapping map[string]string
	contains    []string
}{
	{
		label: "plain",
		contains: []string{
			// Check that we have a package:
			"package testswagger",

			// Check that response structs are generated correctly:
			"type GetTestByNameResponse struct {",

			// Check that response structs contains fallbacks to interface for invalid types:
			// Here an invalid array with no items.
			`
type GetTestByNameResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]Test
	XML200       *[]Test
	JSON422      *[]interface{}
	XML422       *[]interface{}
	JSONDefault  *Error
}`,

			// Check that the helper methods are generated correctly:
			"func (r GetTestByNameResponse) Status() string {",
			"func (r GetTestByNameResponse) StatusCode() int {",
			"func ParseGetTestByNameResponse(rsp *http.Response) (*GetTestByNameResponse, error) {",

			// Check the client method signatures:
			"type GetTestByNameParams struct {",
			"Top *int `form:\"$top,omitempty\" json:\"$top,omitempty\"`",
			"func (c *Client) GetTestByName(ctx context.Context, name string, params *GetTestByNameParams, reqEditors ...RequestEditorFn) (*http.Response, error) {",
			"func (c *ClientWithResponses) GetTestByNameWithResponse(ctx context.Context, name string, params *GetTestByNameParams, reqEditors ...RequestEditorFn) (*GetTestByNameResponse, error) {",
			"DeadSince *time.Time    `json:\"dead_since,omitempty\" tag1:\"value1\" tag2:\"value2\"`",
		},
	},
	{
		label: "map_time_to_string",
		typeMapping: map[string]string{
			"string:date-time": "string",
		},
		contains: []string{
			"DeadSince  *string    `json:\"dead_since,omitempty\" tag1:\"value1\" tag2:\"value2\"`",
		},
	},
}

func TestExampleOpenAPICodeGeneration(t *testing.T) {
	for _, tc := range codegenTestCases {
		t.Run(tc.label, func(t *testing.T) {
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
				TypeMapping: tc.typeMapping,
			}

			// Get a spec from the test definition in this file:
			swagger, err := openapi3.NewLoader().LoadFromData([]byte(testOpenAPIDefinition))
			assert.NoError(t, err)

			// Run our code generation:
			code, err := Generate(swagger, opts)
			assert.NoError(t, err)
			assert.NotEmpty(t, code)

			// Check that we have valid (formattable) code:
			_, err = format.Source([]byte(code))
			assert.NoError(t, err)

			// Check that the code contains all required snippets
			for _, snippet := range tc.contains {
				assert.Contains(t, code, snippet)
			}

			// Make sure the generated code is valid:
			linter := new(lint.Linter)
			problems, err := linter.Lint("test.gen.go", []byte(code))
			assert.NoError(t, err)
			assert.Len(t, problems, 0)
		})
	}
}

//go:embed test_spec.yaml
var testOpenAPIDefinition string
