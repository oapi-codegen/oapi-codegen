package codegen

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProperty_GoTypeDef(t *testing.T) {
	type fields struct {
		GlobalStateDisableRequiredReadOnlyAsPointer bool
		Schema                                      Schema
		Required                                    bool
		Nullable                                    bool
		ReadOnly                                    bool
		WriteOnly                                   bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			// When pointer is skipped by setting flag SkipOptionalPointer, the
			// flag will never be pointer irrespective of other flags.
			name: "Set skip optional pointer type for go type",
			fields: fields{
				Schema: Schema{
					SkipOptionalPointer: true,
					RefType:             "",
					GoType:              "int",
				},
			},
			want: "int",
		},

		{
			// if the field is optional, it will always be pointer irrespective of other
			// flags, given that pointer type is not skipped by setting SkipOptionalPointer
			// flag to true
			name: "When the field is optional",
			fields: fields{
				Schema: Schema{
					SkipOptionalPointer: false,
					RefType:             "",
					GoType:              "int",
				},
				Required: false,
			},
			want: "*int",
		},

		{
			// if the field(custom-type) is optional, it will NOT be a pointer if
			// SkipOptionalPointer flag is set to true
			name: "Set skip optional pointer type for ref type",
			fields: fields{
				Schema: Schema{
					SkipOptionalPointer: true,
					RefType:             "CustomType",
					GoType:              "int",
				},
				Required: false,
			},
			want: "CustomType",
		},

		// For the following test cases, SkipOptionalPointer flag is false.
		{
			name: "When field is required and not nullable",
			fields: fields{
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				Required: true,
				Nullable: false,
			},
			want: "int",
		},

		{
			name: "When field is required and nullable",
			fields: fields{
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				Required: true,
				Nullable: true,
			},
			want: "*int",
		},

		{
			name: "When field is optional and not nullable",
			fields: fields{
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				Required: false,
				Nullable: false,
			},
			want: "*int",
		},

		{
			name: "When field is optional and nullable",
			fields: fields{
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				Required: false,
				Nullable: true,
			},
			want: "*int",
		},

		// Following tests cases for non-nullable and required; and skip pointer is not opted
		{
			name: "When field is readOnly it will always be pointer",
			fields: fields{
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				ReadOnly: true,
				Required: true,
			},
			want: "*int",
		},

		{
			name: "When field is readOnly and read only pointer disabled",
			fields: fields{
				GlobalStateDisableRequiredReadOnlyAsPointer: true,
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				ReadOnly: true,
				Required: true,
			},
			want: "int",
		},

		{
			name: "When field is readOnly and optional",
			fields: fields{
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				ReadOnly: true,
				Required: false,
			},
			want: "*int",
		},
		{
			name: "When field is readOnly and optional and read only pointer disabled",
			fields: fields{
				GlobalStateDisableRequiredReadOnlyAsPointer: true,
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				ReadOnly: true,
				Required: false,
			},
			want: "*int",
		},

		// When field is write only, it will always be pointer unless pointer is
		// skipped by setting SkipOptionalPointer flag
		{
			name: "When field is write only and read only pointer disabled",
			fields: fields{
				GlobalStateDisableRequiredReadOnlyAsPointer: true,
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				WriteOnly: true,
			},
			want: "*int",
		},

		{
			name: "When field is write only and read only pointer enabled",
			fields: fields{
				GlobalStateDisableRequiredReadOnlyAsPointer: false,
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				WriteOnly: true,
			},
			want: "*int",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globalState.options.Compatibility.DisableRequiredReadOnlyAsPointer = tt.fields.GlobalStateDisableRequiredReadOnlyAsPointer
			p := Property{
				Schema:    tt.fields.Schema,
				Required:  tt.fields.Required,
				Nullable:  tt.fields.Nullable,
				ReadOnly:  tt.fields.ReadOnly,
				WriteOnly: tt.fields.WriteOnly,
			}
			assert.Equal(t, tt.want, p.GoTypeDef())
		})
	}
}

