package codegen

// These tests cover the allOf merge rules of schema-merging-behavior v3. They
// started as a copy of v2's (merge_schemas_v2_test.go).

import (
	"maps"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mergeTwoV3 merges two allOf members, allOf/0 and allOf/1, with
// schema-merging-behavior v3's rules.
func mergeTwoV3(s1, s2 openapi3.Schema) (openapi3.Schema, error) {
	m := newAllOfMerge(newGenContext(nil))
	if err := m.add(s1, "allOf/0", map[string]bool{}); err != nil {
		return openapi3.Schema{}, err
	}
	if err := m.add(s2, "allOf/1", map[string]bool{}); err != nil {
		return openapi3.Schema{}, err
	}
	return m.result()
}

func TestMergeOpenapiSchemas_DiscriminatorPropagationV3(t *testing.T) {
	disc := &openapi3.Discriminator{
		PropertyName: "type",
	}

	t.Run("allOf with single discriminator on s1 propagates it", func(t *testing.T) {
		s1 := openapi3.Schema{Discriminator: disc}
		s2 := openapi3.Schema{}

		result, err := mergeTwoV3(s1, s2)
		require.NoError(t, err)
		assert.Equal(t, disc, result.Discriminator)
	})

	t.Run("allOf with single discriminator on s2 propagates it", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{Discriminator: disc}

		result, err := mergeTwoV3(s1, s2)
		require.NoError(t, err)
		assert.Equal(t, disc, result.Discriminator)
	})

	t.Run("allOf with different discriminators errors", func(t *testing.T) {
		disc2 := &openapi3.Discriminator{PropertyName: "kind"}
		s1 := openapi3.Schema{Discriminator: disc}
		s2 := openapi3.Schema{Discriminator: disc2}

		_, err := mergeTwoV3(s1, s2)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "allOf can't merge allOf/0 (discriminator type) with allOf/1 (discriminator kind)")
	})

	t.Run("allOf with the same discriminator on both schemas merges", func(t *testing.T) {
		mapping := map[string]openapi3.MappingRef{"cat": {Ref: "#/components/schemas/Cat"}}
		s1 := openapi3.Schema{Discriminator: &openapi3.Discriminator{PropertyName: "type", Mapping: mapping}}
		s2 := openapi3.Schema{Discriminator: &openapi3.Discriminator{PropertyName: "type", Mapping: maps.Clone(mapping)}}

		result, err := mergeTwoV3(s1, s2)
		require.NoError(t, err)
		assert.Equal(t, "type", result.Discriminator.PropertyName)

		s2.Discriminator.Mapping["dog"] = openapi3.MappingRef{Ref: "#/components/schemas/Dog"}
		_, err = mergeTwoV3(s1, s2)
		require.Error(t, err, "a different mapping is a different discriminator")
	})

	t.Run("allOf with no discriminators succeeds with nil discriminator", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{}

		result, err := mergeTwoV3(s1, s2)
		require.NoError(t, err)
		assert.Nil(t, result.Discriminator)
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

		result, err := mergeTwoV3(s1, s2)
		require.NoError(t, err)
		assert.Equal(t, stringType, result.Type)
	})

	t.Run("type on s1 only propagates", func(t *testing.T) {
		s1 := openapi3.Schema{Type: stringType}
		s2 := openapi3.Schema{}

		result, err := mergeTwoV3(s1, s2)
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

		result, err := mergeTwoV3(s1, s2)
		require.NoError(t, err)
		assert.Equal(t, unionType, result.Type)
	})

	t.Run("equal types on both members merge", func(t *testing.T) {
		s1 := openapi3.Schema{Type: stringType}
		s2 := openapi3.Schema{Type: stringType}

		result, err := mergeTwoV3(s1, s2)
		require.NoError(t, err)
		assert.Equal(t, stringType, result.Type)
	})

	t.Run("different types on both members error", func(t *testing.T) {
		s1 := openapi3.Schema{Type: stringType}
		s2 := openapi3.Schema{Type: numberType}

		_, err := mergeTwoV3(s1, s2)
		require.Error(t, err)
		assert.EqualError(t, err, "allOf can't merge allOf/0 (type string) with allOf/1 (type number): no value has both types")
	})

	t.Run("integer narrows number", func(t *testing.T) {
		for _, types := range [][2]*openapi3.Types{{numberType, {"integer"}}, {{"integer"}, numberType}} {
			result, err := mergeTwoV3(openapi3.Schema{Type: types[0]}, openapi3.Schema{Type: types[1]})
			require.NoError(t, err)
			assert.Equal(t, []string{"integer"}, result.Type.Slice())
		}
	})

	t.Run("a type list with number and integer is number", func(t *testing.T) {
		intNum := func() *openapi3.Types { return &openapi3.Types{"integer", "number"} }
		for want, other := range map[string]*openapi3.Types{"number": numberType, "integer": {"integer"}} {
			result, err := mergeTwoV3(openapi3.Schema{Type: intNum()}, openapi3.Schema{Type: other})
			require.NoError(t, err)
			assert.Equal(t, []string{want}, result.Type.Slice())
		}
	})

	t.Run("multi-type arrays intersect", func(t *testing.T) {
		result, err := mergeTwoV3(openapi3.Schema{Type: &openapi3.Types{"string", "number", "boolean"}},
			openapi3.Schema{Type: &openapi3.Types{"integer", "string"}})
		require.NoError(t, err)
		assert.Equal(t, []string{"string", "integer"}, result.Type.Slice())
	})

	t.Run("neither member typed stays typeless", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{}

		result, err := mergeTwoV3(s1, s2)
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

		result, err := mergeTwoV3(s1, s2)
		require.NoError(t, err)
		assert.Equal(t, "uuid", result.Format)
	})

	t.Run("format on s1 only propagates", func(t *testing.T) {
		s1 := openapi3.Schema{Format: "uuid"}
		s2 := openapi3.Schema{}

		result, err := mergeTwoV3(s1, s2)
		require.NoError(t, err)
		assert.Equal(t, "uuid", result.Format)
	})

	t.Run("equal formats merge", func(t *testing.T) {
		s1 := openapi3.Schema{Format: "uuid"}
		s2 := openapi3.Schema{Format: "uuid"}

		result, err := mergeTwoV3(s1, s2)
		require.NoError(t, err)
		assert.Equal(t, "uuid", result.Format)
	})

	t.Run("different formats error", func(t *testing.T) {
		s1 := openapi3.Schema{Format: "uuid"}
		s2 := openapi3.Schema{Format: "date-time"}

		_, err := mergeTwoV3(s1, s2)
		require.Error(t, err)
		assert.EqualError(t, err, "allOf can't merge allOf/0 (format uuid) with allOf/1 (format date-time): a value can't have both formats")
	})

	t.Run("nullable decorator over format-carrying scalar merges", func(t *testing.T) {
		s1 := openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "uuid"}
		s2 := openapi3.Schema{Nullable: true}

		result, err := mergeTwoV3(s1, s2)
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

		result, err := mergeTwoV3(s1, s2)
		require.NoError(t, err)
		assert.True(t, result.Nullable)
	})

	t.Run("nullable on s1 propagates to result", func(t *testing.T) {
		s1 := openapi3.Schema{Nullable: true}
		s2 := openapi3.Schema{}

		result, err := mergeTwoV3(s1, s2)
		require.NoError(t, err)
		assert.True(t, result.Nullable)
	})

	t.Run("both nullable stays nullable", func(t *testing.T) {
		s1 := openapi3.Schema{Nullable: true}
		s2 := openapi3.Schema{Nullable: true}

		result, err := mergeTwoV3(s1, s2)
		require.NoError(t, err)
		assert.True(t, result.Nullable)
	})

	t.Run("neither nullable stays non-nullable", func(t *testing.T) {
		s1 := openapi3.Schema{}
		s2 := openapi3.Schema{}

		result, err := mergeTwoV3(s1, s2)
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
		result, err := mergeTwoV3(s1, s2)
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
		set := openapi3.Schema{ReadOnly: true, WriteOnly: true, AllowEmptyValue: true}
		for name, pair := range map[string][2]openapi3.Schema{
			"s1": {set, {}},
			"s2": {{}, set},
		} {
			t.Run(name, func(t *testing.T) {
				result := merge(t, pair[0], pair[1])
				assert.True(t, result.ReadOnly, "readOnly")
				assert.True(t, result.WriteOnly, "writeOnly")
				assert.True(t, result.AllowEmptyValue, "allowEmptyValue")
			})
		}
	})
}

