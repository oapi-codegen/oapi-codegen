package codegen

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestPruningDiscriminatorMappings(t *testing.T) {
	spec := `
openapi: 3.0.1
info:
  version: 1.0.0
  title: Test Discriminator Mappings Pruning
paths:
  /test:
    get:
      operationId: getTest
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/BaseType'
components:
  schemas:
    BaseType:
      type: object
      required:
        - type
      properties:
        type:
          type: string
      discriminator:
        propertyName: type
        mapping:
          typeA: '#/components/schemas/TypeA'
          typeB: '#/components/schemas/TypeB'
    TypeA:
      allOf:
        - $ref: '#/components/schemas/BaseType'
        - type: object
          properties:
            fieldA:
              type: string
    TypeB:
      allOf:
        - $ref: '#/components/schemas/BaseType'
        - type: object
          properties:
            fieldB:
              type: integer
    UnusedType:
      type: object
      properties:
        unused:
          type: string
`

	swagger, err := openapi3.NewLoader().LoadFromData([]byte(spec))
	assert.NoError(t, err)

	// Before pruning: should have all 4 schemas
	assert.Len(t, swagger.Components.Schemas, 4)
	assert.Contains(t, swagger.Components.Schemas, "BaseType")
	assert.Contains(t, swagger.Components.Schemas, "TypeA")
	assert.Contains(t, swagger.Components.Schemas, "TypeB")
	assert.Contains(t, swagger.Components.Schemas, "UnusedType")

	pruneUnusedComponents(swagger)

	// After pruning: should have 3 schemas (BaseType, TypeA, TypeB)
	// UnusedType should be removed, but TypeA and TypeB should be kept
	// because they're referenced in discriminator.mapping
	assert.Len(t, swagger.Components.Schemas, 3, "Should keep discriminator mapped types")
	assert.Contains(t, swagger.Components.Schemas, "BaseType", "BaseType should be kept")
	assert.Contains(t, swagger.Components.Schemas, "TypeA", "TypeA should be kept (discriminator mapping)")
	assert.Contains(t, swagger.Components.Schemas, "TypeB", "TypeB should be kept (discriminator mapping)")
	assert.NotContains(t, swagger.Components.Schemas, "UnusedType", "UnusedType should be removed")
}
