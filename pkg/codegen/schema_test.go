package codegen

import (
	"math"
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

func TestDescribeWithExamples(t *testing.T) {
	for _, tc := range []struct {
		name        string
		is31        bool
		description string
		schema      *openapi3.Schema
		want        string
	}{
		{
			name: "nil schema keeps description",
			want: "",
		},
		{
			name:        "3.0 singular example",
			description: "The widget name.",
			schema:      &openapi3.Schema{Example: "hammer"},
			want:        "The widget name.\n\nExample: hammer",
		},
		{
			name:   "3.0 singular example with no description",
			schema: &openapi3.Schema{Example: "hammer"},
			want:   "Example: hammer",
		},
		{
			name:        "3.0 ignores plural examples",
			description: "The widget name.",
			schema:      &openapi3.Schema{Examples: []any{"hammer"}},
			want:        "The widget name.",
		},
		{
			name:        "3.1 plural examples",
			is31:        true,
			description: "The widget name.",
			schema:      &openapi3.Schema{Examples: []any{"hammer", "wrench"}},
			want:        "The widget name.\n\nExamples: hammer, wrench",
		},
		{
			name:        "3.1 falls back to singular example",
			is31:        true,
			description: "The widget name.",
			schema:      &openapi3.Schema{Example: "hammer"},
			want:        "The widget name.\n\nExample: hammer",
		},
		{
			name:   "3.1 prefers plural examples over singular",
			is31:   true,
			schema: &openapi3.Schema{Example: "ignored", Examples: []any{"hammer"}},
			want:   "Examples: hammer",
		},
		{
			name:        "3.1 with neither keeps description",
			is31:        true,
			description: "The widget name.",
			schema:      &openapi3.Schema{},
			want:        "The widget name.",
		},
		{
			name:   "structured example is JSON encoded",
			schema: &openapi3.Schema{Example: map[string]any{"k": "v"}},
			want:   `Example: {"k":"v"}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			old := globalState.is31
			defer func() { globalState.is31 = old }()
			globalState.is31 = tc.is31

			assert.Equal(t, tc.want, describeWithExamples(tc.description, tc.schema))
		})
	}
}

// constScalarFamily is the type-inference core behind a no-outer-type
// enum-via-oneOf: it maps a branch's `const` to the OpenAPI scalar family it
// belongs to, and to the integer format wide enough to hold it. kin-openapi
// decodes every JSON number into a float64, so the numeric cases all arrive
// in that shape.
func TestConstScalarFamily(t *testing.T) {
	tests := []struct {
		name       string
		in         any
		wantFamily string // "" = must not match
		wantFormat string
		ok         bool
	}{
		{name: "string", in: "available", wantFamily: "string", ok: true},
		{name: "empty string", in: "", wantFamily: "string", ok: true},
		{name: "bool is its own family", in: true, wantFamily: "boolean", ok: true},
		{name: "fractional is a number, not an integer", in: 1.5, wantFamily: "number", ok: true},
		{name: "whole float64 (kin-openapi yaml number)", in: float64(8080), wantFamily: "integer", ok: true},
		{name: "zero", in: float64(0), wantFamily: "integer", ok: true},
		{name: "negative", in: float64(-7), wantFamily: "integer", ok: true},

		// int holds anything inside the int32 range on every build; past it
		// the enum widens to int64 so a 32-bit build stays correct.
		{name: "int32 max still fits plain int", in: float64(math.MaxInt32), wantFamily: "integer", ok: true},
		{name: "int32 min still fits plain int", in: float64(math.MinInt32), wantFamily: "integer", ok: true},
		{name: "past int32 widens to int64", in: float64(math.MaxInt32) + 1, wantFamily: "integer", wantFormat: "int64", ok: true},
		{name: "below int32 widens to int64", in: float64(math.MinInt32) - 1, wantFamily: "integer", wantFormat: "int64", ok: true},

		// Classification answers "what family", which stays true however big
		// the number is. Whether the value survived parsing is a separate
		// question, asked later by checkIntegerConstsExact.
		{name: "largest unambiguous integer is still just an integer", in: float64(1<<53) - 1, wantFamily: "integer", wantFormat: "int64", ok: true},
		{name: "2^53 classifies, exactness is not asked here", in: float64(1 << 53), wantFamily: "integer", wantFormat: "int64", ok: true},
		{name: "uint64 max classifies too", in: float64(18446744073709551615), wantFamily: "integer", wantFormat: "int64", ok: true},

		// Not scalars at all.
		{name: "nil", in: nil, ok: false},
		{name: "map (an object)", in: map[string]any{"a": 1}, ok: false},
		{name: "slice (an array)", in: []any{1, 2}, ok: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fam, format, ok := constScalarFamily(tc.in)
			assert.Equal(t, tc.ok, ok, "constScalarFamily(%v) ok mismatch", tc.in)
			assert.Equal(t, tc.wantFamily, fam, "family mismatch")
			assert.Equal(t, tc.wantFormat, format, "format mismatch")
		})
	}
}

// checkIntegerConstsExact is the guard that turns a value the parser mangled
// into an error. The bound excludes 2^53 itself: 2^53+1 arrives as exactly
// 2^53, so a value sitting on the boundary cannot be traced back to one
// integer.
func TestCheckIntegerConstsExact(t *testing.T) {
	refs := func(consts ...any) openapi3.SchemaRefs {
		out := make(openapi3.SchemaRefs, 0, len(consts))
		for _, c := range consts {
			out = append(out, &openapi3.SchemaRef{Value: &openapi3.Schema{Title: "T", Const: c}})
		}
		return out
	}

	for _, tc := range []struct {
		name    string
		in      openapi3.SchemaRefs
		wantErr bool
	}{
		{name: "small values", in: refs(float64(0), float64(-1), float64(8080))},
		{name: "largest unambiguous", in: refs(float64(1<<53) - 1)},
		{name: "smallest unambiguous", in: refs(-float64(1<<53) + 1)},
		{name: "past int32 but exact", in: refs(float64(5_000_000_000))},
		{name: "strings are not this check's business", in: refs("available", "sold")},
		{name: "a nil branch is someone else's problem", in: openapi3.SchemaRefs{nil}},

		{name: "2^53 is ambiguous with 2^53+1", in: refs(float64(1 << 53)), wantErr: true},
		{name: "2^53+1 arrives already rounded", in: refs(float64(9007199254740993)), wantErr: true},
		{name: "-2^53 is ambiguous too", in: refs(-float64(1 << 53)), wantErr: true},
		{name: "-(2^53+1) arrives already rounded", in: refs(-float64(9007199254740993)), wantErr: true},
		{name: "uint64 max", in: refs(float64(18446744073709551615)), wantErr: true},
		{name: "one bad value among good ones", in: refs(float64(1), float64(2), float64(1<<60)), wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := checkIntegerConstsExact(tc.in)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

// inferEnumViaOneOfType folds the per-branch families into the one type the
// whole enum takes, and widens the integer format to fit the largest value.
func TestInferEnumViaOneOfType(t *testing.T) {
	refs := func(consts ...any) openapi3.SchemaRefs {
		out := make(openapi3.SchemaRefs, 0, len(consts))
		for _, c := range consts {
			out = append(out, &openapi3.SchemaRef{Value: &openapi3.Schema{Const: c}})
		}
		return out
	}

	for _, tc := range []struct {
		name       string
		in         openapi3.SchemaRefs
		wantType   *openapi3.Types
		wantFormat string
		ok         bool
	}{
		{
			name:     "all strings",
			in:       refs("available", "pending", "sold"),
			wantType: &openapi3.Types{"string"},
			ok:       true,
		},
		{
			name:     "small integers keep the natural int",
			in:       refs(float64(8080), float64(9090)),
			wantType: &openapi3.Types{"integer"},
			ok:       true,
		},
		{
			name:       "one large value widens the whole enum",
			in:         refs(float64(1), float64(5_000_000_000)),
			wantType:   &openapi3.Types{"integer"},
			wantFormat: "int64",
			ok:         true,
		},
		{
			name: "a string beside a number has no single Go type",
			in:   refs("available", float64(1)),
			ok:   false,
		},
		{
			name: "an integer beside a fractional number disagrees too",
			in:   refs(float64(1), 1.5),
			ok:   false,
		},
		{
			// Reported honestly; the caller's scalar gate is what rejects it.
			name:     "all booleans",
			in:       refs(true, false),
			wantType: &openapi3.Types{"boolean"},
			ok:       true,
		},
		{
			name: "a non-scalar const stops inference",
			in:   refs("available", []any{1, 2}),
			ok:   false,
		},
		{
			name: "a missing const stops inference",
			in:   refs("available", nil),
			ok:   false,
		},
		{
			name: "no branches",
			in:   nil,
			ok:   false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gotType, gotFormat, ok := inferEnumViaOneOfType(tc.in)
			assert.Equal(t, tc.ok, ok)
			assert.Equal(t, tc.wantType, gotType)
			assert.Equal(t, tc.wantFormat, gotFormat)
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

// detectEnumViaOneOf splices the inferred type into a shallow copy, so the
// caller's schema keeps whatever it declared (here: nothing) and the copy is
// what carries the type and format on to oapiSchemaToGoType.
func TestDetectEnumViaOneOfInfersTypeWithoutMutating(t *testing.T) {
	prev := globalState
	t.Cleanup(func() { globalState = prev })
	globalState.is31 = true
	globalState.options.OutputOptions.SkipEnumViaOneOf = false

	schema := &openapi3.Schema{
		OneOf: openapi3.SchemaRefs{
			{Value: &openapi3.Schema{Title: "Http", Const: float64(8080)}},
			{Value: &openapi3.Schema{Title: "Ephemeral", Const: float64(5_000_000_000)}},
		},
	}

	items, typeSource, err := detectEnumViaOneOf(schema)
	require.NoError(t, err)
	assert.Equal(t, []enumViaOneOfValue{
		{Title: "Http", Value: "8080"},
		{Title: "Ephemeral", Value: "5000000000"},
	}, items)

	assert.NotSame(t, schema, typeSource, "the inferred type belongs on a copy")
	assert.Equal(t, &openapi3.Types{"integer"}, typeSource.Type)
	assert.Equal(t, "int64", typeSource.Format, "5e9 does not fit a 32-bit int")
	assert.Nil(t, schema.Type, "the caller's schema must be left alone")
	assert.Empty(t, schema.Format)
}

// A const that lost precision in the parser is an error, not a fall-through.
// The branches say plainly that an enum was meant, so quietly substituting a
// union would hide the fact that the value cannot be generated faithfully.
// Both spellings must report it: the inferred one, and the declared one that
// never consults the consts on its way to a Go type.
func TestEnumViaOneOfInexactIntegerErrors(t *testing.T) {
	for _, tc := range []struct {
		name  string
		types *openapi3.Types
	}{
		{name: "type inferred from consts", types: nil},
		{name: "type declared by the schema", types: &openapi3.Types{"integer"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prev := globalState
			t.Cleanup(func() { globalState = prev })
			globalState.is31 = true
			globalState.options.OutputOptions.SkipEnumViaOneOf = false

			schema := &openapi3.Schema{
				Type: tc.types,
				OneOf: openapi3.SchemaRefs{
					{Value: &openapi3.Schema{Title: "Fine", Const: float64(1)}},
					// 2^53+1 is the case that matters: it is not representable,
					// so the parser rounded it to 2^53 before we saw it.
					{Value: &openapi3.Schema{Title: "Enormous", Const: float64(9007199254740993)}},
				},
			}

			items, typeSource, err := detectEnumViaOneOf(schema)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "Enormous", "the message should name the offending branch")
			assert.Contains(t, err.Error(), "9007199254740992", "and the value as it actually arrived")
			assert.Nil(t, items)
			assert.Nil(t, typeSource)
		})
	}
}

// skip-enum-via-oneof turns the whole path off before the exactness check runs,
// so opting out of the idiom also opts out of failing over it.
func TestEnumViaOneOfInexactIntegerSkipped(t *testing.T) {
	prev := globalState
	t.Cleanup(func() { globalState = prev })
	globalState.is31 = true
	globalState.options.OutputOptions.SkipEnumViaOneOf = true

	schema := &openapi3.Schema{
		OneOf: openapi3.SchemaRefs{
			{Value: &openapi3.Schema{Title: "Enormous", Const: float64(9007199254740993)}},
		},
	}

	items, _, err := detectEnumViaOneOf(schema)
	require.NoError(t, err, "the escape hatch must cover the error too")
	assert.Nil(t, items)
}

// A oneOf that is not an enum at all still falls through silently -- the error
// is reserved for schemas that are unmistakably enums we cannot generate.
func TestEnumViaOneOfNonEnumsStillFallThrough(t *testing.T) {
	for _, tc := range []struct {
		name string
		refs openapi3.SchemaRefs
	}{
		{
			// Boolean is a scalar family, but not an enumerable one.
			name: "boolean consts",
			refs: openapi3.SchemaRefs{
				{Value: &openapi3.Schema{Title: "Yes", Const: true}},
				{Value: &openapi3.Schema{Title: "No", Const: false}},
			},
		},
		{
			name: "families disagree",
			refs: openapi3.SchemaRefs{
				{Value: &openapi3.Schema{Title: "Word", Const: "one"}},
				{Value: &openapi3.Schema{Title: "Number", Const: float64(1)}},
			},
		},
		{
			// Even alongside an unrepresentable number: the shape is not an
			// enum, so there is nothing to fail about.
			name: "non-scalar const beside a huge one",
			refs: openapi3.SchemaRefs{
				{Value: &openapi3.Schema{Title: "Obj", Const: map[string]any{"a": 1}}},
				{Value: &openapi3.Schema{Title: "Enormous", Const: float64(9007199254740993)}},
			},
		},
		{
			name: "a branch without a title",
			refs: openapi3.SchemaRefs{
				{Value: &openapi3.Schema{Title: "Fine", Const: float64(1)}},
				{Value: &openapi3.Schema{Const: float64(9007199254740993)}},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prev := globalState
			t.Cleanup(func() { globalState = prev })
			globalState.is31 = true
			globalState.options.OutputOptions.SkipEnumViaOneOf = false

			items, _, err := detectEnumViaOneOf(&openapi3.Schema{OneOf: tc.refs})
			require.NoError(t, err, "a non-enum oneOf belongs to the union generator")
			assert.Nil(t, items)
		})
	}
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
