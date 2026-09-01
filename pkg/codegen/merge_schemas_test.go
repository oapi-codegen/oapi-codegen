package codegen

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeOpenapiSchemas_DiscriminatorPropagation(t *testing.T) {
	disc := &openapi3.Discriminator{
		PropertyName: "type",
	}

	t.Run("allOf with single discriminator on s1 propagates it", func(t *testing.T) {
		s1 := openapi3.Schema{Discriminator: disc}
		s2 := openapi3.Schema{}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool), nil)
		require.NoError(t, err)
		assert.Equal(t, disc, result.Discriminator)
	})

	t.Run("allOf with single discriminator on s2 propagates it", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{Discriminator: disc}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool), nil)
		require.NoError(t, err)
		assert.Equal(t, disc, result.Discriminator)
	})

	t.Run("allOf with discriminators on both schemas errors", func(t *testing.T) {
		disc2 := &openapi3.Discriminator{PropertyName: "kind"}
		s1 := openapi3.Schema{Discriminator: disc}
		s2 := openapi3.Schema{Discriminator: disc2}

		_, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "discriminators")
	})

	t.Run("allOf with no discriminators succeeds with nil discriminator", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool), nil)
		require.NoError(t, err)
		assert.Nil(t, result.Discriminator)
	})

	t.Run("non-allOf with discriminator on s1 errors", func(t *testing.T) {
		s1 := openapi3.Schema{Discriminator: disc}
		s2 := openapi3.Schema{}

		_, err := mergeOpenapiSchemas(s1, s2, false, make(map[string]bool), nil)
		require.Error(t, err)
	})

	t.Run("non-allOf with discriminator on s2 errors", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{Discriminator: disc}

		_, err := mergeOpenapiSchemas(s1, s2, false, make(map[string]bool), nil)
		require.Error(t, err)
	})
}

// TestMergeOpenapiSchemas_TypePropagation covers the one-sided type rule
// (issue #2524): a type declared by only one allOf member propagates to the
// merged result regardless of member order, while two members declaring
// different types remain an error.
func TestMergeOpenapiSchemas_TypePropagation(t *testing.T) {
	stringType := &openapi3.Types{"string"}
	numberType := &openapi3.Types{"number"}
	unionType := &openapi3.Types{"string", "number"}

	t.Run("type on s2 only propagates", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{Type: stringType}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool), nil)
		require.NoError(t, err)
		assert.Equal(t, stringType, result.Type)
	})

	t.Run("type on s1 only propagates", func(t *testing.T) {
		s1 := openapi3.Schema{Type: stringType}
		s2 := openapi3.Schema{}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool), nil)
		require.NoError(t, err)
		assert.Equal(t, stringType, result.Type)
	})

	t.Run("multi-type union on s2 only propagates", func(t *testing.T) {
		s1 := openapi3.Schema{
			Properties: openapi3.Schemas{
				"name": openapi3.NewSchemaRef("", openapi3.NewStringSchema()),
			},
		}
		s2 := openapi3.Schema{Type: unionType}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool), nil)
		require.NoError(t, err)
		assert.Equal(t, unionType, result.Type)
	})

	t.Run("equal types on both members merge", func(t *testing.T) {
		s1 := openapi3.Schema{Type: stringType}
		s2 := openapi3.Schema{Type: stringType}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool), nil)
		require.NoError(t, err)
		assert.Equal(t, stringType, result.Type)
	})

	t.Run("different types on both members error", func(t *testing.T) {
		s1 := openapi3.Schema{Type: stringType}
		s2 := openapi3.Schema{Type: numberType}

		_, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incompatible types")
	})

	t.Run("neither member typed stays typeless", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool), nil)
		require.NoError(t, err)
		assert.Nil(t, result.Type.Slice())
	})
}

