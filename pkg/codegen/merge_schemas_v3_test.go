package codegen

// These tests cover the allOf merge rules of schema-merging-behavior v3. They
// started as a copy of v2's (merge_schemas_v2_test.go).

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeOpenapiSchemas_DiscriminatorPropagationV3(t *testing.T) {
	disc := &openapi3.Discriminator{
		PropertyName: "type",
	}

	t.Run("allOf with single discriminator on s1 propagates it", func(t *testing.T) {
		s1 := openapi3.Schema{Discriminator: disc}
		s2 := openapi3.Schema{}

		result, err := mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.Equal(t, disc, result.Discriminator)
	})

	t.Run("allOf with single discriminator on s2 propagates it", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{Discriminator: disc}

		result, err := mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.Equal(t, disc, result.Discriminator)
	})

	t.Run("allOf with discriminators on both schemas errors", func(t *testing.T) {
		disc2 := &openapi3.Discriminator{PropertyName: "kind"}
		s1 := openapi3.Schema{Discriminator: disc}
		s2 := openapi3.Schema{Discriminator: disc2}

		_, err := mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "discriminators")
	})

	t.Run("allOf with no discriminators succeeds with nil discriminator", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{}

		result, err := mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.Nil(t, result.Discriminator)
	})

	t.Run("non-allOf with discriminator on s1 errors", func(t *testing.T) {
		s1 := openapi3.Schema{Discriminator: disc}
		s2 := openapi3.Schema{}

		_, err := mergeOpenapiSchemasV3(s1, s2, false, make(map[string]bool))
		require.Error(t, err)
	})

	t.Run("non-allOf with discriminator on s2 errors", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{Discriminator: disc}

		_, err := mergeOpenapiSchemasV3(s1, s2, false, make(map[string]bool))
		require.Error(t, err)
	})
}

// TestMergeOpenapiSchemas_TypePropagationV3 covers the one-sided type rule
// (issue #2524): a type declared by only one allOf member propagates to the
// merged result regardless of member order, while two members declaring
// different types remain an error.
func TestMergeOpenapiSchemas_TypePropagationV3(t *testing.T) {
	stringType := &openapi3.Types{"string"}
	numberType := &openapi3.Types{"number"}
	unionType := &openapi3.Types{"string", "number"}

	t.Run("type on s2 only propagates", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{Type: stringType}

		result, err := mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.Equal(t, stringType, result.Type)
	})

	t.Run("type on s1 only propagates", func(t *testing.T) {
		s1 := openapi3.Schema{Type: stringType}
		s2 := openapi3.Schema{}

		result, err := mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
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

		result, err := mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.Equal(t, unionType, result.Type)
	})

	t.Run("equal types on both members merge", func(t *testing.T) {
		s1 := openapi3.Schema{Type: stringType}
		s2 := openapi3.Schema{Type: stringType}

		result, err := mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.Equal(t, stringType, result.Type)
	})

	t.Run("different types on both members error", func(t *testing.T) {
		s1 := openapi3.Schema{Type: stringType}
		s2 := openapi3.Schema{Type: numberType}

		_, err := mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incompatible types")
	})

	t.Run("neither member typed stays typeless", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{}

		result, err := mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.Nil(t, result.Type.Slice())
	})
}

// TestMergeOpenapiSchemas_FormatPropagationV3 covers the same one-sided rule
// for format: a format declared by only one member propagates instead of
// erroring, which is what the allOf decorator idiom over format-carrying
// scalars produces (e.g. $ref to {type: string, format: uuid} + nullable).
func TestMergeOpenapiSchemas_FormatPropagationV3(t *testing.T) {
	t.Run("format on s2 only propagates", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{Format: "uuid"}

		result, err := mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.Equal(t, "uuid", result.Format)
	})

	t.Run("format on s1 only propagates", func(t *testing.T) {
		s1 := openapi3.Schema{Format: "uuid"}
		s2 := openapi3.Schema{}

		result, err := mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.Equal(t, "uuid", result.Format)
	})

	t.Run("equal formats merge", func(t *testing.T) {
		s1 := openapi3.Schema{Format: "uuid"}
		s2 := openapi3.Schema{Format: "uuid"}

		result, err := mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.Equal(t, "uuid", result.Format)
	})

	t.Run("different formats error", func(t *testing.T) {
		s1 := openapi3.Schema{Format: "uuid"}
		s2 := openapi3.Schema{Format: "date-time"}

		_, err := mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incompatible formats")
	})

	t.Run("nullable decorator over format-carrying scalar merges", func(t *testing.T) {
		s1 := openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "uuid"}
		s2 := openapi3.Schema{Nullable: true}

		result, err := mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.Equal(t, "uuid", result.Format)
		assert.True(t, result.Nullable)
	})
}

