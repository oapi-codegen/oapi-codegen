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

func TestProperty_ZeroValueIsNil(t *testing.T) {
	newType := func(typ string) *openapi3.Types {
		return &openapi3.Types{typ}
	}

	tests := []struct {
		name        string
		oapiSchema  *openapi3.Schema
		goType      string
		expectIsNil bool
	}{
		{
			name:        "when an array, returns true",
			oapiSchema:  &openapi3.Schema{Type: newType("array")},
			expectIsNil: true,
		},
		{
			name:        "when an object, returns false",
			oapiSchema:  &openapi3.Schema{Type: newType("object")},
			expectIsNil: false,
		},
		{
			name:        "when an object rendered as a map, returns true",
			oapiSchema:  &openapi3.Schema{Type: newType("object")},
			goType:      "map[string]string",
			expectIsNil: true,
		},
		{
			name:        "when a string, returns false",
			oapiSchema:  &openapi3.Schema{Type: newType("string")},
			expectIsNil: false,
		},
		{
			name:        "when an integer, returns false",
			oapiSchema:  &openapi3.Schema{Type: newType("integer")},
			expectIsNil: false,
		},
		{
			name:        "when a number, returns false",
			oapiSchema:  &openapi3.Schema{Type: newType("number")},
			expectIsNil: false,
		},
		{
			name:        "when OAPISchema is nil, returns false",
			oapiSchema:  nil,
			expectIsNil: false,
		},
		{
			name:        "when OAPISchema is zero value, returns false",
			oapiSchema:  &openapi3.Schema{},
			expectIsNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop := Property{
				Schema: Schema{
					OAPISchema: tt.oapiSchema,
					GoType:     tt.goType,
				},
			}
			if tt.expectIsNil {
				require.True(t, prop.ZeroValueIsNil())
			} else {
				require.False(t, prop.ZeroValueIsNil())
			}
		})
	}
}

// A bare OpenAPI 3.1 `type: "null"` schema validates exactly the JSON
// value null. Go has no such type, so it maps to `any`, with the optional
// pointer skipped (nil is already `any`'s zero value). Issue #2430.
func TestOapiSchemaToGoType_NullType(t *testing.T) {
	schema := &openapi3.Schema{Type: &openapi3.Types{"null"}}
	var out Schema
	require.NoError(t, oapiSchemaToGoType(schema, []string{"Challenger"}, &out))
	assert.Equal(t, "any", out.GoType)
	assert.True(t, out.SkipOptionalPointer)
	assert.True(t, out.DefineViaAlias)
}

// enumViaOneOfConstType is the type-inference core behind a no-outer-type
// enum-via-oneOf. It maps a branch's `const` value to a scalar family and
// rejects anything that cannot form a typed Go enum.
func TestEnumViaOneOfConstType(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string // expected family on success, "" = must not match
		ok   bool
	}{
		{name: "string", in: "available", want: "string", ok: true},
		{name: "empty string", in: "", want: "string", ok: true},
		{name: "whole float64 (kin-openapi yaml number)", in: float64(8080), want: "integer", ok: true},
		{name: "zero float64", in: float64(0), want: "integer", ok: true},
		{name: "float64 int64", in: float64(2), want: "integer", ok: true},
		{name: "int", in: 42, want: "integer", ok: true},
		{name: "int64", in: int64(7), want: "integer", ok: true},
		{name: "fractional float", in: 1.5, want: "", ok: false},
		{name: "bool", in: true, want: "", ok: false},
		{name: "nil", in: nil, want: "", ok: false},
		{name: "map (an object)", in: map[string]any{"a": 1}, want: "", ok: false},
		{name: "slice (an array)", in: []any{1, 2}, want: "", ok: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fam, ok := enumViaOneOfConstType(tc.in)
			assert.Equal(t, tc.ok, ok, "enumViaOneOfConstType(%v) ok mismatch", tc.in)
			assert.Equal(t, tc.want, fam, "family mismatch")
		})
	}
}
