package issue1219_test

import (
	"reflect"
	"testing"

	issue1219 "github.com/deepmap/oapi-codegen/internal/test/issues/issue-1219"
	"github.com/stretchr/testify/assert"
)

// Test treatment additionalProperties in mergeOpenapiSchemas()
func TestIssue1219(t *testing.T) {
	var exist bool
	// In the current oapi-codegen, additionalProperties is treated as `false’ unlike the openapi specification
	// when it is not specified, so the case where additionalProperties is unspecified and
	// the case where it is explicitly `true’ are treated separately.

	// When properties "additionalProperties" in both schemas are explicitly "true",
	// the property "additionalProperties" in merged schema must be "true".
	assert.IsType(t, map[string]interface{}{}, issue1219.MergeWithAnyWithAny{}.AdditionalProperties)
	// Old behavior: generate fail

	// When property "additionalProperties" in one schema is explicitly "true" and
	// property "additionalProperties" in another schema specifies sub-schema,
	// the property "additionalProperties" in merged schema must use specified sub-schema in later source schema.
	assert.IsType(t, map[string]string{}, issue1219.MergeWithAnyWithString{}.AdditionalProperties)
	assert.IsType(t, map[string]string{}, issue1219.MergeWithStringWithAny{}.AdditionalProperties)
	// Old behavior: generate fail

	// When property "additionalProperties" in one schema is explicitly "true" and
	// property "additionalProperties" in another schema is not specified,
	// the property "additionalProperties" in merged schema must be "true".
	// This is because properties "additionalProperties" in both schemas are treated as "true" in the openapi specification.
	assert.IsType(t, map[string]interface{}{}, issue1219.MergeWithAnyDefault{}.AdditionalProperties)
	assert.IsType(t, map[string]interface{}{}, issue1219.MergeDefaultWithAny{}.AdditionalProperties)
	// Old behavior: additionalProperties is treated as unspecified

	// When property "additionalProperties" in one schema is explicitly "true" and
	// property "additionalProperties" in another schema is "false",
	// the property "additionalProperties" in merged schema must be "false".
	_, exist = reflect.TypeOf(issue1219.MergeWithAnyWithout{}).FieldByName("AdditionalProperties")
	assert.False(t, exist)
	_, exist = reflect.TypeOf(issue1219.MergeWithoutWithAny{}).FieldByName("AdditionalProperties")
	assert.False(t, exist)
	// Old behavior: additionalProperties is treated as unspecified

	// When properties "additionalProperties" in both schemas specify sub-schema,
	// sub-schemas must be merged.
	// But this is not yet implemented.
	// issue1219.MergeWithStringWithString{}

	// When property "additionalProperties" in one schema specifies sub-schema and
	// property "additionalProperties" in another schema is not specified,
	// the property "additionalProperties" in merged schema must use specified sub-schema.
	assert.IsType(t, map[string]string{}, issue1219.MergeWithStringDefault{}.AdditionalProperties)
	assert.IsType(t, map[string]string{}, issue1219.MergeDefaultWithString{}.AdditionalProperties)

	// When property "additionalProperties" in one schema specifies sub-schema and
	// property "additionalProperties" in another schema is "false",
	// the property "additionalProperties" in merged schema must be "false".
	_, exist = reflect.TypeOf(issue1219.MergeWithStringWithout{}).FieldByName("AdditionalProperties")
	assert.False(t, exist)
	_, exist = reflect.TypeOf(issue1219.MergeWithoutWithString{}).FieldByName("AdditionalProperties")
	assert.False(t, exist)
	// Old behavior: additionalProperties use specified sub-schema.

	// When properties "additionalProperties" in both schemas are not specified,
	// the property "additionalProperties" in merged schema must be "true" in the openapi specification.
	// But to avoid compatibility issue, property "additionalProperties" in merged schema is treated as unspecified.
	_, exist = reflect.TypeOf(issue1219.MergeDefaultDefault{}).FieldByName("AdditionalProperties")
	assert.False(t, exist)

	// When property "additionalProperties" in one schema is not specified and
	// property "additionalProperties" in another schema is "false",
	// the property "additionalProperties" in merged schema must be "false".
	_, exist = reflect.TypeOf(issue1219.MergeDefaultWithout{}).FieldByName("AdditionalProperties")
	assert.False(t, exist)
	_, exist = reflect.TypeOf(issue1219.MergeWithoutDefault{}).FieldByName("AdditionalProperties")
	assert.False(t, exist)

	// When properties "additionalProperties" in both schemas are "false",
	// the property "additionalProperties" in merged schema must be "false".
	_, exist = reflect.TypeOf(issue1219.MergeWithoutWithout{}).FieldByName("AdditionalProperties")
	assert.False(t, exist)
}
