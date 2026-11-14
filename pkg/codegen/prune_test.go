package codegen

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestFindReferences(t *testing.T) {
	t.Run("unfiltered", func(t *testing.T) {
		swagger, err := openapi3.NewLoader().LoadFromData([]byte(pruneSpecTestFixture))
		assert.NoError(t, err)

		refs := findComponentRefs(swagger)
		assert.Len(t, refs, 14)
	})
	t.Run("only cat", func(t *testing.T) {
		swagger, err := openapi3.NewLoader().LoadFromData([]byte(pruneSpecTestFixture))
		assert.NoError(t, err)
		opts := Configuration{
			OutputOptions: OutputOptions{
				IncludeTags: []string{"cat"},
			},
		}

		filterOperationsByTag(swagger, opts)

		refs := findComponentRefs(swagger)
		assert.Len(t, refs, 7)
	})
	t.Run("only dog", func(t *testing.T) {
		swagger, err := openapi3.NewLoader().LoadFromData([]byte(pruneSpecTestFixture))
		assert.NoError(t, err)

		opts := Configuration{
			OutputOptions: OutputOptions{
				IncludeTags: []string{"dog"},
			},
		}

		filterOperationsByTag(swagger, opts)

		refs := findComponentRefs(swagger)
		assert.Len(t, refs, 7)
	})
}

func TestFindReferencesInOpenAPI31SchemaFields(t *testing.T) {
	const target = "#/components/schemas/Target"
	newRef := func() *openapi3.SchemaRef {
		return &openapi3.SchemaRef{Ref: target}
	}

	swagger := &openapi3.T{
		OpenAPI: "3.1.0",
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{
				"Container": {
					Value: &openapi3.Schema{
						PrefixItems:           openapi3.SchemaRefs{newRef()},
						Contains:              newRef(),
						PatternProperties:     openapi3.Schemas{"pattern": newRef()},
						DependentSchemas:      openapi3.Schemas{"dependency": newRef()},
						PropertyNames:         newRef(),
						UnevaluatedItems:      openapi3.BoolSchema{Schema: newRef()},
						UnevaluatedProperties: openapi3.BoolSchema{Schema: newRef()},
						If:                    newRef(),
						Then:                  newRef(),
						Else:                  newRef(),
						Defs:                  openapi3.Schemas{"definition": newRef()},
						ContentSchema:         newRef(),
					},
				},
			},
		},
	}

	refs := findComponentRefs(swagger)
	assert.Len(t, refs, 12)
	for _, ref := range refs {
		assert.Equal(t, target, ref)
	}
}

func TestPruningPreservesReferencesInOpenAPI31RefSiblings(t *testing.T) {
	swagger, err := openapi3.NewLoader().LoadFromData([]byte(`
openapi: 3.1.0
info:
  title: Schema ref siblings
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Container"
components:
  schemas:
    Base:
      type: object
      properties:
        parent:
          $ref: "#/components/schemas/Base"
    Target:
      type: string
    Container:
      $ref: "#/components/schemas/Base"
      propertyNames:
        $ref: "#/components/schemas/Target"
    Unused:
      type: string
`))
	assert.NoError(t, err)

	pruneUnusedComponents(swagger)

	assert.Contains(t, swagger.Components.Schemas, "Base")
	assert.Contains(t, swagger.Components.Schemas, "Container")
	assert.Contains(t, swagger.Components.Schemas, "Target")
	assert.NotContains(t, swagger.Components.Schemas, "Unused")
}