// TestMergeOpenapiSchemas_FormatPropagation covers the same one-sided rule
// for format: a format declared by only one member propagates instead of
// erroring, which is what the allOf decorator idiom over format-carrying
// scalars produces (e.g. $ref to {type: string, format: uuid} + nullable).
func TestMergeOpenapiSchemas_FormatPropagation(t *testing.T) {
	t.Run("format on s2 only propagates", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{Format: "uuid"}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool), nil)
		require.NoError(t, err)
		assert.Equal(t, "uuid", result.Format)
	})

	t.Run("format on s1 only propagates", func(t *testing.T) {
		s1 := openapi3.Schema{Format: "uuid"}
		s2 := openapi3.Schema{}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool), nil)
		require.NoError(t, err)
		assert.Equal(t, "uuid", result.Format)
	})

	t.Run("equal formats merge", func(t *testing.T) {
		s1 := openapi3.Schema{Format: "uuid"}
		s2 := openapi3.Schema{Format: "uuid"}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool), nil)
		require.NoError(t, err)
		assert.Equal(t, "uuid", result.Format)
	})

	t.Run("different formats error", func(t *testing.T) {
		s1 := openapi3.Schema{Format: "uuid"}
		s2 := openapi3.Schema{Format: "date-time"}

		_, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incompatible formats")
	})

	t.Run("nullable decorator over format-carrying scalar merges", func(t *testing.T) {
		s1 := openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "uuid"}
		s2 := openapi3.Schema{Nullable: true}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool), nil)
		require.NoError(t, err)
		assert.Equal(t, "uuid", result.Format)
		assert.True(t, result.Nullable)
	})
}

// TestMergeOpenapiSchemas_NullableUnion covers the OpenAPI 3.0 idiom of
// decorating a $ref with `nullable: true` via allOf (issue #1898). Members
// disagreeing on nullability must merge (union) rather than error.
func TestMergeOpenapiSchemas_NullableUnion(t *testing.T) {
	t.Run("nullable on s2 propagates to result", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{Nullable: true}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool), nil)
		require.NoError(t, err)
		assert.True(t, result.Nullable)
	})

	t.Run("nullable on s1 propagates to result", func(t *testing.T) {
		s1 := openapi3.Schema{Nullable: true}
		s2 := openapi3.Schema{}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool), nil)
		require.NoError(t, err)
		assert.True(t, result.Nullable)
	})

	t.Run("both nullable stays nullable", func(t *testing.T) {
		s1 := openapi3.Schema{Nullable: true}
		s2 := openapi3.Schema{Nullable: true}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool), nil)
		require.NoError(t, err)
		assert.True(t, result.Nullable)
	})

	t.Run("neither nullable stays non-nullable", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool), nil)
		require.NoError(t, err)
		assert.False(t, result.Nullable)
	})
}

// TestIsSelfRef covers the back-reference detection that guards recursive
// allOf merging (issue #2542). Only a local component-schema ref whose target
// is the top-level schema currently being generated (path[0]), referenced
// from a nested position, counts as a self-reference.
func TestIsSelfRef(t *testing.T) {
	ref := "#/components/schemas/Node"

	t.Run("nested reference back to the schema being generated", func(t *testing.T) {
		assert.True(t, isSelfRef(ref, []string{"Node", "1", "children", "items"}))
	})

	t.Run("top-level fixed-point composition is not a nested self-reference", func(t *testing.T) {
		// RecursiveObject-style allOf: [$ref: A, $ref: A, {...}] at the top
		// level is handled by the seenSchemaRef cycle detection instead.
		assert.False(t, isSelfRef(ref, []string{"Node"}))
	})

	t.Run("reference to a different schema", func(t *testing.T) {
		assert.False(t, isSelfRef(ref, []string{"PRFile"}))
	})

	t.Run("external reference", func(t *testing.T) {
		assert.False(t, isSelfRef("./common/spec.yaml#/components/schemas/Node", []string{"Node", "items"}))
	})

	t.Run("inline member", func(t *testing.T) {
		assert.False(t, isSelfRef("", []string{"Node", "items"}))
	})
}

// TestBackRefSchema covers the substitute schema built for a
// self-referential allOf member: its only content is the $ref as a single
// anyOf branch, so the merged result references the named Go type instead of
// inlining the recursive schema's body (issue #2542).
func TestBackRefSchema(t *testing.T) {
	ref := &openapi3.SchemaRef{
		Ref: "#/components/schemas/Node",
		Value: &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			Properties: openapi3.Schemas{
				"leaf": openapi3.NewSchemaRef("", openapi3.NewStringSchema()),
			},
		},
	}

	s := backRefSchema(ref)
	require.Len(t, s.AnyOf, 1)
	assert.Equal(t, "#/components/schemas/Node", s.AnyOf[0].Ref)
	assert.Same(t, ref.Value, s.AnyOf[0].Value, "Value must be carried so GenerateGoSchema can dereference it safely")
	assert.Empty(t, s.Properties, "the recursive schema's body must not be inlined")
	assert.Nil(t, s.Type.Slice(), "the recursive schema's type must not be inlined")
	assert.Nil(t, s.Discriminator)
}