// TestMergeOpenapiSchemas_NullInTypeArrayV3 covers OpenAPI 3.1 type arrays,
// where "null" spells nullability. It is unioned like 3.0's `nullable`
// instead of taking part in the type comparison.
func TestMergeOpenapiSchemas_NullInTypeArrayV3(t *testing.T) {
	object := &openapi3.Types{"object"}
	nullableObject := &openapi3.Types{"object", "null"}

	merge := func(s1, s2 openapi3.Schema) (openapi3.Schema, error) {
		return mergeTwoV3(s1, s2)
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

	t.Run("types that already agree stay as they are", func(t *testing.T) {
		result, err := merge(openapi3.Schema{Type: object}, openapi3.Schema{Type: &openapi3.Types{"object"}})
		require.NoError(t, err)
		assert.Equal(t, []string{"object"}, result.Type.Slice())
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
		assert.Contains(t, err.Error(), "no value has both types")
	})
}

// TestMergeOpenapiSchemas_AdditionalPropertiesV3 covers two members that both
// declare an additionalProperties schema.
func TestMergeOpenapiSchemas_AdditionalPropertiesV3(t *testing.T) {
	withAP := func(ap *openapi3.SchemaRef) openapi3.Schema {
		return openapi3.Schema{AdditionalProperties: openapi3.AdditionalProperties{Schema: ap}}
	}
	merge := func(s1, s2 openapi3.Schema) (openapi3.Schema, error) {
		return mergeTwoV3(s1, s2)
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

	t.Run("different schemas become an allOf of both", func(t *testing.T) {
		s1 := openapi3.NewSchemaRef("", openapi3.NewStringSchema())
		s2 := openapi3.NewSchemaRef("", openapi3.NewStringSchema().WithMaxLength(3))
		result, err := merge(withAP(s1), withAP(s2))
		require.NoError(t, err)
		assertAllOfOf(t, result.AdditionalProperties.Schema, s1, s2)
	})

	t.Run("false on either member closes the object", func(t *testing.T) {
		closed := openapi3.Schema{}
		closed.WithoutAdditionalProperties()
		result, err := merge(withAP(openapi3.NewSchemaRef("", openapi3.NewStringSchema())), closed)
		require.NoError(t, err)
		assert.True(t, isAdditionalPropertiesExplicitFalse(&result))
	})
}

// assertAllOfOf asserts that ref is an allOf of copies of the want schemas,
// which is how the merge combines schemas that members declare for one
// position.
func assertAllOfOf(t *testing.T, ref *openapi3.SchemaRef, want ...*openapi3.SchemaRef) {
	t.Helper()
	require.NotNil(t, ref)
	require.Len(t, ref.Value.AllOf, len(want))
	for i, w := range want {
		assert.Same(t, w.Value, ref.Value.AllOf[i].Value)
		assert.Equal(t, w.Ref, ref.Value.AllOf[i].Ref)
	}
}

// TestMergeOpenapiSchemas_ItemsV3 covers array items, which used to be dropped
// by every allOf merge.
func TestMergeOpenapiSchemas_ItemsV3(t *testing.T) {
	merge := func(s1, s2 openapi3.Schema) (openapi3.Schema, error) {
		return mergeTwoV3(s1, s2)
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

	t.Run("different item schemas become an allOf of both", func(t *testing.T) {
		a := openapi3.NewSchemaRef("", openapi3.NewObjectSchema().WithProperty("a", openapi3.NewStringSchema()))
		b := openapi3.NewSchemaRef("", openapi3.NewObjectSchema().WithProperty("b", openapi3.NewStringSchema()))
		result, err := merge(openapi3.Schema{Items: a}, openapi3.Schema{Items: b})
		require.NoError(t, err)
		assertAllOfOf(t, result.Items, a, b)
	})
}

func TestMergeOpenapiSchemas_EnumV3(t *testing.T) {
	named := func(enum []any, names ...any) openapi3.Schema {
		s := openapi3.Schema{Enum: enum}
		if len(names) > 0 {
			s.Extensions = map[string]any{extEnumVarNames: names}
		}
		return s
	}

	t.Run("an enum on one member is kept", func(t *testing.T) {
		result, err := mergeTwoV3(named([]any{"a", "b"}, "A", "B"), openapi3.Schema{})
		require.NoError(t, err)
		assert.Equal(t, []any{"a", "b"}, result.Enum)
		assert.Equal(t, []any{"A", "B"}, result.Extensions[extEnumVarNames])
	})

	t.Run("two enums intersect, in the first member's order, with the later member's names (issue #1633)", func(t *testing.T) {
		result, err := mergeTwoV3(named([]any{"a", "b", "c"}, "A", "B", "C"), named([]any{"c", "a", "z"}, "X", "Y", "Z"))
		require.NoError(t, err)
		assert.Equal(t, []any{"a", "c"}, result.Enum)
		assert.Equal(t, []any{"Y", "X"}, result.Extensions[extEnumVarNames])
	})

	t.Run("a later member without names keeps the earlier names", func(t *testing.T) {
		result, err := mergeTwoV3(named([]any{"a", "b", "c"}, "A", "B", "C"), named([]any{"c", "a"}))
		require.NoError(t, err)
		assert.Equal(t, []any{"A", "C"}, result.Extensions[extEnumVarNames])
	})

	t.Run("names come from the second member when the first has none", func(t *testing.T) {
		result, err := mergeTwoV3(named([]any{"a", "b", "c"}), named([]any{"c", "a"}, "X", "Y"))
		require.NoError(t, err)
		assert.Equal(t, []any{"a", "c"}, result.Enum)
		assert.Equal(t, []any{"Y", "X"}, result.Extensions[extEnumVarNames])
	})

	t.Run("a member without an enum can rename its values", func(t *testing.T) {
		result, err := mergeTwoV3(named([]any{"a", "b"}, "A", "B"),
			openapi3.Schema{Extensions: map[string]any{extEnumNames: []any{"First", "Second"}}})
		require.NoError(t, err)
		assert.Equal(t, []any{"First", "Second"}, result.Extensions[extEnumVarNames])
	})

	t.Run("enums with no value in common error", func(t *testing.T) {
		_, err := mergeTwoV3(named([]any{"a", "b"}), named([]any{"c"}))
		assert.EqualError(t, err, "allOf can't merge allOf/0 (enum [a b]) with allOf/1 (enum [c]): no value is in both. "+
			"To allow the values of either, set x-oapi-codegen-enum-merge: union on the schema with this allOf")
	})

	t.Run("x-oapi-codegen-enum-merge: union takes the values of either", func(t *testing.T) {
		m := newAllOfMerge(newGenContext(nil))
		m.unionEnums = true
		require.NoError(t, m.add(named([]any{"a", "b"}, "A", "B"), "allOf/0", map[string]bool{}))
		require.NoError(t, m.add(named([]any{"b", "c"}), "allOf/1", map[string]bool{}))
		result, err := m.result()
		require.NoError(t, err)
		assert.Equal(t, []any{"a", "b", "c"}, result.Enum)
		assert.Equal(t, []any{"A", "B", "c"}, result.Extensions[extEnumVarNames],
			"a value no member names is named after itself")
	})
}

func TestMergeOpenapiSchemas_ConstV3(t *testing.T) {
	t.Run("the same const merges", func(t *testing.T) {
		result, err := mergeTwoV3(openapi3.Schema{Const: "cat"}, openapi3.Schema{Const: "cat"})
		require.NoError(t, err)
		assert.Equal(t, "cat", result.Const)
	})

	t.Run("different consts error", func(t *testing.T) {
		_, err := mergeTwoV3(openapi3.Schema{Const: "cat"}, openapi3.Schema{Const: "dog"})
		assert.EqualError(t, err, "allOf can't merge allOf/0 (const cat) with allOf/1 (const dog): no value is both")
	})

	t.Run("a const narrows an enum", func(t *testing.T) {
		result, err := mergeTwoV3(openapi3.Schema{Enum: []any{"cat", "dog"}}, openapi3.Schema{Const: "dog"})
		require.NoError(t, err)
		assert.Equal(t, []any{"dog"}, result.Enum)
	})

	t.Run("a const the enum doesn't allow errors", func(t *testing.T) {
		_, err := mergeTwoV3(openapi3.Schema{Enum: []any{"cat", "dog"}}, openapi3.Schema{Const: "cow"})
		assert.EqualError(t, err, "allOf can't merge allOf/1 (const cow) with allOf/0 (enum [cat dog]): the enum doesn't allow the const")
	})
}

func TestMergeOpenapiSchemas_RequiredV3(t *testing.T) {
	result, err := mergeTwoV3(openapi3.Schema{Required: []string{"a", "b"}}, openapi3.Schema{Required: []string{"b", "c"}})
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, result.Required)
}

// TestMergeOpenapiSchemas_NestedAllOfV3: a member that is itself an allOf
// contributes its members and its own keywords; v2 dropped the latter.
func TestMergeOpenapiSchemas_NestedAllOfV3(t *testing.T) {
	nested := openapi3.Schema{
		Properties: openapi3.Schemas{"own": openapi3.NewSchemaRef("", openapi3.NewStringSchema())},
		AllOf: openapi3.SchemaRefs{openapi3.NewSchemaRef("", &openapi3.Schema{
			Properties: openapi3.Schemas{"inner": openapi3.NewSchemaRef("", openapi3.NewStringSchema())},
		})},
	}
	result, err := mergeTwoV3(nested, openapi3.Schema{
		Properties: openapi3.Schemas{"other": openapi3.NewSchemaRef("", openapi3.NewStringSchema())},
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"own", "inner", "other"}, slices.Collect(maps.Keys(result.Properties)))

	nested.AllOf[0].Value.Type = &openapi3.Types{"array"}
	_, err = mergeTwoV3(nested, openapi3.Schema{Type: &openapi3.Types{"object"}})
	assert.EqualError(t, err, "allOf can't merge allOf/0/allOf/0 (type array) with allOf/1 (type object): no value has both types",
		"the error names the nested member")
}

func TestMergeSchemasV3EndToEnd(t *testing.T) {
	t.Run("number, and number or integer, is number (3.1)", func(t *testing.T) {
		code := generateSpec(t, opaqueSpecHeader31+`
    Amount:
      allOf:
        - type: [integer, number]
        - type: number
`, withV3)
		assert.Contains(t, code, "type Amount = float32")
	})

	t.Run("integer narrows number", func(t *testing.T) {
		code := generateSpec(t, opaqueSpecHeader+`
    Count:
      allOf:
        - type: number
        - type: integer
          format: int64
`, withV3)
		assert.Contains(t, code, "type Count = int64")
	})

	t.Run("a conflict names both members", func(t *testing.T) {
		_, err := generateSpecErr(opaqueSpecHeader+`
    Base:
      allOf:
        - type: object
          properties:
            name: {type: string}
    Odd:
      type: object
      properties:
        extra: {type: string}
      allOf:
        - $ref: '#/components/schemas/Base'
        - type: string
`, withV3)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error converting Schema Odd to Go type: error merging schemas: "+
			"allOf can't merge #/components/schemas/Base/allOf/0 (type object) with allOf/1 (type string): no value has both types")
	})

	t.Run("the schema's own keywords are the schema itself", func(t *testing.T) {
		_, err := generateSpecErr(opaqueSpecHeader+`
    Odd:
      type: string
      required: [x]
      allOf:
        - type: integer
`, withV3)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "allOf can't merge allOf/0 (type integer) with the schema itself (type string)")
	})
}