func TestFindReferencesInWebhooksAndContent(t *testing.T) {
	newRef := func(name string) *openapi3.SchemaRef {
		return &openapi3.SchemaRef{Ref: "#/components/schemas/" + name}
	}

	encodingHeader := &openapi3.HeaderRef{Value: &openapi3.Header{
		Parameter: openapi3.Parameter{
			Content: openapi3.Content{
				"application/json": {Schema: newRef("EncodingHeader")},
			},
		},
	}}
	parameter := &openapi3.ParameterRef{Value: &openapi3.Parameter{
		Content: openapi3.Content{
			"application/json": {
				Schema:     newRef("ParameterSchema"),
				ItemSchema: newRef("ParameterItemSchema"),
				Encoding: openapi3.Encodings{
					"value": {Headers: openapi3.Headers{"X-Metadata": encodingHeader}},
				},
			},
		},
	}}
	requestBody := &openapi3.RequestBodyRef{Value: &openapi3.RequestBody{
		Content: openapi3.Content{
			"application/json": {ItemSchema: newRef("RequestBodyItemSchema")},
		},
	}}
	response := &openapi3.ResponseRef{Value: &openapi3.Response{
		Content: openapi3.Content{
			"application/json": {ItemSchema: newRef("ResponseItemSchema")},
		},
	}}
	header := &openapi3.HeaderRef{Value: &openapi3.Header{
		Parameter: openapi3.Parameter{
			Content: openapi3.Content{
				"application/json": {Schema: newRef("HeaderSchema")},
			},
		},
	}}
	webhookResponse := &openapi3.ResponseRef{Value: &openapi3.Response{
		Content: openapi3.Content{
			"application/json": {Schema: newRef("WebhookSchema")},
		},
	}}

	swagger := &openapi3.T{
		OpenAPI: "3.2.0",
		Webhooks: map[string]*openapi3.PathItem{
			"event": {
				Post: &openapi3.Operation{
					Responses: openapi3.NewResponses(openapi3.WithStatus(200, webhookResponse)),
				},
			},
		},
		Components: &openapi3.Components{
			Parameters:    openapi3.ParametersMap{"Parameter": parameter},
			RequestBodies: openapi3.RequestBodies{"RequestBody": requestBody},
			Responses:     openapi3.ResponseBodies{"Response": response},
			Headers:       openapi3.Headers{"Header": header},
		},
	}

	assert.ElementsMatch(t, []string{
		"#/components/schemas/EncodingHeader",
		"#/components/schemas/HeaderSchema",
		"#/components/schemas/ParameterItemSchema",
		"#/components/schemas/ParameterSchema",
		"#/components/schemas/RequestBodyItemSchema",
		"#/components/schemas/ResponseItemSchema",
		"#/components/schemas/WebhookSchema",
	}, findComponentRefs(swagger))
}

func TestFilterOnlyCat(t *testing.T) {
	// Get a spec from the test definition in this file:
	swagger, err := openapi3.NewLoader().LoadFromData([]byte(pruneSpecTestFixture))
	assert.NoError(t, err)

	opts := Configuration{
		OutputOptions: OutputOptions{
			IncludeTags: []string{"cat"},
		},
	}

	refs := findComponentRefs(swagger)
	assert.Len(t, refs, 14)

	assert.Len(t, swagger.Components.Schemas, 5)

	filterOperationsByTag(swagger, opts)

	refs = findComponentRefs(swagger)
	assert.Len(t, refs, 7)

	assert.NotEmpty(t, swagger.Paths.Value("/cat"), "/cat path should still be in spec")
	assert.NotEmpty(t, swagger.Paths.Value("/cat").Get, "GET /cat operation should still be in spec")
	assert.Empty(t, swagger.Paths.Value("/dog").Get, "GET /dog should have been removed from spec")

	pruneUnusedComponents(swagger)

	assert.Len(t, swagger.Components.Schemas, 3)
}

func TestFilterOnlyDog(t *testing.T) {
	// Get a spec from the test definition in this file:
	swagger, err := openapi3.NewLoader().LoadFromData([]byte(pruneSpecTestFixture))
	assert.NoError(t, err)

	opts := Configuration{
		OutputOptions: OutputOptions{
			IncludeTags: []string{"dog"},
		},
	}

	refs := findComponentRefs(swagger)
	assert.Len(t, refs, 14)

	filterOperationsByTag(swagger, opts)

	refs = findComponentRefs(swagger)
	assert.Len(t, refs, 7)

	assert.Len(t, swagger.Components.Schemas, 5)

	assert.NotEmpty(t, swagger.Paths.Value("/dog"))
	assert.NotEmpty(t, swagger.Paths.Value("/dog").Get)
	assert.Empty(t, swagger.Paths.Value("/cat").Get)

	pruneUnusedComponents(swagger)

	assert.Len(t, swagger.Components.Schemas, 3)
}

func TestPruningUnusedComponents(t *testing.T) {
	// Get a spec from the test definition in this file:
	swagger, err := openapi3.NewLoader().LoadFromData([]byte(pruneComprehensiveTestFixture))
	assert.NoError(t, err)

	assert.Len(t, swagger.Components.Schemas, 8)
	assert.Len(t, swagger.Components.Parameters, 1)
	assert.Len(t, swagger.Components.SecuritySchemes, 2)
	assert.Len(t, swagger.Components.RequestBodies, 1)
	assert.Len(t, swagger.Components.Responses, 2)
	assert.Len(t, swagger.Components.Headers, 3)
	assert.Len(t, swagger.Components.Examples, 1)
	assert.Len(t, swagger.Components.Links, 1)
	assert.Len(t, swagger.Components.Callbacks, 1)

	pruneUnusedComponents(swagger)

	assert.Len(t, swagger.Components.Schemas, 0)
	assert.Len(t, swagger.Components.Parameters, 0)
	// securitySchemes are an exception. definitions in securitySchemes
	// are referenced directly by name. and not by $ref
	assert.Len(t, swagger.Components.SecuritySchemes, 2)
	assert.Len(t, swagger.Components.RequestBodies, 0)
	assert.Len(t, swagger.Components.Responses, 0)
	assert.Len(t, swagger.Components.Headers, 0)
	assert.Len(t, swagger.Components.Examples, 0)
	assert.Len(t, swagger.Components.Links, 0)
	assert.Len(t, swagger.Components.Callbacks, 0)
}