func TestProperty_GoTypeDef_nullable(t *testing.T) {
	type fields struct {
		GlobalStateDisableRequiredReadOnlyAsPointer bool
		GlobalStateNullableType                     bool
		Schema                                      Schema
		Required                                    bool
		Nullable                                    bool
		ReadOnly                                    bool
		WriteOnly                                   bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			// Field not nullable.
			// When pointer is skipped by setting flag SkipOptionalPointer, the
			// flag will never be pointer irrespective of other flags.
			name: "Set skip optional pointer type for go type",
			fields: fields{
				GlobalStateNullableType: true,
				Schema: Schema{
					SkipOptionalPointer: true,
					RefType:             "",
					GoType:              "int",
				},
			},
			want: "int",
		},

		{
			// Field not nullable.
			// if the field is optional, it will always be pointer irrespective of other
			// flags, given that pointer type is not skipped by setting SkipOptionalPointer
			// flag to true
			name: "When the field is optional",
			fields: fields{
				GlobalStateNullableType: true,
				Schema: Schema{
					SkipOptionalPointer: false,
					RefType:             "",
					GoType:              "int",
				},
				Required: false,
			},
			want: "*int",
		},

		{
			// Field not nullable.
			// if the field(custom type) is optional, it will NOT be a pointer if
			// SkipOptionalPointer flag is set to true
			name: "Set skip optional pointer type for ref type",
			fields: fields{
				GlobalStateNullableType: true,
				Schema: Schema{
					SkipOptionalPointer: true,
					RefType:             "CustomType",
					GoType:              "int",
				},
				Required: false,
			},
			want: "CustomType",
		},

		// Field not nullable.
		// For the following test case, SkipOptionalPointer flag is false.
		{
			name: "When field is required and not nullable",
			fields: fields{
				GlobalStateNullableType: true,
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				Required: true,
				Nullable: false,
			},
			want: "int",
		},

		{
			name: "When field is required and nullable",
			fields: fields{
				GlobalStateNullableType: true,
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				Required: true,
				Nullable: true,
			},
			want: "nullable.Nullable[int]",
		},

		{
			name: "When field is optional and not nullable",
			fields: fields{
				GlobalStateNullableType: true,
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				Required: false,
				Nullable: false,
			},
			want: "*int",
		},

		{
			name: "When field is optional and nullable",
			fields: fields{
				GlobalStateNullableType: true,
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				Required: false,
				Nullable: true,
			},
			want: "nullable.Nullable[int]",
		},

		{
			name: "When field is readOnly, non-nullable and required and skip pointer is not opted",
			fields: fields{
				GlobalStateNullableType: true,
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				ReadOnly: true,
				Required: true,
			},
			want: "*int",
		},

		{
			name: "When field is readOnly, required, non-nullable and read only pointer disabled",
			fields: fields{
				GlobalStateNullableType:                     true,
				GlobalStateDisableRequiredReadOnlyAsPointer: true,
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				ReadOnly: true,
				Required: true,
			},
			want: "int",
		},

		{
			name: "When field is readOnly, optional and non nullable",
			fields: fields{
				GlobalStateNullableType: true,
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				ReadOnly: true,
				Required: false,
			},
			want: "*int",
		},
		{
			name: "When field is readOnly and optional and read only pointer disabled",
			fields: fields{
				GlobalStateNullableType:                     true,
				GlobalStateDisableRequiredReadOnlyAsPointer: true,
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				ReadOnly: true,
				Required: false,
			},
			want: "*int",
		},

		{
			name: "When field is write only and non nullable",
			fields: fields{
				GlobalStateNullableType:                     true,
				GlobalStateDisableRequiredReadOnlyAsPointer: true,
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				WriteOnly: true,
			},
			want: "*int",
		},

		{
			name: "When field is write only and nullable",
			fields: fields{
				GlobalStateNullableType:                     true,
				GlobalStateDisableRequiredReadOnlyAsPointer: true,
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				WriteOnly: true,
				Nullable:  true,
			},
			want: "nullable.Nullable[int]",
		},

		{
			name: "When field is write only, nullable and read only pointer enabled",
			fields: fields{
				GlobalStateNullableType: true,
				Schema: Schema{
					SkipOptionalPointer: false,
					GoType:              "int",
				},
				WriteOnly: true,
				Nullable:  true,
			},
			want: "nullable.Nullable[int]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globalState.options.Compatibility.DisableRequiredReadOnlyAsPointer = tt.fields.GlobalStateDisableRequiredReadOnlyAsPointer
			globalState.options.OutputOptions.NullableType = tt.fields.GlobalStateNullableType
			p := Property{
				Schema:    tt.fields.Schema,
				Required:  tt.fields.Required,
				Nullable:  tt.fields.Nullable,
				ReadOnly:  tt.fields.ReadOnly,
				WriteOnly: tt.fields.WriteOnly,
			}
			assert.Equal(t, tt.want, p.GoTypeDef())
		})
	}
}

