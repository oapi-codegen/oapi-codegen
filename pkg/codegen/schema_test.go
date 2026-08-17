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
		{name: "float64 overflow int64 max", in: float64(9223372036854775808), want: "", ok: false},
		{name: "float64 overflow int64 negative", in: -9.3e18, want: "", ok: false},
		{name: "uint64 overflow int64 max", in: uint64(18446744073709551615), want: "", ok: false},
		{name: "uint64 within int64", in: uint64(9223372036854775807), want: "integer", ok: true},
		{name: "int64 min boundary", in: int64(-9223372036854775808), want: "integer", ok: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fam, ok := enumViaOneOfConstType(tc.in)
			assert.Equal(t, tc.ok, ok, "enumViaOneOfConstType(%v) ok mismatch", tc.in)
			assert.Equal(t, tc.want, fam, "family mismatch")
		})
	}
}

// An OpenAPI 3.1 `type` list holding more than one type after "null" is
// stripped is a multi-type union. Go has no type accepting exactly those
// types, so the union maps to `any`.
func TestOapiSchemaToGoType_MultiTypeUnion(t *testing.T) {
	for _, tc := range []struct {
		name     string
		types    openapi3.Types
		wantErr  bool
		wantType string
		wantSkip bool
	}{
		{
			name:     "three scalar types",
			types:    openapi3.Types{"string", "number", "boolean"},
			wantType: "any",
			wantSkip: true,
		},
		{
			name:     "array in a union does not take the array branch",
			types:    openapi3.Types{"array", "string"},
			wantType: "any",
			wantSkip: true,
		},
		{
			name:     "union alongside null",
			types:    openapi3.Types{"string", "number", "null"},
			wantType: "any",
			wantSkip: true,
		},
		{
			name:     "single type alongside null is not a union",
			types:    openapi3.Types{"string", "null"},
			wantType: "string",
			wantSkip: false,
		},
		{
			name:    "misspelled type name is still an error",
			types:   openapi3.Types{"strng", "number"},
			wantErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prev := globalState
			t.Cleanup(func() { globalState = prev })
			globalState.is31 = true
			globalState.typeMapping = DefaultTypeMapping

			var out Schema
			err := oapiSchemaToGoType(&openapi3.Schema{Type: &tc.types}, []string{"Value"}, &out)
			if tc.wantErr {
				assert.ErrorContains(t, err, "unhandled Schema type")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantType, out.GoType)
			assert.Equal(t, tc.wantSkip, out.SkipOptionalPointer)
			assert.True(t, out.DefineViaAlias)
		})
	}
}

// An integer enum-via-oneOf whose const exceeds the Go int64 range must not
// be emitted as a typed enum: the verbatim value would not fit the target
// integer type and the generated source would fail to format. Instead
// detectEnumViaOneOf must fall through (ok=false) so the union path is used.
// Issue: overflowing constants were treated as valid integers.
func TestEnumViaOneOfOverflowFallsThrough(t *testing.T) {
	prev := globalState
	t.Cleanup(func() { globalState = prev })
	globalState.is31 = true
	globalState.options.OutputOptions.SkipEnumViaOneOf = false

	schema := &openapi3.Schema{
		Type: nil, // no outer type -> family inferred from const values
		OneOf: openapi3.SchemaRefs{
			{Value: &openapi3.Schema{Title: "Enormous", Const: float64(9223372036854775808)}}, // 2^63, overflows int64
		},
	}

	items, enumType, ok := detectEnumViaOneOf(schema)
	assert.False(t, ok, "overflowing const must fall through, not form an enum")
	assert.Nil(t, items)
	assert.Nil(t, enumType)
}

// A list-valued `type` is only legal syntax in OpenAPI 3.1. A 3.0 document
// carrying one is malformed, so it keeps failing generation instead of
// picking up the 3.1 union mapping.
func TestOapiSchemaToGoType_MultiTypeUnionRequires31(t *testing.T) {
	prev := globalState
	t.Cleanup(func() { globalState = prev })
	globalState.is31 = false
	globalState.typeMapping = DefaultTypeMapping

	var out Schema
	err := oapiSchemaToGoType(&openapi3.Schema{Type: &openapi3.Types{"string", "number"}}, []string{"Value"}, &out)
	assert.ErrorContains(t, err, "unhandled Schema type")
}

// schemaUnionTypes feeds the Types bind option emitted for union
// parameters: the member list with "null" stripped for genuine 3.1
// multi-type unions, nil for everything else (single types bind through
// their concrete Go type, and a list is not legal syntax in 3.0).
func TestSchemaUnionTypes(t *testing.T) {
	for _, tc := range []struct {
		name string
		is31 bool
		in   *openapi3.Types
		want []string
	}{
		{
			name: "union",
			is31: true,
			in:   &openapi3.Types{"string", "integer"},
			want: []string{"string", "integer"},
		},
		{
			name: "nullable union strips the null marker",
			is31: true,
			in:   &openapi3.Types{"integer", "string", "null"},
			want: []string{"integer", "string"},
		},
		{
			name: "nullable single type is not a union",
			is31: true,
			in:   &openapi3.Types{"string", "null"},
			want: nil,
		},
		{
			name: "single type is not a union",
			is31: true,
			in:   &openapi3.Types{"string"},
			want: nil,
		},
		{
			name: "nil types",
			is31: true,
			in:   nil,
			want: nil,
		},
		{
			name: "3.0 documents never produce a union",
			is31: false,
			in:   &openapi3.Types{"string", "integer"},
			want: nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prev := globalState
			t.Cleanup(func() { globalState = prev })
			globalState.is31 = tc.is31

			assert.Equal(t, tc.want, schemaUnionTypes(tc.in))
		})
	}
}

// SchemaType feeds the Type bind option: the single declared type with the
// 3.1 "null" marker stripped, or "" for unions (carried by SchemaTypes) and
// typeless schemas. The previous first-entry behavior emitted "null" for
// ["null", "string"] and a lone member for unions.
func TestParameterDefinitionSchemaType(t *testing.T) {
	prev := globalState
	t.Cleanup(func() { globalState = prev })
	globalState.is31 = true

	paramWithTypes := func(types *openapi3.Types) *ParameterDefinition {
		return &ParameterDefinition{
			Spec: &openapi3.Parameter{
				Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{Type: types}},
			},
		}
	}

	assert.Equal(t, "string", paramWithTypes(&openapi3.Types{"string"}).SchemaType())
	assert.Equal(t, "string", paramWithTypes(&openapi3.Types{"null", "string"}).SchemaType(),
		"null is the nullability marker, not the parameter's type")
	assert.Equal(t, "", paramWithTypes(&openapi3.Types{"string", "integer"}).SchemaType(),
		"unions carry no single Type; SchemaTypes has the members")
	assert.Equal(t, []string{"string", "integer"}, paramWithTypes(&openapi3.Types{"string", "integer", "null"}).SchemaTypes())
	assert.Nil(t, paramWithTypes(&openapi3.Types{"string"}).SchemaTypes())
}