// TestMergeOpenapiSchemas_NullableUnionV3 covers the OpenAPI 3.0 idiom of
// decorating a $ref with `nullable: true` via allOf (issue #1898). Members
// disagreeing on nullability must merge (union) rather than error.
func TestMergeOpenapiSchemas_NullableUnionV3(t *testing.T) {
	t.Run("nullable on s2 propagates to result", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{Nullable: true}

		result, err := mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.True(t, result.Nullable)
	})

	t.Run("nullable on s1 propagates to result", func(t *testing.T) {
		s1 := openapi3.Schema{Nullable: true}
		s2 := openapi3.Schema{}

		result, err := mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.True(t, result.Nullable)
	})

	t.Run("both nullable stays nullable", func(t *testing.T) {
		s1 := openapi3.Schema{Nullable: true}
		s2 := openapi3.Schema{Nullable: true}

		result, err := mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.True(t, result.Nullable)
	})

	t.Run("neither nullable stays non-nullable", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{}

		result, err := mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		assert.False(t, result.Nullable)
	})
}

// TestMergeOpenapiSchemas_AnnotationsV3 covers keywords that don't shape the Go
// type. Merging them must never fail: the allOf decorator idiom sets them on
// one member only, and kin-openapi can't tell an unset flag from false.
func TestMergeOpenapiSchemas_AnnotationsV3(t *testing.T) {
	merge := func(t *testing.T, s1, s2 openapi3.Schema) openapi3.Schema {
		t.Helper()
		result, err := mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
		require.NoError(t, err)
		return result
	}

	t.Run("default on one member is kept (issue #1379)", func(t *testing.T) {
		assert.Equal(t, "asc", merge(t, openapi3.Schema{Default: "asc"}, openapi3.Schema{}).Default)
		assert.Equal(t, "asc", merge(t, openapi3.Schema{}, openapi3.Schema{Default: "asc"}).Default)
	})

	t.Run("defaults on both members keep the later one", func(t *testing.T) {
		assert.Equal(t, "desc", merge(t, openapi3.Schema{Default: "asc"}, openapi3.Schema{Default: "desc"}).Default)
	})

	t.Run("flags set on either member are set on the result", func(t *testing.T) {
		set := openapi3.Schema{UniqueItems: true, ReadOnly: true, WriteOnly: true, AllowEmptyValue: true}
		for name, pair := range map[string][2]openapi3.Schema{
			"s1": {set, {}},
			"s2": {{}, set},
		} {
			t.Run(name, func(t *testing.T) {
				result := merge(t, pair[0], pair[1])
				assert.True(t, result.UniqueItems, "uniqueItems")
				assert.True(t, result.ReadOnly, "readOnly")
				assert.True(t, result.WriteOnly, "writeOnly")
				assert.True(t, result.AllowEmptyValue, "allowEmptyValue")
			})
		}
	})

	t.Run("an exclusive bound on one member carries over", func(t *testing.T) {
		bound := openapi3.ExclusiveBound{Value: new(float64)}
		assert.Equal(t, bound, merge(t, openapi3.Schema{ExclusiveMin: bound}, openapi3.Schema{}).ExclusiveMin)
		assert.Equal(t, bound, merge(t, openapi3.Schema{}, openapi3.Schema{ExclusiveMax: bound}).ExclusiveMax)
	})
}

