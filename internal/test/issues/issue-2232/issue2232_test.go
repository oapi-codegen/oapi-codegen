package issue2232

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtraTagsOnQueryParams verifies that x-oapi-codegen-extra-tags is applied
// to query parameter struct fields regardless of whether the extension is placed
// at the parameter level or at the schema level within the parameter.
// This is a regression test for https://github.com/oapi-codegen/oapi-codegen/issues/2232
func TestExtraTagsOnQueryParams(t *testing.T) {
	paramType := reflect.TypeOf(GetEndpointParams{})

	t.Run("parameter-level extension", func(t *testing.T) {
		field, ok := paramType.FieldByName("EnvParamLevel")
		require.True(t, ok, "field EnvParamLevel should exist")

		assert.Equal(t, `required,oneof=dev live`, field.Tag.Get("validate"),
			"x-oapi-codegen-extra-tags at parameter level should produce validate tag")
	})

	t.Run("schema-level extension", func(t *testing.T) {
		field, ok := paramType.FieldByName("EnvSchemaLevel")
		require.True(t, ok, "field EnvSchemaLevel should exist")

		assert.Equal(t, `required,oneof=dev live`, field.Tag.Get("validate"),
			"x-oapi-codegen-extra-tags at schema level within a parameter should produce validate tag")
	})

	t.Run("schema-level extension on optional param", func(t *testing.T) {
		field, ok := paramType.FieldByName("Limit")
		require.True(t, ok, "field Limit should exist")

		assert.Equal(t, `min=0,max=100`, field.Tag.Get("validate"),
			"x-oapi-codegen-extra-tags at schema level within an optional parameter should produce validate tag")
	})
}
