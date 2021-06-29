package codegen

import (
	"bytes"
	"go/format"
	"io/ioutil"
	"net/http"
	"testing"

	examplePetstoreClient "github.com/deepmap/oapi-codegen/examples/petstore-expanded"
	examplePetstore "github.com/deepmap/oapi-codegen/examples/petstore-expanded/echo/api"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/golangci/lint-1"
	"github.com/stretchr/testify/assert"
)

func TestExamplePetStoreCodeGeneration(t *testing.T) {

	// Input vars for code generation:
	packageName := "api"
	opts := Options{
		GenerateClient:     true,
		GenerateEchoServer: true,
		GenerateTypes:      true,
		EmbedSpec:          true,
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
	opts := Options{
		GenerateTypes: true,
		UserTemplates: userTemplates,
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

func TestExampleOpenAPICodeGeneration(t *testing.T) {

	// Input vars for code generation:
	packageName := "testswagger"
	opts := Options{
		GenerateClient:     true,
		GenerateEchoServer: true,
		GenerateTypes:      true,
		EmbedSpec:          true,
	}

	// Get a spec from the test definition in this file:
	swagger, err := openapi3.NewLoader().LoadFromData([]byte(testOpenAPIDefinition))
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
	assert.Contains(t, code, "Top *int `json:\"$top,omitempty\"`")
	assert.Contains(t, code, "func (c *Client) GetTestByName(ctx context.Context, name string, params *GetTestByNameParams, reqEditors ...RequestEditorFn) (*http.Response, error) {")
	assert.Contains(t, code, "func (c *ClientWithResponses) GetTestByNameWithResponse(ctx context.Context, name string, params *GetTestByNameParams, reqEditors ...RequestEditorFn) (*GetTestByNameResponse, error) {")
	assert.Contains(t, code, "DeadSince *time.Time    `json:\"dead_since,omitempty\" tag1:\"value1\" tag2:\"value2\"`")

	// Make sure the generated code is valid:
	linter := new(lint.Linter)
	problems, err := linter.Lint("test.gen.go", []byte(code))
	assert.NoError(t, err)
	assert.Len(t, problems, 0)
}

func TestGenerateWithOAuth2Specification(t *testing.T) {
	// Input vars for code generation:
	packageName := "testswagger"
	opts := Options{
		GenerateClient:     true,
		GenerateEchoServer: true,
		GenerateTypes:      true,
		EmbedSpec:          true,
	}

	// Get a spec from the test definition in this file:
	swagger, err := openapi3.NewLoader().LoadFromData([]byte(testOpenAPIDefinitionWithOAuth2))
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

	assert.Contains(t, code, "router.GET(baseURL+\"/pets\", wrapper.FindPets, sv([]string{\"read:pets\", \"pet:pets\"}))")
	assert.Contains(t, code, "router.GET(baseURL+\"/pets/:id\", wrapper.FindPetByID, sv([]string{}))")

}

const testOpenAPIDefinitionWithOAuth2 = `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Swagger Petstore
  description: A sample API that uses a petstore as an example to demonstrate features in the OpenAPI 3.0 specification
  termsOfService: http://swagger.io/terms/
  contact:
    name: Swagger API Team
    email: apiteam@swagger.io
    url: http://swagger.io
  license:
    name: Apache 2.0
    url: https://www.apache.org/licenses/LICENSE-2.0.html
servers:
  - url: http://petstore.swagger.io/api

paths:
  /pets:
    get:
      summary: Returns all pets
      description: |
        Returns all pets from the system that the user has access to
        Nam sed condimentum est. Maecenas tempor sagittis sapien, nec rhoncus sem sagittis sit amet. Aenean at gravida augue, ac iaculis sem. Curabitur odio lorem, ornare eget elementum nec, cursus id lectus. Duis mi turpis, pulvinar ac eros ac, tincidunt varius justo. In hac habitasse platea dictumst. Integer at adipiscing ante, a sagittis ligula. Aenean pharetra tempor ante molestie imperdiet. Vivamus id aliquam diam. Cras quis velit non tortor eleifend sagittis. Praesent at enim pharetra urna volutpat venenatis eget eget mauris. In eleifend fermentum facilisis. Praesent enim enim, gravida ac sodales sed, placerat id erat. Suspendisse lacus dolor, consectetur non augue vel, vehicula interdum libero. Morbi euismod sagittis libero sed lacinia.

        Sed tempus felis lobortis leo pulvinar rutrum. Nam mattis velit nisl, eu condimentum ligula luctus nec. Phasellus semper velit eget aliquet faucibus. In a mattis elit. Phasellus vel urna viverra, condimentum lorem id, rhoncus nibh. Ut pellentesque posuere elementum. Sed a varius odio. Morbi rhoncus ligula libero, vel eleifend nunc tristique vitae. Fusce et sem dui. Aenean nec scelerisque tortor. Fusce malesuada accumsan magna vel tempus. Quisque mollis felis eu dolor tristique, sit amet auctor felis gravida. Sed libero lorem, molestie sed nisl in, accumsan tempor nisi. Fusce sollicitudin massa ut lacinia mattis. Sed vel eleifend lorem. Pellentesque vitae felis pretium, pulvinar elit eu, euismod sapien.
      operationId: findPets
      security:
        - OAuth2:
            - read:pets
            - pet:pets
        - BasicAuth:

      parameters:
        - name: tags
          in: query
          description: tags to filter by
          required: false
          style: form
          schema:
            type: array
            items:
              type: string
        - name: limit
          in: query
          description: maximum number of results to return
          required: false
          schema:
            type: integer
            format: int32
      responses:
        '200':
          description: pet response
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Pet'
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
    post:
      summary: Creates a new pet
      description: Creates a new pet in the store. Duplicates are allowed
      operationId: addPet
      requestBody:
        description: Pet to add to the store
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/NewPet'
      responses:
        '200':
          description: pet response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Pet'
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
  /pets/{id}:
    get:
      summary: Returns a pet by ID
      description: Returns a pet based on a single ID
      operationId: findPetByID
      parameters:
        - name: id
          in: path
          description: ID of pet to fetch
          required: true
          schema:
            type: integer
            format: int64
      responses:
        '200':
          description: pet response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Pet'
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
    delete:
      summary: Deletes a pet by ID
      description: deletes a single pet based on the ID supplied
      operationId: deletePet
      parameters:
        - name: id
          in: path
          description: ID of pet to delete
          required: true
          schema:
            type: integer
            format: int64
      responses:
        '204':
          description: pet deleted
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  securitySchemes:
    OAuth2:
      type: oauth2
      description: This API protects it's endpoints using JWT tokens issued by the authorization server
      flows:
        clientCredentials:
          authorizationUrl: "https://foo.com/auth"
          tokenUrl: "https://foo.com/oauth/token"
          scopes:
            "read:pets": Read pets
            "pet:pets": Pet pets
    BasicAuth:
      type: http
      scheme: basic
      description: foo
      bearerFormat: JWT
      in: header

  schemas:
    Pet:
      allOf:
        - $ref: '#/components/schemas/NewPet'
        - required:
          - id
          properties:
            id:
              type: integer
              format: int64
              description: Unique id of the pet

    NewPet:
      required:
        - name
      properties:
        name:
          type: string
          description: Name of the pet
        tag:
          type: string
          description: Type of the pet

    Error:
      required:
        - code
        - message
      properties:
        code:
          type: integer
          format: int32
          description: Error code
        message:
          type: string
          description: Error message

`
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
      - test
      summary: Get test
      operationId: getTestByName
      parameters:
      - name: name
        in: path
        required: true
        schema:
          type: string
      - name: $top
        in: query
        required: false
        schema:
          type: integer
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
        422:
          description: InvalidArray
          content:
            application/xml:
              schema:
                type: array
            application/json:
              schema:
                type: array
        default:
          description: Error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
  /cat:
    get:
      tags:
      - cat
      summary: Get cat status
      operationId: getCatStatus
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                oneOf:
                - $ref: '#/components/schemas/CatAlive'
                - $ref: '#/components/schemas/CatDead'
            application/xml:
              schema:
                anyOf:
                - $ref: '#/components/schemas/CatAlive'
                - $ref: '#/components/schemas/CatDead'
            application/yaml:
              schema:
                allOf:
                - $ref: '#/components/schemas/CatAlive'
                - $ref: '#/components/schemas/CatDead'
        default:
          description: Error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

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

    CatAlive:
      properties:
        name:
          type: string
        alive_since:
          type: string
          format: date-time

    CatDead:
      properties:
        name:
          type: string
        dead_since:
          type: string
          format: date-time
          x-oapi-codegen-extra-tags:
            tag1: value1
            tag2: value2
        cause:
          type: string
          enum: [car, dog, oldage]
`
