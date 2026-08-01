package optionsstructtags

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fieldTag returns the struct tag of the named field.
func fieldTag(t *testing.T, typ reflect.Type, field string) reflect.StructTag {
	t.Helper()
	f, ok := typ.FieldByName(field)
	require.True(t, ok, "field %s not found on %s", field, typ)
	return f.Tag
}

func TestStructTags(t *testing.T) {
	pet := reflect.TypeOf(Pet{})

	t.Run("default json template is kept", func(t *testing.T) {
		assert.Equal(t, "name", fieldTag(t, pet, "Name").Get("json"))
		assert.Equal(t, "tag,omitempty", fieldTag(t, pet, "Tag").Get("json"))
	})

	t.Run("user yaml entry supersedes yaml-tags flag", func(t *testing.T) {
		// The overridden template has no omitempty, unlike the yaml-tags default.
		assert.Equal(t, "name", fieldTag(t, pet, "Name").Get("yaml"))
		assert.Equal(t, "tag", fieldTag(t, pet, "Tag").Get("yaml"))
	})

	t.Run("added tags are rendered", func(t *testing.T) {
		assert.Equal(t, "name", fieldTag(t, pet, "Name").Get("db"))
		assert.Equal(t, "required", fieldTag(t, pet, "Name").Get("validate"))
		// Empty render suppresses the tag entirely on optional fields.
		_, hasValidate := fieldTag(t, pet, "Tag").Lookup("validate")
		assert.False(t, hasValidate)
	})

	t.Run("form tag still gated on form-style parameters", func(t *testing.T) {
		params := reflect.TypeOf(ListPetsParams{})
		assert.Equal(t, "limit,omitempty", fieldTag(t, params, "Limit").Get("form"))
		// Schema fields are not form-bound, so no form tag is emitted.
		_, hasForm := fieldTag(t, pet, "Name").Lookup("form")
		assert.False(t, hasForm)
	})
}
