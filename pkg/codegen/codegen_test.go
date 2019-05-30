package codegen

import (
	"go/format"
	"testing"

	examplePetstore "github.com/deepmap/oapi-codegen/examples/petstore-expanded/api"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/golangci/lint-1"
	"github.com/stretchr/testify/assert"
)

func TestExamplePetStoreCodeGeneration(t *testing.T) {

	// Input vars for code generation:
	packageName := "api"
	opts := Options{
		GenerateClient: true,
		GenerateServer: true,
		GenerateTypes:  true,
		EmbedSpec:      true,
	}

	// Get a spec from the example PetStore definition:
	swagger, err := examplePetstore.GetSwagger()
	assert.NoError(t, err)

	// Run our code generation:
	code, err := Generate(swagger, packageName, opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, code)

	// Check that we have valid (formattable) code:
	_, err = format.Source([]byte(code))
	assert.NoError(t, err)

	// Check that we have a package:
	assert.Contains(t, code, "package api")

	// Check that the client method signatures return response structs:
	assert.Contains(t, code, "func (c *Client) FindPetById(ctx context.Context, id int64) (*http.Response, error) {")

	// Make sure the generated code is valid:
	linter := new(lint.Linter)
	problems, err := linter.Lint("test.gen.go", []byte(code))
	assert.NoError(t, err)
	assert.Len(t, problems, 0)
}

func TestExampleOpenAPICodeGeneration(t *testing.T) {

	// Input vars for code generation:
	packageName := "testswagger"
	opts := Options{
		GenerateClientWithResponses: true,
		GenerateServer:              true,
		GenerateTypes:               true,
		EmbedSpec:                   true,
	}

	// Get a spec from the test definition in this file:
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(testOpenAPIDefinition))
	assert.NoError(t, err)

	// Run our code generation:
	code, err := Generate(swagger, packageName, opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, code)

	// Check that we have valid (formattable) code:
	_, err = format.Source([]byte(code))
	assert.NoError(t, err)

	// Check that we have a package:
	assert.Contains(t, code, "package testswagger")

	// Check that response structs are generated:
	assert.Contains(t, code, "type GetTestByNameResponse struct {")

	// Check that the client method signatures return response structs:
	assert.Contains(t, code, "func (c *Client) GetTestByName(ctx context.Context, name string) (*GetTestByNameResponse, error) {")

	// Make sure the generated code is valid:
	linter := new(lint.Linter)
	problems, err := linter.Lint("test.gen.go", []byte(code))
	assert.NoError(t, err)
	assert.Len(t, problems, 0)
}

const testOpenAPIDefinition = `
openapi: 3.0.1

info:
  title: OpenAPI-CodeGen Test
  description: 'This is a test OpenAPI Spec'
  version: 1.0.0

servers:
- url: https://test.oapi-codegen.com/v2
- url: http://test.oapi-codegen.com/v2

paths:
  /test/{name}:
    get:
      tags:
      - pet
      summary: Get test
      operationId: getTestByName
      parameters:
      - name: name
        in: path
        required: true
        schema:
          type: string
      responses:
        200:
          description: Success
          content:
            application/xml:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Test'
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Test'
        default:
          description: Error
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Error'
      security:
      - petstore_auth:
        - write:pets
        - read:pets



components:
  schemas:
  
    Test:
      properties:
        name:
          type: string
        cases:
          type: array
          items:
            $ref: '#/components/schemas/TestCase'

    TestCase:
      properties:
        name:
          type: string
        command:
          type: string
  
    Error:
      properties:
        code:
          type: integer
          format: int32
        message:
          type: string
`
