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

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.Equal(t, disc, result.Discriminator)
	})

	t.Run("allOf with single discriminator on s2 propagates it", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{Discriminator: disc}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.Equal(t, disc, result.Discriminator)
	})

	t.Run("allOf with discriminators on both schemas errors", func(t *testing.T) {
		disc2 := &openapi3.Discriminator{PropertyName: "kind"}
		s1 := openapi3.Schema{Discriminator: disc}
		s2 := openapi3.Schema{Discriminator: disc2}

		_, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "discriminators")
	})

	t.Run("allOf with no discriminators succeeds with nil discriminator", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.Nil(t, result.Discriminator)
	})

	t.Run("non-allOf with discriminator on s1 errors", func(t *testing.T) {
		s1 := openapi3.Schema{Discriminator: disc}
		s2 := openapi3.Schema{}

		_, err := mergeOpenapiSchemas(s1, s2, false, make(map[string]bool))
		require.Error(t, err)
	})

	t.Run("non-allOf with discriminator on s2 errors", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{Discriminator: disc}

		_, err := mergeOpenapiSchemas(s1, s2, false, make(map[string]bool))
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

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.Equal(t, stringType, result.Type)
	})

	t.Run("type on s1 only propagates", func(t *testing.T) {
		s1 := openapi3.Schema{Type: stringType}
		s2 := openapi3.Schema{}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool))
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

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.Equal(t, unionType, result.Type)
	})

	t.Run("equal types on both members merge", func(t *testing.T) {
		s1 := openapi3.Schema{Type: stringType}
		s2 := openapi3.Schema{Type: stringType}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.Equal(t, stringType, result.Type)
	})

	t.Run("different types on both members error", func(t *testing.T) {
		s1 := openapi3.Schema{Type: stringType}
		s2 := openapi3.Schema{Type: numberType}

		_, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incompatible types")
	})

	t.Run("neither member typed stays typeless", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool))
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

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.Equal(t, "uuid", result.Format)
	})

	t.Run("format on s1 only propagates", func(t *testing.T) {
		s1 := openapi3.Schema{Format: "uuid"}
		s2 := openapi3.Schema{}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.Equal(t, "uuid", result.Format)
	})

	t.Run("equal formats merge", func(t *testing.T) {
		s1 := openapi3.Schema{Format: "uuid"}
		s2 := openapi3.Schema{Format: "uuid"}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.Equal(t, "uuid", result.Format)
	})

	t.Run("different formats error", func(t *testing.T) {
		s1 := openapi3.Schema{Format: "uuid"}
		s2 := openapi3.Schema{Format: "date-time"}

		_, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incompatible formats")
	})

	t.Run("nullable decorator over format-carrying scalar merges", func(t *testing.T) {
		s1 := openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "uuid"}
		s2 := openapi3.Schema{Nullable: true}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool))
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

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.True(t, result.Nullable)
	})

	t.Run("nullable on s1 propagates to result", func(t *testing.T) {
		s1 := openapi3.Schema{Nullable: true}
		s2 := openapi3.Schema{}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.True(t, result.Nullable)
	})

	t.Run("both nullable stays nullable", func(t *testing.T) {
		s1 := openapi3.Schema{Nullable: true}
		s2 := openapi3.Schema{Nullable: true}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.True(t, result.Nullable)
	})

	t.Run("neither nullable stays non-nullable", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{}

		result, err := mergeOpenapiSchemas(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.False(t, result.Nullable)
	})
}