// TestMergeSchemasV3Properties covers properties, items and
// additionalProperties that several members declare (issue #2107).
func TestMergeSchemasV3Properties(t *testing.T) {
	const base = `
    Base:
      type: object
      required: [name]
      properties:
        name: {type: string, description: The name.}
        age: {type: integer}
        address:
          type: object
          properties:
            street: {type: string}
`

	t.Run("a member refines a property instead of replacing it", func(t *testing.T) {
		code := generateSpec(t, opaqueSpecHeader+base+`
    Patch:
      allOf:
        - $ref: '#/components/schemas/Base'
        - properties:
            name:
              nullable: true
              x-go-name: Moniker
            address:
              properties:
                zip: {type: string}
`, withV3)
		assert.Regexp(t, `// Moniker The name\.\n\tMoniker \*string `+"`json:\"name\"`", code,
			"name keeps Base's type and description, and gains nullable and the member's field name")
		assertField(t, code, "Age", "*int")
		assert.Regexp(t, `Address \*struct \{\n\t\tStreet \*string [^\n]*\n\t\tZip    \*string`, code)
	})

	t.Run("a property whose types conflict names both members", func(t *testing.T) {
		_, err := generateSpecErr(opaqueSpecHeader+base+`
    Odd:
      allOf:
        - $ref: '#/components/schemas/Base'
        - properties:
            age: {type: string}
`, withV3)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error generating Go schema for property 'age': error merging schemas: "+
			"allOf can't merge #/components/schemas/Base/properties/age (type integer) with allOf/1/properties/age (type string): "+
			"no value has both types")
	})

	t.Run("items and additionalProperties merge the same way", func(t *testing.T) {
		code := generateSpec(t, opaqueSpecHeader+`
    Tags:
      allOf:
        - type: array
          items: {type: string}
        - type: array
          items: {maxLength: 10, nullable: true}
    Labels:
      allOf:
        - additionalProperties: {type: string}
        - additionalProperties: {maxLength: 10}
`, withV3)
		assert.Contains(t, code, "type Tags = []*string")
		assert.Contains(t, code, "type Labels map[string]string")

		_, err := generateSpecErr(opaqueSpecHeader+`
    Tags:
      allOf:
        - type: array
          items: {type: string}
        - type: array
          items: {type: integer}
`, withV3)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "allOf can't merge allOf/0/items (type string) with allOf/1/items (type integer)")
	})

	t.Run("a merged property that refers back to the composition", func(t *testing.T) {
		code := generateSpec(t, opaqueSpecHeader+`
    Node:
      allOf:
        - type: object
          properties:
            name: {type: string}
            next: {$ref: '#/components/schemas/Node'}
        - properties:
            next: {nullable: true}
`, withV3)
		// next is an allOf of Node and {nullable: true}, which only annotates
		// Node: it is Node.
		assert.Contains(t, code, "type Node struct {")
		assertField(t, code, "Next", "*Node")
	})

	t.Run("the same $ref with extensions of its own stays that $ref", func(t *testing.T) {
		code := generateSpec(t, opaqueSpecHeader+`
    User:
      type: object
      properties:
        name: {type: string}
    Base:
      type: object
      properties:
        owner: {$ref: '#/components/schemas/User'}
    Patch:
      allOf:
        - $ref: '#/components/schemas/Base'
        - properties:
            owner:
              $ref: '#/components/schemas/User'
              x-go-name: Proprietor
`, withV3)
		patch := code[strings.Index(code, "type Patch struct"):]
		assertField(t, patch, "Proprietor", "*User")
	})

	t.Run("a recursive $ref with extensions of its own", func(t *testing.T) {
		code := generateSpec(t, opaqueSpecHeader+`
    Node:
      allOf:
        - type: object
          properties:
            name: {type: string}
            next:
              $ref: '#/components/schemas/Node'
              x-go-name: First
        - properties:
            next:
              $ref: '#/components/schemas/Node'
              x-go-name: Second
        - properties:
            next: {nullable: true}
`, withV3)
		assert.Contains(t, code, "type Node struct {")
		assertField(t, code, "Second", "*Node")
	})

	t.Run("recursive compositions still generate", func(t *testing.T) {
		code := generateSpec(t, specRecursiveObject, withV3)
		assert.Contains(t, code, "type Node struct {")
	})
}

