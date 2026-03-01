package codegen

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func schemaRefFromDoc(t *testing.T, docYAML string, schemaName string) (*openapi3.SchemaRef, string) {
	t.Helper()
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(docYAML))
	require.NoError(t, err, "failed to load OpenAPI document")
	ref, ok := doc.Components.Schemas[schemaName]
	require.Truef(t, ok, "schema %q not found in components", schemaName)
	return ref, doc.OpenAPI
}

func TestIsSchemaNullable_31_TypeOnlyNull(t *testing.T) {
	doc := `
openapi: 3.1.0
info: { title: t, version: v }
paths: {}
components:
  schemas:
    S:
      type: 'null'
`
	ref, ver := schemaRefFromDoc(t, doc, "S")
	assert.True(t, IsSchemaNullable(ver, ref))
}

func TestIsSchemaNullable_31_TypeOnlyString(t *testing.T) {
	doc := `
openapi: 3.1.0
info: { title: t, version: v }
paths: {}
components:
  schemas:
    S:
      type: 'string'
`
	ref, ver := schemaRefFromDoc(t, doc, "S")
	assert.False(t, IsSchemaNullable(ver, ref))
}

func TestIsSchemaNullable_31_TypeUnionIncludesNull(t *testing.T) {
	doc := `
openapi: 3.1.0
info: { title: t, version: v }
paths: {}
components:
  schemas:
    S:
      type: [string, 'null']
`
	ref, ver := schemaRefFromDoc(t, doc, "S")
	assert.True(t, IsSchemaNullable(ver, ref))
}

func TestIsSchemaNullable_31_NonNullable(t *testing.T) {
	doc := `
openapi: 3.1.0
info: { title: t, version: v }
paths: {}
components:
  schemas:
    S:
      type: [string, integer]
`
	ref, ver := schemaRefFromDoc(t, doc, "S")
	assert.False(t, IsSchemaNullable(ver, ref))
}

func TestIsSchemaNullable_31_OneOfWithNullOnly(t *testing.T) {
	doc := `
openapi: 3.1.0
info: { title: t, version: v }
paths: {}
components:
  schemas:
    S:
      oneOf:
        - { type: string }
        - { type: 'null' }
`
	ref, ver := schemaRefFromDoc(t, doc, "S")
	assert.True(t, IsSchemaNullable(ver, ref))
}

func TestIsSchemaNullable_31_OneOf_AllNullArms(t *testing.T) {
	doc := `
openapi: 3.1.0
info: { title: t, version: v }
paths: {}
components:
  schemas:
    S:
      oneOf:
        - { type: 'null' }
        - { type: 'null' }
`
	ref, ver := schemaRefFromDoc(t, doc, "S")
	// JSON Schema oneOf requires exactly one match; null matches both arms → not nullable via oneOf
	assert.False(t, IsSchemaNullable(ver, ref))
}

func TestIsSchemaNullable_31_OneOf_MixedMultipleNullAllowingArms(t *testing.T) {
	doc := `
openapi: 3.1.0
info: { title: t, version: v }
paths: {}
components:
  schemas:
    S:
      oneOf:
        - { type: ['string','null'] }
        - { type: 'null' }
        - { type: integer }
`
	ref, ver := schemaRefFromDoc(t, doc, "S")
	// Two subschemas explicitly allow null -> oneOf(null) must fail (violates exactly-one)
	assert.False(t, IsSchemaNullable(ver, ref))
}

func TestIsSchemaNullable_31_AnyOfWithNullOnly(t *testing.T) {
	doc := `
openapi: 3.1.0
info: { title: t, version: v }
paths: {}
components:
  schemas:
    S:
      anyOf:
        - { type: integer }
        - { type: 'null' }
`
	ref, ver := schemaRefFromDoc(t, doc, "S")
	assert.True(t, IsSchemaNullable(ver, ref))
}