const pruneComprehensiveTestFixture = `
openapi: 3.0.1

info:
  title: OpenAPI-CodeGen Test
  description: 'This is a test OpenAPI Spec'
  version: 1.0.0

servers:
- url: https://test.oapi-codegen.com/v2
- url: http://test.oapi-codegen.com/v2

paths:
  /test:
    get:
      operationId: doesNothing
      summary: does nothing
      tags: [nothing]
      responses:
        default:
          description: returns nothing
          content:
            application/json:
              schema:
                type: object
components:
  schemas:
    Object1:
      type: object
      properties:
        object:
          $ref: "#/components/schemas/Object2"
    Object2:
      type: object
      properties:
        object:
          $ref: "#/components/schemas/Object3"
    Object3:
      type: object
      properties:
        object:
          $ref: "#/components/schemas/Object4"
    Object4:
      type: object
      properties:
        object:
          $ref: "#/components/schemas/Object5"
    Object5:
      type: object
      properties:
        object:
          $ref: "#/components/schemas/Object6"
    Object6:
      type: object
    Pet:
      type: object
      required:
        - id
        - name
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
        tag:
          type: string
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
  parameters:
    offsetParam:
      name: offset
      in: query
      description: Number of items to skip before returning the results.
      required: false
      schema:
        type: integer
        format: int32
        minimum: 0
        default: 0
  securitySchemes:
    BasicAuth:
      type: http
      scheme: basic
    BearerAuth:
      type: http
      scheme: bearer
  requestBodies:
    PetBody:
      description: A JSON object containing pet information
      required: true
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Pet'
  responses:
    NotFound:
      description: The specified resource was not found
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    Unauthorized:
      description: Unauthorized
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
  headers:
    X-RateLimit-Limit:
      schema:
        type: integer
      description: Request limit per hour.
    X-RateLimit-Remaining:
      schema:
        type: integer
      description: The number of requests left for the time window.
    X-RateLimit-Reset:
      schema:
        type: string
        format: date-time
      description: The UTC date/time at which the current rate limit window resets
  examples:
    objectExample:
      value:
        id: 1
        name: new object
      summary: A sample object
  links:
    GetUserByUserId:
      description: >
        The id value returned in the response can be used as
        the userId parameter in GET /users/{userId}.
      operationId: getUser
      parameters:
        userId: '$response.body#/id'
  callbacks:
    MyCallback:
      '{$request.body#/callbackUrl}':
        post:
          requestBody:
            required: true
            content:
              application/json:
                schema:
                  type: object
                  properties:
                    message:
                      type: string
                      example: Some event happened
                  required:
                    - message
          responses:
            '200':
              description: Your server returns this code if it accepts the callback
`

const pruneSpecTestFixture = `
openapi: 3.0.1

info:
  title: OpenAPI-CodeGen Test
  description: 'This is a test OpenAPI Spec'
  version: 1.0.0

servers:
- url: https://test.oapi-codegen.com/v2
- url: http://test.oapi-codegen.com/v2

paths:
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

  /dog:
    get:
      tags:
        - dog
      summary: Get dog status
      operationId: getDogStatus
      responses:
        200:
          description: Success
          content:
            application/json:
              schema:
                oneOf:
                  - $ref: '#/components/schemas/DogAlive'
                  - $ref: '#/components/schemas/DogDead'
            application/xml:
              schema:
                anyOf:
                  - $ref: '#/components/schemas/DogAlive'
                  - $ref: '#/components/schemas/DogDead'
            application/yaml:
              schema:
                allOf:
                  - $ref: '#/components/schemas/DogAlive'
                  - $ref: '#/components/schemas/DogDead'
        default:
          description: Error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

components:
  schemas:

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
        cause:
          type: string
          enum: [car, dog, oldage]

    DogAlive:
      properties:
        name:
          type: string
        alive_since:
          type: string
          format: date-time

    DogDead:
      properties:
        name:
          type: string
        dead_since:
          type: string
          format: date-time
        cause:
          type: string
          enum: [car, cat, oldage]

`