// TestEnumMergeExtensionV3 covers x-oapi-codegen-enum-merge, which lets a
// composition add values to an enum.
func TestEnumMergeExtensionV3(t *testing.T) {
	const statuses = `
    BaseStatus:
      type: string
      enum: [active, inactive]
    ExtendedStatus:
      x-oapi-codegen-enum-merge: union
      allOf:
        - $ref: '#/components/schemas/BaseStatus'
        - enum: [archived]
`
	enumValues := func(t *testing.T, code, typeName string) []string {
		t.Helper()
		start := strings.Index(code, "// Defines values for "+typeName+".")
		require.GreaterOrEqual(t, start, 0, "no enum %s", typeName)
		block := code[start:]
		block = block[:strings.Index(block, ")")]
		var values []string
		for _, m := range regexp.MustCompile(typeName+` = "([^"]*)"`).FindAllStringSubmatch(block, -1) {
			values = append(values, m[1])
		}
		return values
	}

	t.Run("union", func(t *testing.T) {
		code := generateSpec(t, opaqueSpecHeader+statuses+`
    Decorated:
      allOf:
        - $ref: '#/components/schemas/ExtendedStatus'
        - description: A decorated ExtendedStatus.
    Narrowed:
      allOf:
        - $ref: '#/components/schemas/ExtendedStatus'
        - enum: [inactive, archived]
    Holder:
      type: object
      properties:
        status:
          x-oapi-codegen-enum-merge: union
          allOf:
            - $ref: '#/components/schemas/BaseStatus'
            - enum: [archived]
    Base:
      type: object
      properties:
        status: {$ref: '#/components/schemas/BaseStatus'}
    Patch:
      allOf:
        - $ref: '#/components/schemas/Base'
        - properties:
            status:
              enum: [archived]
              x-oapi-codegen-enum-merge: union
`, withV3)
		for typeName, why := range map[string]string{
			"ExtendedStatus": "a component",
			"HolderStatus":   "a property that is a composition",
			"PatchStatus":    "a property a member refines, where the extension is on the member's property",
		} {
			assert.ElementsMatch(t, []string{"active", "inactive", "archived"}, enumValues(t, code, typeName), why)
		}
		assert.Contains(t, code, "type Decorated = ExtendedStatus")
		assert.ElementsMatch(t, []string{"inactive", "archived"}, enumValues(t, code, "Narrowed"),
			"merging the union composition with another enum starts from its values")
	})

	t.Run("intersection is the default", func(t *testing.T) {
		_, err := generateSpecErr(opaqueSpecHeader+strings.ReplaceAll(statuses, "x-oapi-codegen-enum-merge: union", "x-oapi-codegen-enum-merge: intersection"), withV3)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no value is in both")
	})

	t.Run("an invalid value is an error", func(t *testing.T) {
		_, err := generateSpecErr(opaqueSpecHeader+strings.ReplaceAll(statuses, "union", "both"), withV3)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `invalid value for "x-oapi-codegen-enum-merge": must be "union" or "intersection", not both`)
	})
}