func TestIsSchemaNullable_31_AnyOf_AllNullArms(t *testing.T) {
	doc := `
openapi: 3.1.0
info: { title: t, version: v }
paths: {}
components:
  schemas:
    S:
      anyOf:
        - { type: 'null' }
        - { type: 'null' }
`
	ref, ver := schemaRefFromDoc(t, doc, "S")
	assert.True(t, IsSchemaNullable(ver, ref))
}

func TestIsSchemaNullable_31_AllOf_OneNullArm(t *testing.T) {
	doc := `
openapi: 3.1.0
info: { title: t, version: v }
paths: {}
components:
  schemas:
    S:
      allOf:
        - { type: 'null' }
        - { type: string }
`
	ref, ver := schemaRefFromDoc(t, doc, "S")
	assert.False(t, IsSchemaNullable(ver, ref))
}

func TestIsSchemaNullable_31_AllOf_AllNullArms(t *testing.T) {
	doc := `
openapi: 3.1.0
info: { title: t, version: v }
paths: {}
components:
  schemas:
    S:
      allOf:
        - { type: 'null' }
        - { type: 'null' }
`
	ref, ver := schemaRefFromDoc(t, doc, "S")
	assert.True(t, IsSchemaNullable(ver, ref))
}
func TestIsSchemaNullable_31_BaseTypeBlocksNull_InAnyOf(t *testing.T) {
	doc := `
openapi: 3.1.0
info: { title: t, version: v }
paths: {}
components:
  schemas:
    S:
      type: string
      anyOf:
        - { type: 'null' }
        - { minLength: 1 }
`
	ref, ver := schemaRefFromDoc(t, doc, "S")
	assert.False(t, IsSchemaNullable(ver, ref))
}

func TestIsSchemaNullable_31_EnumAllowsNull_NoType(t *testing.T) {
	doc := `
openapi: 3.1.0
info: { title: t, version: v }
paths: {}
components:
  schemas:
    S:
      enum: [null, "x"]
`
	ref, ver := schemaRefFromDoc(t, doc, "S")
	assert.True(t, IsSchemaNullable(ver, ref))
}

func TestIsSchemaNullable_31_EnumAllowsNull_WithTypeString_Disallowed(t *testing.T) {
	doc := `
openapi: 3.1.0
info: { title: t, version: v }
paths: {}
components:
  schemas:
    S:
      type: string
      enum: [null, "x"]
`
	ref, ver := schemaRefFromDoc(t, doc, "S")
	assert.False(t, IsSchemaNullable(ver, ref))
}

func TestIsSchemaNullable_31_NotBlocksNull(t *testing.T) {
	doc := `
openapi: 3.1.0
info: { title: t, version: v }
paths: {}
components:
  schemas:
    S:
      type: ['string','null']
      not:
        type: 'null'
`
	ref, ver := schemaRefFromDoc(t, doc, "S")
	assert.False(t, IsSchemaNullable(ver, ref))
}

func TestIsSchemaNullable_31_EmptySchema_NotNullableByPolicy(t *testing.T) {
	doc := `
openapi: 3.1.0
info: { title: t, version: v }
paths: {}
components:
  schemas:
    S: {}
`
	ref, ver := schemaRefFromDoc(t, doc, "S")
	assert.False(t, IsSchemaNullable(ver, ref))
}

func TestIsSchemaNullable_30_FallbackTrue(t *testing.T) {
	doc := `
openapi: 3.0.3
info: { title: t, version: v }
paths: {}
components:
  schemas:
    S:
      type: string
      nullable: true
`
	ref, ver := schemaRefFromDoc(t, doc, "S")
	assert.True(t, IsSchemaNullable(ver, ref))
}

func TestIsSchemaNullable_30_FallbackFalse(t *testing.T) {
	doc := `
openapi: 3.0.3
info: { title: t, version: v }
paths: {}
components:
  schemas:
    S:
      type: string
`
	ref, ver := schemaRefFromDoc(t, doc, "S")
	assert.False(t, IsSchemaNullable(ver, ref))
}