// testSchemaToGoTypeAlias is a helper function to test the schema to go type conversion.
func testSchemaToGoTypeAlias(t *testing.T, schemaType, schemaFormat, goType string) {
	t.Helper()

	testSchemaToGoTypeWant(t, schemaType, schemaFormat, Schema{
		GoType:         goType,
		DefineViaAlias: true,
	})
}

// testSchemaToGoTypeWant is a helper function to test the schema to go type conversion.
// It asserts that the schema type and format are converted to the expected go type.
func testSchemaToGoTypeWant(t *testing.T, schemaType, schemaFormat string, expected Schema) {
	t.Helper()

	outSchema := &Schema{}
	err := oapiSchemaToGoType(
		&openapi3.Schema{
			Type:   &openapi3.Types{schemaType},
			Format: schemaFormat,
		},
		nil, outSchema,
	)
	require.NoError(t, err)
	require.Equal(t, expected, *outSchema)
}

// testSchemaToGoTypeError is a helper function to test the schema to go type conversion.
// It asserts that the schema type and format result in to the expected error.
func testSchemaToGoTypeError(t *testing.T, schemaType, schemaFormat, expectedError string) {
	t.Helper()

	outSchema := &Schema{}
	err := oapiSchemaToGoType(
		&openapi3.Schema{
			Type:   &openapi3.Types{schemaType},
			Format: schemaFormat,
		},
		nil, outSchema,
	)
	require.EqualError(t, err, expectedError)
}

// testSchemaToGoTypeOverride is a helper function to test the schema to go type conversion with overrides.
// It asserts that the schema type and format are converted to the expected go type using the given type mapping.
func testSchemaToGoTypeOverride(t *testing.T, schemaType, schemaFormat, goType string, typeMapping TypeMapping, skipOptionalPointer, defineViaAlias bool) {
	t.Helper()

	old := globalState.options.OutputOptions.TypeMappings
	globalState.options.OutputOptions.TypeMappings = map[string]TypeMapping{
		schemaType + "-" + schemaFormat: typeMapping,
	}
	t.Cleanup(func() {
		globalState.options.OutputOptions.TypeMappings = old
	})

	testSchemaToGoTypeWant(t, schemaType, schemaFormat, Schema{
		GoType:              goType,
		SkipOptionalPointer: skipOptionalPointer,
		DefineViaAlias:      defineViaAlias,
	})
}

