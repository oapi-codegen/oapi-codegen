package codegen

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, "openapi_types.Duration", dm.String.Resolve("duration").Type)
	assert.Equal(t, "json.RawMessage", dm.String.Resolve("json").Type)
	assert.Equal(t, "openapi_types.UUID", dm.String.Resolve("uuid").Type)
	assert.Equal(t, "openapi_types.File", dm.String.Resolve("binary").Type)
	assert.Equal(t, "string", dm.String.Resolve("unknown").Type)
}

// TestDurationFormatMapping verifies that `type: string, format: duration`
// generates the openapi_types.Duration runtime type by default, and that the
// historical plain-string behavior is restored by explicitly mapping the
// format back to string via type-mapping.
// Please see https://github.com/oapi-codegen/oapi-codegen/issues/2456
func TestDurationFormatMapping(t *testing.T) {
	const spec = `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Durations
paths: {}
components:
  schemas:
    RetryPolicy:
      type: object
      required: [backoff]
      properties:
        backoff:
          type: string
          format: duration
`
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	opts := Configuration{
		PackageName:   "api",
		Generate:      GenerateOptions{Models: true},
		OutputOptions: OutputOptions{SkipPrune: true},
	}

	code, err := Generate(swagger, opts)
	require.NoError(t, err)
	assert.Contains(t, code, "Backoff openapi_types.Duration `json:\"backoff\"`")
	assert.Contains(t, code, `openapi_types "github.com/oapi-codegen/runtime/types"`)

	// The pre-duration behavior is a plain string; restore it by mapping
	// the format back explicitly.
	opts.OutputOptions.TypeMapping = &TypeMapping{
		String: FormatMapping{
			Formats: map[string]SimpleTypeSpec{
				"duration": {Type: "string"},
			},
		},
	}
	code, err = Generate(swagger, opts)
	require.NoError(t, err)
	assert.Contains(t, code, "Backoff string `json:\"backoff\"`")
	assert.NotContains(t, code, "openapi_types.Duration")
}
