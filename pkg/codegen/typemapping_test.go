package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatMapping_Resolve(t *testing.T) {
	fm := FormatMapping{
		Default: SimpleTypeSpec{Type: "int"},
		Formats: map[string]SimpleTypeSpec{
			"int32": {Type: "int32"},
			"int64": {Type: "int64"},
		},
	}

	assert.Equal(t, "int", fm.Resolve("").Type)
	assert.Equal(t, "int32", fm.Resolve("int32").Type)
	assert.Equal(t, "int64", fm.Resolve("int64").Type)
	assert.Equal(t, "int", fm.Resolve("unknown-format").Type)
}

func TestTypeMapping_Merge(t *testing.T) {
	base := DefaultTypeMapping

	user := TypeMapping{
		Integer: FormatMapping{
			Default: SimpleTypeSpec{Type: "int64"},
		},
		String: FormatMapping{
			Formats: map[string]SimpleTypeSpec{
				"date-time": {Type: "civil.DateTime", Import: "cloud.google.com/go/civil"},
			},
		},
	}

	merged := base.Merge(user)

	// Integer default overridden
	assert.Equal(t, "int64", merged.Integer.Default.Type)
	// Integer formats still inherited from base
	assert.Equal(t, "int32", merged.Integer.Formats["int32"].Type)

	// String date-time overridden
	assert.Equal(t, "civil.DateTime", merged.String.Formats["date-time"].Type)
	assert.Equal(t, "cloud.google.com/go/civil", merged.String.Formats["date-time"].Import)
	// String default still inherited from base
	assert.Equal(t, "string", merged.String.Default.Type)
	// Other string formats still inherited
	assert.Equal(t, "openapi_types.UUID", merged.String.Formats["uuid"].Type)

	// Number and Boolean unchanged
	assert.Equal(t, "float32", merged.Number.Default.Type)
	assert.Equal(t, "bool", merged.Boolean.Default.Type)
}

func TestDefaultTypeMapping_Completeness(t *testing.T) {
	// Verify all the default mappings match what was previously hardcoded
	dm := DefaultTypeMapping

	// Integer
	assert.Equal(t, "int", dm.Integer.Resolve("").Type)
	assert.Equal(t, "int32", dm.Integer.Resolve("int32").Type)
	assert.Equal(t, "int64", dm.Integer.Resolve("int64").Type)
	assert.Equal(t, "uint32", dm.Integer.Resolve("uint32").Type)
	assert.Equal(t, "int", dm.Integer.Resolve("unknown").Type)

	// Number
	assert.Equal(t, "float32", dm.Number.Resolve("").Type)
	assert.Equal(t, "float32", dm.Number.Resolve("float").Type)
	assert.Equal(t, "float64", dm.Number.Resolve("double").Type)
	assert.Equal(t, "float32", dm.Number.Resolve("unknown").Type)

	// Boolean
	assert.Equal(t, "bool", dm.Boolean.Resolve("").Type)

	// String
	assert.Equal(t, "string", dm.String.Resolve("").Type)
	assert.Equal(t, "[]byte", dm.String.Resolve("byte").Type)
	assert.Equal(t, "openapi_types.Email", dm.String.Resolve("email").Type)
	assert.Equal(t, "openapi_types.Date", dm.String.Resolve("date").Type)
	assert.Equal(t, "time.Time", dm.String.Resolve("date-time").Type)
	assert.Equal(t, "json.RawMessage", dm.String.Resolve("json").Type)
	assert.Equal(t, "openapi_types.UUID", dm.String.Resolve("uuid").Type)
	assert.Equal(t, "openapi_types.File", dm.String.Resolve("binary").Type)
	assert.Equal(t, "string", dm.String.Resolve("unknown").Type)
}