// TestEnumMergeScopeV3: x-oapi-codegen-enum-merge applies to the schema it's
// on. A nested composition merges its own way, the default included, and a
// merged property that needs the union says where to put it.
func TestEnumMergeScopeV3(t *testing.T) {
	code := generateSpec(t, opaqueSpecHeader+`
    Inner:
      allOf:
        - type: string
          enum: [a, b]
        - enum: [b, c]
    Outer:
      x-oapi-codegen-enum-merge: union
      allOf:
        - $ref: '#/components/schemas/Inner'
        - enum: [x]
`, withV3)
	assert.Contains(t, code, `OuterB Outer = "b"`)
	assert.Contains(t, code, `OuterX Outer = "x"`)
	assert.NotContains(t, code, "OuterA", "Inner only allows b")
	assert.NotContains(t, code, "OuterC")

	_, err := generateSpecErr(opaqueSpecHeader+`
    Base:
      type: object
      properties:
        status: {type: string, enum: ["on", "off"]}
    Extended:
      x-oapi-codegen-enum-merge: union
      allOf:
        - $ref: '#/components/schemas/Base'
        - properties:
            status: {enum: [unknown]}
`, withV3)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "set x-oapi-codegen-enum-merge: union on allOf/1/properties/status")
}

// TestEnumRenameMakesANewEnumV3: x-enum-varnames next to a $ref'd enum
// renames its values, which takes a new enum type.
func TestEnumRenameMakesANewEnumV3(t *testing.T) {
	code := generateSpec(t, opaqueSpecHeader+decoratorComponents+`
    Renamed:
      allOf:
        - $ref: '#/components/schemas/Color'
        - x-enum-varnames: [Rouge, Vert]
`, withV3)
	assert.Contains(t, code, "type Renamed string")
	assert.Contains(t, code, `Rouge Renamed = "red"`)
	assert.Contains(t, code, `Vert  Renamed = "green"`)

	// Restating the enum's type doesn't make the names an annotation.
	code = generateSpec(t, opaqueSpecHeader+decoratorComponents+`
    Renamed:
      allOf:
        - $ref: '#/components/schemas/Color'
        - type: string
          x-enumNames: [Rouge, Vert]
`, withV3)
	assert.Contains(t, code, "type Renamed string")
	assert.Contains(t, code, `Rouge Renamed = "red"`)
}