// TestBackRefSchema_Discriminator verifies that a discriminator declared by
// the self-referenced schema is carried onto the substitute so union codegen
// keeps mapping it.
func TestBackRefSchema_Discriminator(t *testing.T) {
	disc := &openapi3.Discriminator{PropertyName: "diff_type"}
	ref := &openapi3.SchemaRef{
		Ref:   "#/components/schemas/Node",
		Value: &openapi3.Schema{Discriminator: disc},
	}

	s := backRefSchema(ref)
	require.Len(t, s.AnyOf, 1)
	assert.Equal(t, "#/components/schemas/Node", s.AnyOf[0].Ref)
	assert.Equal(t, disc, s.Discriminator)
}

// generateSpec loads an inline OpenAPI spec and generates models from it.
func generateSpec(t *testing.T, spec string) string {
	t.Helper()
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)
	code, err := Generate(swagger, Configuration{
		PackageName: "repro",
		Generate: GenerateOptions{
			Models: true,
		},
		OutputOptions: OutputOptions{
			SkipPrune: true,
		},
	})
	require.NoError(t, err)
	return code
}

// TestMergeSchemasRecursiveAnyOfAllOf reproduces the self-referential
// anyOf+allOf schema from issue #2542 end-to-end. Generation must terminate,
// and — unlike v2.7.2, which silently dropped the recursive member — the
// item type must keep a reference back to Node as a union branch, so the
// recursive payload can still be represented, marshaled, and unmarshaled.
func TestMergeSchemasRecursiveAnyOfAllOf(t *testing.T) {
	const spec = `openapi: 3.0.0
info:
  title: repro
  version: "1.0.0"
paths: {}
components:
  schemas:
    Node:
      anyOf:
        - type: object
          properties:
            leaf:
              type: string
        - type: object
          properties:
            children:
              type: array
              items:
                allOf:
                  - $ref: '#/components/schemas/Node'
                  - type: object
                    properties:
                      extra:
                        type: string
`

	code := generateSpec(t, spec)
	assert.Contains(t, code, "type Node struct {")
	assert.Contains(t, code, "Extra *string")
	// The recursive reference must be preserved: the merged item type carries
	// the self-$ref as a union branch referencing the named Node type.
	assert.Contains(t, code, "type Node_1_Children_Item struct {")
	assert.Contains(t, code, "func (t Node_1_Children_Item) AsNode() (Node, error)")
	assert.Contains(t, code, "func (t *Node_1_Children_Item) FromNode(v Node) error")
}

// TestMergeSchemasRecursiveObjectAllOf reproduces the object-recursion
// variant of issue #2542: a plain object whose array items are
// allOf: [$ref: self, {extra}]. This shape also overflowed the stack, and the
// generated item must likewise keep the reference back to the parent type.
func TestMergeSchemasRecursiveObjectAllOf(t *testing.T) {
	const spec = `openapi: 3.0.0
info:
  title: repro
  version: "1.0.0"
paths: {}
components:
  schemas:
    Node:
      type: object
      properties:
        children:
          type: array
          items:
            allOf:
              - $ref: '#/components/schemas/Node'
              - type: object
                properties:
                  extra:
                    type: string
`

	code := generateSpec(t, spec)
	assert.Contains(t, code, "type Node struct {")
	assert.Contains(t, code, "type Node_Children_Item struct {")
	assert.Contains(t, code, "Extra *string")
	assert.Contains(t, code, "func (t Node_Children_Item) AsNode() (Node, error)")
	assert.Contains(t, code, "func (t *Node_Children_Item) FromNode(v Node) error")
}

// TestMergeSchemasNestedAllOfSelfRef reproduces the transitive variant of
// issue #2542 (review issue 2 on PR #2543): the self-$ref hides behind a
// nested allOf member, which previously escaped the cycle guard and hung
// generation indefinitely.
func TestMergeSchemasNestedAllOfSelfRef(t *testing.T) {
	const spec = `openapi: 3.0.0
info:
  title: repro
  version: "1.0.0"
paths: {}
components:
  schemas:
    Node:
      anyOf:
        - type: object
          properties:
            leaf:
              type: string
        - type: object
          properties:
            children:
              type: array
              items:
                allOf:
                  - allOf:
                      - $ref: '#/components/schemas/Node'
                  - type: object
                    properties:
                      extra:
                        type: string
`

	code := generateSpec(t, spec)
	assert.Contains(t, code, "type Node_1_Children_Item struct {")
	assert.Contains(t, code, "Extra *string")
	assert.Contains(t, code, "func (t Node_1_Children_Item) AsNode() (Node, error)")
	assert.Contains(t, code, "func (t *Node_1_Children_Item) FromNode(v Node) error")
}
