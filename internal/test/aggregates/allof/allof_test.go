package aggregatesallof

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// issue #1219: test additionalProperties merge-precedence rules in allOf.
// In oapi-codegen, an unspecified additionalProperties is treated as false
// (unlike the OpenAPI specification default of true), so "default" and
// explicitly-false cases are handled separately.
func TestIssue1219(t *testing.T) {
	var exist bool

	// When both schemas have additionalProperties: true, the merged schema must have
	// additionalProperties: true (map[string]interface{}).
	assert.IsType(t, map[string]interface{}{}, MergeWithAnyWithAny{}.AdditionalProperties)

	// When one schema has additionalProperties: true and the other specifies a sub-schema,
	// the merged schema uses the sub-schema (the more specific wins).
	assert.IsType(t, map[string]string{}, MergeWithAnyWithString{}.AdditionalProperties)
	assert.IsType(t, map[string]string{}, MergeWithStringWithAny{}.AdditionalProperties)

	// When one schema has additionalProperties: true and the other is unspecified,
	// the merged schema has additionalProperties: true (both treated as "true" per spec).
	assert.IsType(t, map[string]interface{}{}, MergeWithAnyDefault{}.AdditionalProperties)
	assert.IsType(t, map[string]interface{}{}, MergeDefaultWithAny{}.AdditionalProperties)

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