// TestOwnKeywordsV3: the schema's own keywords next to allOf constrain the
// value like one more member: `{type: string, allOf: [A]}` is
// `{allOf: [A, {type: string}]}`.
func TestOwnKeywordsV3(t *testing.T) {
	code := generateSpec(t, opaqueSpecHeader+`
    Str: {type: string}
    Letters:
      enum: [a, b]
      allOf:
        - $ref: '#/components/schemas/Str'
    Day:
      format: date
      allOf:
        - $ref: '#/components/schemas/Str'
    Counts:
      type: array
      items: {type: integer}
      allOf:
        - type: array
          minItems: 1
`, withV3)
	assert.Contains(t, code, `A Letters = "a"`)
	assert.Contains(t, code, "type Letters string")
	assert.Contains(t, code, "type Day = openapi_types.Date")
	assert.Contains(t, code, "type Counts = []int")

	_, err := generateSpecErr(opaqueSpecHeader+`
    Base:
      type: object
      properties:
        name: {type: string}
    Odd:
      type: string
      allOf:
        - $ref: '#/components/schemas/Base'
`, withV3)
	require.Error(t, err)
	assert.Contains(t, err.Error(),
		"allOf can't merge #/components/schemas/Base (type object) with the schema itself (type string): no value has both types")
}

func TestConstraintOnlyUnionV3(t *testing.T) {
	object := &openapi3.Schema{Type: &openapi3.Types{"object"}}
	required := func(names ...string) *openapi3.SchemaRef {
		return openapi3.NewSchemaRef("", &openapi3.Schema{Required: names})
	}
	null := openapi3.NewSchemaRef("", &openapi3.Schema{Type: &openapi3.Types{"null"}})
	closed := false
	for name, tc := range map[string]struct {
		branches openapi3.SchemaRefs
		want     bool
	}{
		"required only":                {openapi3.SchemaRefs{required("a"), required("b")}, true},
		"restated type":                {openapi3.SchemaRefs{openapi3.NewSchemaRef("", &openapi3.Schema{Type: &openapi3.Types{"object"}, Required: []string{"a"}})}, true},
		"validation too":               {openapi3.SchemaRefs{openapi3.NewSchemaRef("", &openapi3.Schema{Required: []string{"a"}, MinProps: 1})}, true},
		"another type":                 {openapi3.SchemaRefs{openapi3.NewSchemaRef("", &openapi3.Schema{Type: &openapi3.Types{"string"}})}, false},
		"properties":                   {openapi3.SchemaRefs{openapi3.NewSchemaRef("", openapi3.NewObjectSchema().WithProperty("a", openapi3.NewStringSchema()))}, false},
		"a type branch":                {openapi3.SchemaRefs{required("a"), openapi3.NewSchemaRef("#/components/schemas/A", openapi3.NewObjectSchema().WithProperty("a", openapi3.NewStringSchema()))}, false},
		"a local $ref to a constraint": {openapi3.SchemaRefs{openapi3.NewSchemaRef("#/components/schemas/NeedA", &openapi3.Schema{Required: []string{"a"}})}, true},
		"a $ref into another document": {openapi3.SchemaRefs{openapi3.NewSchemaRef("./other.yaml#/components/schemas/NeedA", &openapi3.Schema{Required: []string{"a"}})}, false},
		"a null branch":                {openapi3.SchemaRefs{required("a"), null}, true},
		"only a null branch":           {openapi3.SchemaRefs{null}, false},
		"closed":                       {openapi3.SchemaRefs{openapi3.NewSchemaRef("", &openapi3.Schema{Required: []string{"a"}, AdditionalProperties: openapi3.AdditionalProperties{Has: &closed}})}, true},
		"additionalProperties schema":  {openapi3.SchemaRefs{openapi3.NewSchemaRef("", &openapi3.Schema{AdditionalProperties: openapi3.AdditionalProperties{Schema: openapi3.NewStringSchema().NewRef()}})}, false},
		"x-go-type":                    {openapi3.SchemaRefs{openapi3.NewSchemaRef("", &openapi3.Schema{Required: []string{"a"}, Extensions: map[string]any{extPropGoType: "T"}})}, false},
		"x-go-type-name":               {openapi3.SchemaRefs{openapi3.NewSchemaRef("", &openapi3.Schema{Required: []string{"a"}, Extensions: map[string]any{extGoTypeName: "NeedsA"}})}, false},
		"empty":                        {nil, false},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, isConstraintOnlyUnionV3(tc.branches, object))
		})
	}

	// An owner with no type of its own has the type its allOf members declare.
	composed := &openapi3.Schema{AllOf: openapi3.SchemaRefs{openapi3.NewSchemaRef("#/components/schemas/Contact", object)}}
	restated := openapi3.SchemaRefs{openapi3.NewSchemaRef("", &openapi3.Schema{Type: &openapi3.Types{"object"}, Required: []string{"a"}})}
	assert.True(t, isConstraintOnlyUnionV3(restated, composed))
	assert.False(t, isConstraintOnlyUnionV3(restated, &openapi3.Schema{}), "nothing is declared to restate")
	annotated := &openapi3.Schema{AllOf: openapi3.SchemaRefs{
		{Value: &openapi3.Schema{AllOf: openapi3.SchemaRefs{{Value: &openapi3.Schema{Description: "x"}}}}},
		openapi3.NewSchemaRef("#/components/schemas/Contact", object),
	}}
	assert.True(t, isConstraintOnlyUnionV3(restated, annotated), "a later member declares the type")
}

