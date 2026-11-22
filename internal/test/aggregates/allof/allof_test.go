package aggregatesallof

import (
	"reflect"
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
)

// issue #1905: a oneOf nested inside allOf-of-allOf must survive the merge.
// The generated struct must have a union field, and the two variant types must exist.
func TestIssue1905(t *testing.T) {
	typ := reflect.TypeOf(NestedOneOfInAllOf{})

	// The merged struct must carry the flat property from the outer allOf member.
	_, hasAFoo := typ.FieldByName("AFoo")
	assert.True(t, hasAFoo, "NestedOneOfInAllOf should have AFoo field")

	// The union field signals that the oneOf was preserved.
	_, hasUnion := typ.FieldByName("union")
	assert.True(t, hasUnion, "NestedOneOfInAllOf should have a union field (oneOf was dropped)")

	// The two oneOf variant types must be generated.
	_ = NestedOneOfInAllOf0{}
	_ = NestedOneOfInAllOf1{}
}

// issue #1219: test additionalProperties merge-precedence rules in allOf.
// In oapi-codegen, an unspecified additionalProperties is treated as false
// (unlike the OpenAPI specification default of true), so "default" and
// explicitly-false cases are handled separately.
func TestIssue1219(t *testing.T) {
	var exist bool

	// When both schemas have additionalProperties: true, the merged schema must have
	// additionalProperties: true (map[string]any).
	assert.IsType(t, map[string]any{}, MergeWithAnyWithAny{}.AdditionalProperties)

	// When one schema has additionalProperties: true and the other specifies a sub-schema,
	// the merged schema uses the sub-schema (the more specific wins).
	assert.IsType(t, map[string]string{}, MergeWithAnyWithString{}.AdditionalProperties)
	assert.IsType(t, map[string]string{}, MergeWithStringWithAny{}.AdditionalProperties)

	// When one schema has additionalProperties: true and the other is unspecified,
	// the merged schema has additionalProperties: true (both treated as "true" per spec).
	assert.IsType(t, map[string]any{}, MergeWithAnyDefault{}.AdditionalProperties)
	assert.IsType(t, map[string]any{}, MergeDefaultWithAny{}.AdditionalProperties)

	// When one schema has additionalProperties: true and the other has false,
	// the merged schema must have no AdditionalProperties field (false wins).
	_, exist = reflect.TypeOf(MergeWithAnyWithout{}).FieldByName("AdditionalProperties")
	assert.False(t, exist)
	_, exist = reflect.TypeOf(MergeWithoutWithAny{}).FieldByName("AdditionalProperties")
	assert.False(t, exist)

	// When one schema specifies a sub-schema and the other is unspecified,
	// the merged schema uses the specified sub-schema.
	assert.IsType(t, map[string]string{}, MergeWithStringDefault{}.AdditionalProperties)
	assert.IsType(t, map[string]string{}, MergeDefaultWithString{}.AdditionalProperties)

	// When one schema specifies a sub-schema and the other has false,
	// the merged schema has no AdditionalProperties field (false wins).
	_, exist = reflect.TypeOf(MergeWithStringWithout{}).FieldByName("AdditionalProperties")
	assert.False(t, exist)
	_, exist = reflect.TypeOf(MergeWithoutWithString{}).FieldByName("AdditionalProperties")
	assert.False(t, exist)

	// When both schemas are unspecified, the merged schema has no AdditionalProperties
	// field (treated as unspecified for compatibility, even though spec says true).
	_, exist = reflect.TypeOf(MergeDefaultDefault{}).FieldByName("AdditionalProperties")
	assert.False(t, exist)

	// When one schema is unspecified and the other has false,
	// the merged schema has no AdditionalProperties field.
	_, exist = reflect.TypeOf(MergeDefaultWithout{}).FieldByName("AdditionalProperties")
	assert.False(t, exist)
	_, exist = reflect.TypeOf(MergeWithoutDefault{}).FieldByName("AdditionalProperties")
	assert.False(t, exist)

	// When both schemas have additionalProperties: false,
	// the merged schema has no AdditionalProperties field.
	_, exist = reflect.TypeOf(MergeWithoutWithout{}).FieldByName("AdditionalProperties")
	assert.False(t, exist)
}

// issue #2524: a type or format declared by only one allOf member must
// propagate to the merged schema instead of being silently dropped (type) or
// erroring (format), and member order must not change the generated shape.
// The checks are compile-time: each pair of order-reversed schemas must lower
// to the same alias.
func TestIssue2524(t *testing.T) {
	// A typeless properties-only member alongside `type: string` generates a
	// string (properties don't constrain non-object instances), not a struct.
	// The conversions fail to compile if either shape regresses to a struct.
	assert.IsType(t, "", NewMergeTypeFromSecondMember("s"))
	assert.IsType(t, "", NewMergeTypeFromFirstMember("s"))

	// A 3.1 multi-type union member lowers to `any` in both orders.
	var union1 NewMergeUnionTypeFromSecondMember = "either a string"
	var union2 NewMergeUnionTypeFromFirstMember = 1.5
	assert.NotNil(t, union1)
	assert.NotNil(t, union2)

	// The decorator idiom over a format-carrying scalar propagates the format
	// (previously a hard "can not merge incompatible formats" error). The
	// composite literals fail to compile if the aliases regress to string.
	assert.IsType(t, openapi_types.UUID{}, NewMergeFormatFromRef{})
	assert.IsType(t, openapi_types.UUID{}, NewMergeFormatFromRefReversed{})
}