func Test_oapiSchemaToGoType(t *testing.T) {
	// Integers.
	t.Run("integer-int64", func(t *testing.T) {
		testSchemaToGoTypeAlias(t, "integer", "int64", "int64")
	})

	t.Run("integer-int32", func(t *testing.T) {
		testSchemaToGoTypeAlias(t, "integer", "int32", "int32")
	})

	t.Run("integer-int16", func(t *testing.T) {
		testSchemaToGoTypeAlias(t, "integer", "int16", "int16")
	})

	t.Run("integer-int8", func(t *testing.T) {
		testSchemaToGoTypeAlias(t, "integer", "int8", "int8")
	})

	t.Run("integer-uint64", func(t *testing.T) {
		testSchemaToGoTypeAlias(t, "integer", "uint64", "uint64")
	})

	t.Run("integer-uint32", func(t *testing.T) {
		testSchemaToGoTypeAlias(t, "integer", "uint32", "uint32")
	})

	t.Run("integer-uint16", func(t *testing.T) {
		testSchemaToGoTypeAlias(t, "integer", "uint16", "uint16")
	})

	t.Run("integer-uint8", func(t *testing.T) {
		testSchemaToGoTypeAlias(t, "integer", "uint8", "uint8")
	})

	t.Run("integer-uint", func(t *testing.T) {
		testSchemaToGoTypeAlias(t, "integer", "uint", "uint")
	})

	t.Run("integer-unknown", func(t *testing.T) {
		testSchemaToGoTypeAlias(t, "integer", "unknown", "int")
	})

	// Numbers.
	t.Run("number-double", func(t *testing.T) {
		testSchemaToGoTypeAlias(t, "number", "double", "float64")
	})

	t.Run("number-float", func(t *testing.T) {
		testSchemaToGoTypeAlias(t, "number", "float", "float32")
	})

	t.Run("number-unknown", func(t *testing.T) {
		testSchemaToGoTypeError(t, "number", "unknown", "invalid number format: unknown")
	})

	// Booleans.
	t.Run("boolean", func(t *testing.T) {
		testSchemaToGoTypeAlias(t, "boolean", "", "bool")
	})

	t.Run("boolean-unknown", func(t *testing.T) {
		testSchemaToGoTypeError(t, "boolean", "unknown", "invalid format (unknown) for boolean")
	})

	// Strings.
	t.Run("string-byte", func(t *testing.T) {
		testSchemaToGoTypeAlias(t, "string", "byte", "[]byte")
	})

	t.Run("string-email", func(t *testing.T) {
		testSchemaToGoTypeAlias(t, "string", "email", "openapi_types.Email")
	})

	t.Run("string-date", func(t *testing.T) {
		testSchemaToGoTypeAlias(t, "string", "date", "openapi_types.Date")
	})

	t.Run("string-date-time", func(t *testing.T) {
		testSchemaToGoTypeAlias(t, "string", "date-time", "time.Time")
	})

	t.Run("string-json", func(t *testing.T) {
		testSchemaToGoTypeWant(t, "string", "json", Schema{
			GoType:              "json.RawMessage",
			DefineViaAlias:      true,
			SkipOptionalPointer: true,
		})
	})

	t.Run("string-binary", func(t *testing.T) {
		testSchemaToGoTypeAlias(t, "string", "binary", "openapi_types.File")
	})

	t.Run("string", func(t *testing.T) {
		testSchemaToGoTypeAlias(t, "string", "", "string")
	})

	t.Run("string-other", func(t *testing.T) {
		testSchemaToGoTypeAlias(t, "string", "other", "string")
	})

	// Overrides
	t.Run("override-all", func(t *testing.T) {
		testSchemaToGoTypeOverride(t,
			"string", "uuid", "mypkg.UUID", TypeMapping{
				GoType:              ptr("mypkg.UUID"),
				SkipOptionalPointer: ptr(true),
				DefineViaAlias:      ptr(false),
			}, true, false,
		)
	})

	t.Run("override-go-type", func(t *testing.T) {
		testSchemaToGoTypeOverride(t,
			"string", "uuid", "mypkg.UUID", TypeMapping{
				GoType: ptr("mypkg.UUID"),
			}, false, true,
		)
	})

	t.Run("override-skip", func(t *testing.T) {
		testSchemaToGoTypeOverride(t,
			"string", "uuid", "openapi_types.UUID", TypeMapping{
				SkipOptionalPointer: ptr(true),
			}, true, true,
		)
	})

	t.Run("override-alias", func(t *testing.T) {
		testSchemaToGoTypeOverride(t,
			"string", "uuid", "openapi_types.UUID", TypeMapping{
				DefineViaAlias: ptr(false),
			}, false, false,
		)
	})
}

// ptr is a helper which returns a pointer to the given value.
func ptr[T any](t T) *T {
	return &t
}