// TestConstraintOnlyUnionEndToEndV3: a oneOf or anyOf that only adds
// constraints makes no union wherever it is (issue #839).
func TestConstraintOnlyUnionEndToEndV3(t *testing.T) {
	code := generateSpec(t, opaqueSpecHeader+`
    Contact:
      type: object
      properties:
        email: {type: string}
        phone: {type: string}
      oneOf:
        - required: [email]
        - required: [phone]
    Holder:
      type: object
      properties:
        contact:
          properties:
            email: {type: string}
          anyOf:
            - required: [email]
    Nested:
      allOf:
        - allOf:
            - $ref: '#/components/schemas/Contact'
            - oneOf:
                - required: [email]
        - properties:
            extra: {type: string}
    Dated:
      oneOf:
        - type: string
          format: date
        - type: string
          format: date-time
`, withV3)
	assert.Equal(t, 1, strings.Count(code, "union json.RawMessage"), "only Dated is a union")
	assert.NotContains(t, code, "Contact0")
	assert.Contains(t, code, "type Contact struct {")
	assert.Regexp(t, `Contact \*struct \{\n\t\tEmail \*string`, code)
	assert.Contains(t, code, "type Nested struct {")
	assert.Contains(t, code, "func (t Dated) AsDated0()", "branches with a type and format of their own are types")
}

// TestConstraintOnlyUnionRestatedTypeV3: a oneOf restating the type an allOf
// member declares still only adds constraints, whether it's in a member of
// the allOf or next to it, so the composition is the member's type.
func TestConstraintOnlyUnionRestatedTypeV3(t *testing.T) {
	code := generateSpec(t, opaqueSpecHeader+`
    Contact:
      type: object
      properties:
        email: {type: string}
        phone: {type: string}
    InMember:
      allOf:
        - $ref: '#/components/schemas/Contact'
        - oneOf:
            - type: object
              required: [email]
            - type: object
              required: [phone]
    Beside:
      allOf:
        - $ref: '#/components/schemas/Contact'
      oneOf:
        - type: object
          required: [email]
        - {type: 'null'}
`, withV3)
	assert.NotContains(t, code, "union json.RawMessage")
	assert.Contains(t, code, "type InMember = Contact")
	assert.Contains(t, code, "type Beside = Contact")
}

// TestParentListsChildrenV3: a member whose oneOf or anyOf lists the
// composition being merged, or a schema the merge flattens into it, adds no
// union: a child that is an allOf of its parent is one of the parent's
// branches, not a union of all of them.
func TestParentListsChildrenV3(t *testing.T) {
	code := generateSpec(t, opaqueSpecHeader+`
    Pet:
      type: object
      required: [petType]
      properties:
        petType: {type: string}
      discriminator:
        propertyName: petType
      oneOf:
        - $ref: '#/components/schemas/Cat'
        - $ref: '#/components/schemas/Dog'
    Cat:
      allOf:
        - $ref: '#/components/schemas/Pet'
        - properties:
            meow: {type: string}
    Dog:
      allOf:
        - $ref: '#/components/schemas/Pet'
        - properties:
            bark: {type: string}
    Kitten:
      allOf:
        - $ref: '#/components/schemas/Cat'
        - properties:
            age: {type: integer}
    Shape:
      anyOf:
        - $ref: '#/components/schemas/Square'
    Square:
      allOf:
        - $ref: '#/components/schemas/Shape'
        - properties:
            side: {type: number}
    Unlisted:
      allOf:
        - $ref: '#/components/schemas/Pet'
        - properties:
            purr: {type: string}
`, withV3)
	for _, child := range []string{"Cat", "Dog", "Kitten", "Square"} {
		start := strings.Index(code, "type "+child+" struct {")
		require.GreaterOrEqual(t, start, 0, child)
		body := code[start : start+strings.Index(code[start:], "\n}")]
		assert.NotContains(t, body, "union json.RawMessage", child)
	}
	assert.Contains(t, code, "func (t Pet) AsCat() (Cat, error)", "the parent is still the union")
	assert.Contains(t, code, "func (t Unlisted) AsCat() (Cat, error)",
		"a schema the parent doesn't list is still one of its branches")
}

// TestParentListsChildOnlyV3: a child that is only an allOf of the parent
// that lists it, alone or with members that only annotate, is a struct of the
// parent's fields, not an alias of the parent's union. Only the composition
// being merged counts as listed: a property inside a child is still the
// parent's union.
func TestParentListsChildOnlyV3(t *testing.T) {
	code := generateSpec(t, opaqueSpecHeader+`
    Pet:
      type: object
      required: [petType]
      properties:
        petType: {type: string}
      discriminator:
        propertyName: petType
      oneOf:
        - $ref: '#/components/schemas/Cat'
        - $ref: '#/components/schemas/Dog'
        - $ref: '#/components/schemas/Bird'
    Cat:
      allOf:
        - $ref: '#/components/schemas/Pet'
    Dog:
      allOf:
        - $ref: '#/components/schemas/Pet'
        - description: A dog.
    Bird:
      allOf:
        - $ref: '#/components/schemas/Pet'
        - type: object
          properties:
            best:
              allOf:
                - $ref: '#/components/schemas/Pet'
                - properties:
                    since: {type: string}
`, withV3)
	assert.Contains(t, code, "type Cat struct {\n\tPetType string")
	assert.Contains(t, code, "type Dog struct {\n\tPetType string")
	assert.NotContains(t, code, "type Cat = Pet")
	assert.NotContains(t, code, "type Dog = Pet")
	assert.Contains(t, code, "func (t Bird_Best) AsCat() (Cat, error)", "best is a Pet, one of its branches")
}

