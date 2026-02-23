package issue2091

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// via gpt-4.1 (GitHub Copilot)
func hasOmitEmptyTag(field reflect.StructField) bool {
	tag := field.Tag.Get("json")
	parts := strings.Split(tag, ",")
	return slices.Contains(parts[1:], "omitempty")
}

// via gpt-4.1 (GitHub Copilot)
func TestTypeWithNullableHasOmitEmpty(t *testing.T) {
	typ := reflect.TypeOf(TypeWithNullable{})

	field, ok := typ.FieldByName("Name")
	require.True(t, ok)

	assert.True(t, hasOmitEmptyTag(field), "newfield should have `omitempty` set, given the usage of `nullable: true`")
}