// TestMergeOpenapiSchemas_NullInTypeArrayV3 covers OpenAPI 3.1 type arrays,
// where "null" spells nullability. It is unioned like 3.0's `nullable`
// instead of taking part in the type comparison.
func TestMergeOpenapiSchemas_NullInTypeArrayV3(t *testing.T) {
	object := &openapi3.Types{"object"}
	nullableObject := &openapi3.Types{"object", "null"}

	merge := func(s1, s2 openapi3.Schema) (openapi3.Schema, error) {
		return mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
	}

	t.Run("a nullable member makes the result nullable", func(t *testing.T) {
		for name, pair := range map[string][2]*openapi3.Types{
			"s1 nullable": {nullableObject, object},
			"s2 nullable": {object, nullableObject},
		} {
			t.Run(name, func(t *testing.T) {
				result, err := merge(openapi3.Schema{Type: pair[0]}, openapi3.Schema{Type: pair[1]})
				require.NoError(t, err)
				assert.ElementsMatch(t, []string{"object", "null"}, result.Type.Slice())
			})
		}
	})

	t.Run("types that already agree are passed through unchanged", func(t *testing.T) {
		result, err := merge(openapi3.Schema{Type: object}, openapi3.Schema{Type: &openapi3.Types{"object"}})
		require.NoError(t, err)
		assert.Same(t, object, result.Type)
	})

	t.Run("type arrays compare as sets", func(t *testing.T) {
		_, err := merge(openapi3.Schema{Type: &openapi3.Types{"string", "integer"}},
			openapi3.Schema{Type: &openapi3.Types{"integer", "string"}})
		assert.NoError(t, err)
	})

	t.Run("different non-null types still conflict", func(t *testing.T) {
		_, err := merge(openapi3.Schema{Type: &openapi3.Types{"string"}},
			openapi3.Schema{Type: &openapi3.Types{"integer", "null"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incompatible types")
	})
}

// TestMergeOpenapiSchemas_AdditionalPropertiesV3 covers two members that both
// declare an additionalProperties schema.
func TestMergeOpenapiSchemas_AdditionalPropertiesV3(t *testing.T) {
	withAP := func(ap *openapi3.SchemaRef) openapi3.Schema {
		return openapi3.Schema{AdditionalProperties: openapi3.AdditionalProperties{Schema: ap}}
	}
	merge := func(s1, s2 openapi3.Schema) (openapi3.Schema, error) {
		return mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
	}

	t.Run("identical inline schemas merge", func(t *testing.T) {
		result, err := merge(
			withAP(openapi3.NewSchemaRef("", openapi3.NewStringSchema())),
			withAP(openapi3.NewSchemaRef("", openapi3.NewStringSchema())))
		require.NoError(t, err)
		assert.True(t, result.AdditionalProperties.Schema.Value.Type.Is("string"))
	})

	t.Run("the same $ref merges", func(t *testing.T) {
		result, err := merge(
			withAP(openapi3.NewSchemaRef("#/components/schemas/X", openapi3.NewStringSchema())),
			withAP(openapi3.NewSchemaRef("#/components/schemas/X", openapi3.NewStringSchema())))
		require.NoError(t, err)
		assert.Equal(t, "#/components/schemas/X", result.AdditionalProperties.Schema.Ref)
	})

	t.Run("different schemas still error", func(t *testing.T) {
		_, err := merge(
			withAP(openapi3.NewSchemaRef("", openapi3.NewStringSchema())),
			withAP(openapi3.NewSchemaRef("", openapi3.NewIntegerSchema())))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "additional properties")
	})
}

// TestMergeOpenapiSchemas_ItemsV3 covers array items, which used to be dropped
// by every allOf merge.
func TestMergeOpenapiSchemas_ItemsV3(t *testing.T) {
	merge := func(s1, s2 openapi3.Schema) (openapi3.Schema, error) {
		return mergeOpenapiSchemasV3(s1, s2, true, make(map[string]bool))
	}
	itemRef := openapi3.NewSchemaRef("#/components/schemas/Item", openapi3.NewObjectSchema())

	t.Run("items on one member carry over", func(t *testing.T) {
		for name, pair := range map[string][2]openapi3.Schema{
			"s1": {{Items: itemRef}, {}},
			"s2": {{}, {Items: itemRef}},
		} {
			t.Run(name, func(t *testing.T) {
				result, err := merge(pair[0], pair[1])
				require.NoError(t, err)
				assert.Same(t, itemRef, result.Items)
			})
		}
	})

	t.Run("the same $ref is kept", func(t *testing.T) {
		result, err := merge(openapi3.Schema{Items: itemRef},
			openapi3.Schema{Items: openapi3.NewSchemaRef("#/components/schemas/Item", openapi3.NewObjectSchema())})
		require.NoError(t, err)
		assert.Same(t, itemRef, result.Items)
	})

	t.Run("different item schemas merge", func(t *testing.T) {
		a := openapi3.NewObjectSchema().WithProperty("a", openapi3.NewStringSchema())
		b := openapi3.NewObjectSchema().WithProperty("b", openapi3.NewStringSchema())
		result, err := merge(openapi3.Schema{Items: openapi3.NewSchemaRef("", a)},
			openapi3.Schema{Items: openapi3.NewSchemaRef("", b)})
		require.NoError(t, err)
		require.NotNil(t, result.Items)
		assert.Contains(t, result.Items.Value.Properties, "a")
		assert.Contains(t, result.Items.Value.Properties, "b")
	})

	t.Run("conflicting item types error", func(t *testing.T) {
		_, err := merge(openapi3.Schema{Items: openapi3.NewSchemaRef("", openapi3.NewStringSchema())},
			openapi3.Schema{Items: openapi3.NewSchemaRef("", openapi3.NewIntegerSchema())})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "array items")
	})
}