// TestParentListsMemberV3: a union that lists another member of the same
// allOf, or a schema inside one, is left out whatever the members' order, and
// its discriminator goes with it.
func TestParentListsMemberV3(t *testing.T) {
	code := generateSpec(t, opaqueSpecHeader+`
    Pet:
      type: object
      required: [petType]
      properties:
        petType: {type: string}
      discriminator:
        propertyName: petType
      oneOf:
        - $ref: '#/components/schemas/Cat'
        - $ref: '#/components/schemas/Dog'
    Cat:
      allOf:
        - $ref: '#/components/schemas/Pet'
        - properties:
            meow: {type: string}
    Dog:
      allOf:
        - $ref: '#/components/schemas/Pet'
        - oneOf:
            - $ref: '#/components/schemas/Indoor'
            - $ref: '#/components/schemas/Outdoor'
    Indoor:
      type: object
      properties:
        room: {type: string}
    Outdoor:
      type: object
      properties:
        yard: {type: string}
    KittenA:
      allOf:
        - $ref: '#/components/schemas/Pet'
        - $ref: '#/components/schemas/Cat'
        - properties:
            age: {type: integer}
    KittenB:
      allOf:
        - $ref: '#/components/schemas/Cat'
        - $ref: '#/components/schemas/Pet'
        - properties:
            age: {type: integer}
    Nested:
      allOf:
        - $ref: '#/components/schemas/Pet'
        - allOf:
            - $ref: '#/components/schemas/Cat'
    Card:
      type: object
      properties:
        card: {type: string}
    Cash:
      type: object
      properties:
        amount: {type: number}
    Payment:
      oneOf:
        - $ref: '#/components/schemas/Card'
        - $ref: '#/components/schemas/Cash'
    Wallet:
      oneOf:
        - $ref: '#/components/schemas/Payment'
        - $ref: '#/components/schemas/Cash'
    BothA:
      allOf:
        - $ref: '#/components/schemas/Wallet'
        - $ref: '#/components/schemas/Payment'
    BothB:
      allOf:
        - $ref: '#/components/schemas/Payment'
        - $ref: '#/components/schemas/Wallet'
`, withV3)
	for _, name := range []string{"KittenA", "KittenB", "Nested"} {
		assert.NotContains(t, code, "func (t "+name+") AsDog()", name)
	}
	for _, name := range []string{"BothA", "BothB"} {
		assert.Contains(t, code, "func (t "+name+") AsCard() (Card, error)", name)
		assert.NotContains(t, code, "func (t "+name+") AsPayment()", name)
	}
	assert.Contains(t, code, "func (t Dog) AsIndoor() (Indoor, error)")
	assert.NotContains(t, code, "func (t Dog) ValueByDiscriminator(", "petType tells Pet's children apart, not Dog's own union")
	start := strings.Index(code, "func (t *Dog) FromIndoor(")
	require.GreaterOrEqual(t, start, 0)
	fromIndoor := code[start : start+strings.Index(code[start:], "\n}\n")]
	assert.NotContains(t, fromIndoor, `"Indoor"`, "nothing stamps petType")
}

// TestConstraintOnlyUnionMemberV3: a member's list whose branches restate the
// type another member declares only adds constraints, also when another
// member brings a real union; and a branch that names its Go type is a type.
func TestConstraintOnlyUnionMemberV3(t *testing.T) {
	code := generateSpec(t, opaqueSpecHeader+`
    Card:
      type: object
      properties:
        card: {type: string}
    Cash:
      type: object
      properties:
        amount: {type: number}
    Mixed:
      allOf:
        - type: object
          properties:
            email: {type: string}
            phone: {type: string}
        - oneOf:
            - type: object
              required: [email]
            - type: object
              required: [phone]
        - oneOf:
            - $ref: '#/components/schemas/Card'
            - $ref: '#/components/schemas/Cash'
    Named:
      type: object
      properties:
        email: {type: string}
        phone: {type: string}
      oneOf:
        - required: [email]
          x-go-type-name: EmailContact
        - required: [phone]
          x-go-type-name: PhoneContact
`, withV3)
	assert.Contains(t, code, "func (t Mixed) AsCard() (Card, error)")
	assert.NotContains(t, code, "Mixed0", "the constraint branches are no variants")
	assert.Contains(t, code, "type EmailContact = any")
	assert.Contains(t, code, "func (t Named) AsNamed0() (Named0, error)")

	// The type the branches restate is the one the members narrow to.
	code = generateSpec(t, opaqueSpecHeader31+`
    Card:
      type: object
      properties:
        card: {type: string}
    Cash:
      type: object
      properties:
        amount: {type: number}
    Narrowed:
      allOf:
        - type: [object, string]
        - type: object
          properties:
            email: {type: string}
            phone: {type: string}
        - oneOf:
            - type: object
              required: [email]
            - type: object
              required: [phone]
        - oneOf:
            - $ref: '#/components/schemas/Card'
            - $ref: '#/components/schemas/Cash'
`, withV3)
	assert.Contains(t, code, "func (t Narrowed) AsCard() (Card, error)")
	assert.NotContains(t, code, "Narrowed0", "the constraint branches are no variants")
}

// TestDeclaredTypes: a schema's declared types are its own, narrowed by what
// its allOf members declare, as the merge intersects them.
func TestDeclaredTypes(t *testing.T) {
	types := func(ts ...string) *openapi3.Schema { return &openapi3.Schema{Type: (*openapi3.Types)(&ts)} }
	allOf := func(own *openapi3.Schema, members ...*openapi3.Schema) *openapi3.Schema {
		for _, m := range members {
			own.AllOf = append(own.AllOf, m.NewRef())
		}
		return own
	}
	for name, tc := range map[string]struct {
		schema *openapi3.Schema
		want   []string
	}{
		"own type":                 {types("object", "null"), []string{"object"}},
		"properties":               {openapi3.NewObjectSchema().WithProperty("a", openapi3.NewStringSchema()), []string{"object"}},
		"nothing":                  {&openapi3.Schema{}, nil},
		"a later member narrows":   {allOf(&openapi3.Schema{}, types("object", "string"), &openapi3.Schema{}, types("object")), []string{"object"}},
		"a member narrows its own": {allOf(types("object", "string"), types("string")), []string{"string"}},
		"integer under number":     {allOf(&openapi3.Schema{}, types("number"), types("integer")), []string{"integer"}},
		"nested":                   {allOf(&openapi3.Schema{}, allOf(&openapi3.Schema{}, types("object", "array")), types("array")), []string{"array"}},
		"no value":                 {allOf(&openapi3.Schema{}, types("object"), types("string")), nil},
	} {
		t.Run(name, func(t *testing.T) {
			got := declaredTypes(tc.schema)
			if len(tc.want) == 0 {
				assert.Empty(t, got)
				return
			}
			assert.Equal(t, tc.want, got)
		})
	}
}
