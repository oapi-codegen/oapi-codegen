package schemasnullable

import (
	"encoding/json"
	"testing"

	"github.com/oapi-codegen/nullable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	spec30 "github.com/oapi-codegen/oapi-codegen/v2/internal/test/schemas/nullable/spec30"
	spec31 "github.com/oapi-codegen/oapi-codegen/v2/internal/test/schemas/nullable/spec31"
)

func ptr[T any](v T) *T {
	return &v
}

// ---- issue #1039: required/optional x nullable matrix with nullable-type:true ----

// TestNullableTypesMarshal exercises marshaling of the PatchRequest with
// nullable.Nullable[T] fields under nullable-type:true (issue #1039).
func TestNullableTypesMarshal(t *testing.T) {
	// include all fields in patch request
	patchReq := PatchRequest{
		ComplexRequiredNullable: nullable.NewNullableWithValue(ComplexRequiredNullable{
			Name: ptr("test-name"),
		}),
		SimpleOptionalNonNullable: ptr(SimpleOptionalNonNullable("bar")),
		ComplexOptionalNullable: nullable.NewNullableWithValue(ComplexOptionalNullable{
			AliasName: nullable.NewNullableWithValue("foo-alias"),
			Name:      ptr("foo"),
		}),
		SimpleOptionalNullable: nullable.NewNullableWithValue(10),
		SimpleRequiredNullable: nullable.NewNullableWithValue(5),
	}

	expected := []byte(`{"complex_optional_nullable":{"alias_name":"foo-alias","name":"foo"},"complex_required_nullable":{"name":"test-name"},"simple_optional_non_nullable":"bar","simple_optional_nullable":10,"simple_required_nullable":5}`)

	actual, err := json.Marshal(patchReq)
	require.NoError(t, err)
	require.Equal(t, string(expected), string(actual))

	// omit some fields
	patchReq = PatchRequest{
		ComplexRequiredNullable: nullable.NewNullableWithValue(ComplexRequiredNullable{
			Name: ptr("test-name"),
		}),
		// SimpleOptionalNonNullable is omitted
		ComplexOptionalNullable: nullable.NewNullableWithValue(ComplexOptionalNullable{
			AliasName: nullable.NewNullableWithValue("test-alias-name"),
			Name:      ptr("test-name"),
		}),
		SimpleOptionalNullable: nullable.NewNullableWithValue(10),
		// SimpleRequiredNullable is omitted
	}

	expected = []byte(`{"complex_optional_nullable":{"alias_name":"test-alias-name","name":"test-name"},"complex_required_nullable":{"name":"test-name"},"simple_optional_nullable":10,"simple_required_nullable":0}`)

	actual, err = json.Marshal(patchReq)
	require.NoError(t, err)
	require.Equal(t, string(expected), string(actual))
}

// TestNullableTypesUnmarshal exercises unmarshaling into PatchRequest with
// nullable.Nullable[T] fields under nullable-type:true (issue #1039).
func TestNullableTypesUnmarshal(t *testing.T) {
	type testCase struct {
		name   string
		json   []byte
		assert func(t *testing.T, obj PatchRequest)
	}
	tests := []testCase{
		{
			name: "when empty json is provided",
			json: []byte(`{}`),
			assert: func(t *testing.T, obj PatchRequest) {
				t.Helper()

				// check for nullable fields
				assert.Falsef(t, obj.SimpleRequiredNullable.IsSpecified(), "SimpleRequiredNullable field should not be set")
				assert.Falsef(t, obj.SimpleRequiredNullable.IsNull(), "SimpleRequiredNullable field should not be null")

				assert.Falsef(t, obj.SimpleOptionalNullable.IsSpecified(), "SimpleOptionalNullable field should not be set")
				assert.Falsef(t, obj.SimpleOptionalNullable.IsNull(), "SimpleOptionalNullable field should not be null")

				assert.Falsef(t, obj.ComplexOptionalNullable.IsSpecified(), "ComplexOptionalNullable field should not be set")
				assert.Falsef(t, obj.ComplexOptionalNullable.IsNull(), "ComplexOptionalNullable field should not be null")

				assert.Falsef(t, obj.ComplexRequiredNullable.IsSpecified(), "ComplexRequiredNullable field should not be set")
				assert.Falsef(t, obj.ComplexRequiredNullable.IsNull(), "ComplexRequiredNullable field should not be null")

				// check for non-nullable field
				assert.Nilf(t, obj.SimpleOptionalNonNullable, "SimpleOptionalNonNullable field should be nil")
			},
		},

		{
			name: "when only empty complex_optional_nullable is provided",
			json: []byte(`{"complex_optional_nullable":{}}`),
			assert: func(t *testing.T, obj PatchRequest) {
				t.Helper()
				// check for nullable field
				assert.Truef(t, obj.ComplexOptionalNullable.IsSpecified(), "ComplexOptionalNullable field should be set")
				assert.Falsef(t, obj.ComplexOptionalNullable.IsNull(), "ComplexOptionalNullable field should not be null")

				// other simple nullable fields should not be set and should not be null
				assert.Falsef(t, obj.SimpleRequiredNullable.IsSpecified(), "SimpleRequiredNullable field should not be set")
				assert.Falsef(t, obj.SimpleRequiredNullable.IsNull(), "SimpleRequiredNullable field should not be null")

				assert.Falsef(t, obj.SimpleOptionalNullable.IsSpecified(), "SimpleOptionalNullable field should not be set")
				assert.Falsef(t, obj.SimpleOptionalNullable.IsNull(), "SimpleOptionalNullable field should not be null")

				// other complex nullable fields should not be set and should not be null
				assert.Falsef(t, obj.ComplexRequiredNullable.IsSpecified(), "ComplexRequiredNullable field should not be set")
				assert.Falsef(t, obj.ComplexRequiredNullable.IsNull(), "ComplexRequiredNullable field should not be null")

				// other non-nullable field should have its zero value
				assert.Nilf(t, obj.SimpleOptionalNonNullable, "SimpleOptionalNonNullable field should be nil")

			},
		},

		{
			name: "when only complex_optional_nullable with its `name` child field is provided",
			json: []byte(`{"complex_optional_nullable":{"name":"test-name"}}`),
			assert: func(t *testing.T, obj PatchRequest) {
				t.Helper()

				assert.Truef(t, obj.ComplexOptionalNullable.IsSpecified(), "ComplexOptionalNullable field should be set")
				assert.Falsef(t, obj.ComplexOptionalNullable.IsNull(), "ComplexOptionalNullable field should not be null")

				gotComplexObj, err := obj.ComplexOptionalNullable.Get()
				require.NoError(t, err)
				assert.Equalf(t, "test-name", string(*gotComplexObj.Name), "name should  be test-name")

				assert.Falsef(t, gotComplexObj.AliasName.IsSpecified(), "child field `alias name` should not be specified")
				assert.Falsef(t, gotComplexObj.AliasName.IsNull(), "child field `alias name` should not be null")
			},
		},

		{
			name: "when only complex_optional_nullable child fields `name` and `alias name` are provided with non-zero and null values respectively",
			json: []byte(`{"complex_optional_nullable":{"name":"test-name","alias_name":null}}`),
			assert: func(t *testing.T, obj PatchRequest) {
				t.Helper()

				assert.Truef(t, obj.ComplexOptionalNullable.IsSpecified(), "ComplexOptionalNullable field should be set")
				assert.Falsef(t, obj.ComplexOptionalNullable.IsNull(), "ComplexOptionalNullable field should not be null")

				gotComplexObj, err := obj.ComplexOptionalNullable.Get()
				require.NoError(t, err)
				assert.Equalf(t, "test-name", string(*gotComplexObj.Name), "name should  be test-name")

				assert.Truef(t, gotComplexObj.AliasName.IsSpecified(), "child field `alias name` should be set")
				assert.Truef(t, gotComplexObj.AliasName.IsNull(), "child field `alias name` should be null")
			},
		},

		{
			name: "when simple_required_nullable is null ",
			json: []byte(`{"simple_required_nullable":null}`),
			assert: func(t *testing.T, obj PatchRequest) {
				t.Helper()

				assert.Truef(t, obj.SimpleRequiredNullable.IsSpecified(), "SimpleRequiredNullable field should be set")
				assert.Truef(t, obj.SimpleRequiredNullable.IsNull(), "SimpleRequiredNullable field should be null")
			},
		},

		{
			name: "when simple_required_nullable is null and organization has non zero value",
			json: []byte(`{"complex_optional_nullable":{"name":"foo","alias_name":"bar"},"simple_required_nullable":null}`),
			assert: func(t *testing.T, obj PatchRequest) {
				t.Helper()

				assert.Truef(t, obj.SimpleRequiredNullable.IsSpecified(), "SimpleRequiredNullable field should be set")
				assert.Truef(t, obj.SimpleRequiredNullable.IsNull(), "SimpleRequiredNullable field should be null")

				assert.Truef(t, obj.ComplexOptionalNullable.IsSpecified(), "ComplexOptionalNullable field should be set")
				assert.Falsef(t, obj.ComplexOptionalNullable.IsNull(), "ComplexOptionalNullable field should not be null")

				gotComplexObj, err := obj.ComplexOptionalNullable.Get()
				require.NoError(t, err)
				assert.Equalf(t, "foo", string(*gotComplexObj.Name), "child field `name` should  be foo")

				assert.Truef(t, gotComplexObj.AliasName.IsSpecified(), "child field `alias name` should be set")
				assert.Falsef(t, gotComplexObj.AliasName.IsNull(), "child field `alias name` should not be null")

				gotAliasName, err := gotComplexObj.AliasName.Get()
				require.NoError(t, err)
				assert.Equalf(t, "bar", gotAliasName, "child field `alias name` should be bar")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var obj PatchRequest
			err := json.Unmarshal(tt.json, &obj)
			require.NoError(t, err)

			tt.assert(t, obj)
		})
	}
}

// ---- issue #2185: array of nullable items using nullable.Nullable[T] ----

// TestContainer_UsesNullableType asserts that array items marked nullable:true
// generate as []nullable.Nullable[T] rather than []*T (issue #2185).
func TestContainer_UsesNullableType(t *testing.T) {
	c := Container{
		MayBeNull: []nullable.Nullable[string]{
			nullable.NewNullNullable[string](),
		},
	}

	require.Len(t, c.MayBeNull, 1)
	require.True(t, c.MayBeNull[0].IsNull())
}

// ---- openapi31_nullable: 3.0 vs 3.1 nullable idiom equivalence ----

// TestNicknameIsPointer_3_0 asserts that the 3.0 spec's `nullable: true`
// produces a `*string` field. Compile-time verification: the assignment
// of `&nick` only succeeds if Nickname is `*string`.
func TestNicknameIsPointer_3_0(t *testing.T) {
	nick := "rex"
	p := spec30.Pet{
		Name:     "fluffy",
		Nickname: &nick,
	}
	require.NotNil(t, p.Nickname)
	assert.Equal(t, "rex", *p.Nickname)

	// nil-nickname round-trip
	p2 := spec30.Pet{Name: "fluffy"}
	assert.Nil(t, p2.Nickname)
}

// TestNicknameIsPointer_3_1 asserts that the 3.1 spec's
// `type: ["string","null"]` produces a `*string` field, identical in
// shape to the 3.0 control case.
func TestNicknameIsPointer_3_1(t *testing.T) {
	nick := "rex"
	p := spec31.Pet{
		Name:     "fluffy",
		Nickname: &nick,
	}
	require.NotNil(t, p.Nickname)
	assert.Equal(t, "rex", *p.Nickname)

	p2 := spec31.Pet{Name: "fluffy"}
	assert.Nil(t, p2.Nickname)
}

// TestNullableArrayAndObjectFields_3_1 asserts that nullable array and
// nullable inline-object fields generate as pointer-to-slice and
// pointer-to-struct respectively in 3.1. Regression for the case where
// `type: ["array","null"]` and `type: ["object","null"]` were rejected
// before schemaPrimaryType was applied at the GenerateGoSchema dispatch
// (see pkg/codegen/schema.go).
func TestNullableArrayAndObjectFields_3_1(t *testing.T) {
	tags := []string{"good", "boy"}
	id := "owner-1"
	p := spec31.Pet{
		Name: "fluffy",
		Tags: &tags,
		Owner: &struct {
			Id *string `json:"id,omitempty"`
		}{Id: &id},
	}
	require.NotNil(t, p.Tags)
	assert.Equal(t, []string{"good", "boy"}, *p.Tags)
	require.NotNil(t, p.Owner)
	require.NotNil(t, p.Owner.Id)
	assert.Equal(t, "owner-1", *p.Owner.Id)

	// Zero-value: nullable fields must be nil, not empty.
	p2 := spec31.Pet{Name: "fluffy"}
	assert.Nil(t, p2.Tags)
	assert.Nil(t, p2.Owner)
}

// TestNullableArrayAndObjectFields_3_0 is the matching control case
// asserting that `nullable: true` arrays and inline objects in 3.0
// generate the same pointer shape.
func TestNullableArrayAndObjectFields_3_0(t *testing.T) {
	tags := []string{"good", "boy"}
	id := "owner-1"
	p := spec30.Pet{
		Name: "fluffy",
		Tags: &tags,
		Owner: &struct {
			Id *string `json:"id,omitempty"`
		}{Id: &id},
	}
	require.NotNil(t, p.Tags)
	assert.Equal(t, []string{"good", "boy"}, *p.Tags)
	require.NotNil(t, p.Owner)
	require.NotNil(t, p.Owner.Id)
	assert.Equal(t, "owner-1", *p.Owner.Id)

	p2 := spec30.Pet{Name: "fluffy"}
	assert.Nil(t, p2.Tags)
	assert.Nil(t, p2.Owner)
}

// TestNullableUnspecifiedObject_3_1 asserts that a bare nullable
// object (`type: ["object","null"]` with no `properties:`) generates
// as `*map[string]interface{}`. This is the gap flagged in the
// kin-openapi-3.1 PR review: `Schema.Is(...)` strict equality on a
// 3.1 type-array failed to recognize the primary type as "object",
// routing the schema away from the unspecified-object code path.
// Also asserts that the reversed type-array order
// (`type: ["null","object"]`) resolves identically -- no code path
// may peek only at the first element of the array.
func TestNullableUnspecifiedObject_3_1(t *testing.T) {
	extras := map[string]interface{}{"k": "v"}
	metadata := map[string]interface{}{"j": float64(1)}
	p := spec31.Pet{
		Name:     "fluffy",
		Extras:   &extras,
		Metadata: &metadata,
	}
	require.NotNil(t, p.Extras)
	assert.Equal(t, "v", (*p.Extras)["k"])
	require.NotNil(t, p.Metadata)
	assert.Equal(t, float64(1), (*p.Metadata)["j"])

	// Zero-value: both unspecified-object fields must be nil.
	p2 := spec31.Pet{Name: "fluffy"}
	assert.Nil(t, p2.Extras)
	assert.Nil(t, p2.Metadata)
}

// TestNullableUnspecifiedObject_3_0 is the matching 3.0 control case
// asserting parity with the 3.1 type-array form.
func TestNullableUnspecifiedObject_3_0(t *testing.T) {
	extras := map[string]interface{}{"k": "v"}
	p := spec30.Pet{
		Name:   "fluffy",
		Extras: &extras,
	}
	require.NotNil(t, p.Extras)
	assert.Equal(t, "v", (*p.Extras)["k"])

	p2 := spec30.Pet{Name: "fluffy"}
	assert.Nil(t, p2.Extras)
}

// TestNullableViaAnyOfOneOf_3_1 asserts that `anyOf: [{type: string},
// {type: "null"}]` and the matching `oneOf` form both generate as
// `*string`, identical to the type-array idiom (`type: ["string",
// "null"]`). The `{"type": "null"}` branch is a nullability marker and
// must be skipped during union generation -- before the fix, the
// recursive GenerateGoSchema call on the null-only branch failed with
// `unhandled Schema type: &[null]`. And once null is filtered out,
// the single remaining branch must be collapsed to its underlying
// type instead of being wrapped in a one-variant union, so the two
// idioms produce the same Go API surface.
func TestNullableViaAnyOfOneOf_3_1(t *testing.T) {
	nick := "rex"

	// Compile-time check: the AnyOf and OneOf fields must be *string
	// (assignment of &nick succeeds only for a pointer-to-string field).
	p := spec31.Pet{
		Name:          "fluffy",
		NicknameAnyOf: &nick,
		NicknameOneOf: &nick,
	}
	require.NotNil(t, p.NicknameAnyOf)
	require.NotNil(t, p.NicknameOneOf)
	assert.Equal(t, "rex", *p.NicknameAnyOf)
	assert.Equal(t, "rex", *p.NicknameOneOf)

	// Zero-value: both nullable fields must be nil.
	p2 := spec31.Pet{Name: "fluffy"}
	assert.Nil(t, p2.NicknameAnyOf)
	assert.Nil(t, p2.NicknameOneOf)

	// JSON round-trip: an explicit string in / explicit string out;
	// missing field decodes to nil and re-encodes as absent (omitempty).
	const populated = `{"name":"fluffy","nicknameAnyOf":"rex","nicknameOneOf":"rex"}`
	encoded, err := json.Marshal(p)
	require.NoError(t, err)
	assert.JSONEq(t, populated, string(encoded))

	const empty = `{"name":"fluffy"}`
	var p3 spec31.Pet
	require.NoError(t, json.Unmarshal([]byte(empty), &p3))
	assert.Nil(t, p3.NicknameAnyOf)
	assert.Nil(t, p3.NicknameOneOf)
}

// TestNullableDiscriminatedUnion_3_1 asserts that a `oneOf` with a
// discriminator and a `{"type": "null"}` branch generates without the
// `discriminator: not all schemas were mapped` error. Before the fix,
// the null branch was filtered out of mapping construction but still
// counted toward the expected-mapping total, so the completeness check
// `len(Mapping) < len(elements)` falsely tripped for nullable
// discriminated unions. The fix compares against the number of
// non-null branches actually processed.
func TestNullableDiscriminatedUnion_3_1(t *testing.T) {
	// Cat round-trip via the generated discriminated-union accessors.
	const catJSON = `{"kind":"Cat","meow":true}`
	var pet spec31.DiscriminatedPet
	require.NoError(t, json.Unmarshal([]byte(catJSON), &pet))
	cat, err := pet.AsCat()
	require.NoError(t, err)
	require.NotNil(t, cat.Meow)
	assert.True(t, *cat.Meow)

	encoded, err := json.Marshal(pet)
	require.NoError(t, err)
	assert.JSONEq(t, catJSON, string(encoded))

	// Dog round-trip: the second non-null branch must also map.
	const dogJSON = `{"kind":"Dog","bark":true}`
	var pet2 spec31.DiscriminatedPet
	require.NoError(t, json.Unmarshal([]byte(dogJSON), &pet2))
	dog, err := pet2.AsDog()
	require.NoError(t, err)
	require.NotNil(t, dog.Bark)
	assert.True(t, *dog.Bark)
}

// ---- issue #2430: bare OpenAPI 3.1 `type: "null"` schemas map to Go `any` ----
//
// Regression test for https://github.com/oapi-codegen/oapi-codegen/issues/2430.
//
// OpenAPI 3.1 allows a schema whose only type is "null" (as opposed to a
// type array that merely includes "null"). Such a schema validates exactly
// the JSON value null. The generator used to fail with "unhandled Schema
// type: &[null]"; it now maps the schema to `any`, with no optional
// pointer, since nil is already `any`'s zero value.

func TestNullTypePropertyRoundTrip_3_1(t *testing.T) {
	var v spec31.ChallengeOpenJson
	require.NoError(t, json.Unmarshal([]byte(`{"id":"1","challenger":null}`), &v))
	assert.Equal(t, "1", v.Id)
	assert.Nil(t, v.Challenger)

	out, err := json.Marshal(spec31.RequiredNull{})
	require.NoError(t, err)
	// The required null-typed field must serialize as an explicit null.
	assert.JSONEq(t, `{"value":null}`, string(out))
}

func TestNullTypeComponentIsAny_3_1(t *testing.T) {
	// NullOnly is an alias for `any`; assigning an arbitrary value must
	// compile, and nil round-trips through JSON as null.
	var n spec31.NullOnly
	require.NoError(t, json.Unmarshal([]byte(`null`), &n))
	assert.Nil(t, n)
}

// TestJsonRoundTrip_NullableFields_AcrossVersions asserts that a JSON
// payload with an explicit null nickname unmarshals to (*string)(nil) in
// both spec versions, and that JSON output omits the field when nil due
// to omitempty. The two generated structs must marshal identically for
// the nullable field.
func TestJsonRoundTrip_NullableFields_AcrossVersions(t *testing.T) {
	const withName = `{"name":"fluffy"}`
	const withBoth = `{"name":"fluffy","nickname":"rex"}`

	for _, tc := range []struct {
		name string
		// fn30 / fn31 unmarshal the input into each version's Pet type
		// and return a JSON re-marshal so we can assert equality.
		fn30 func(input string) (string, *string, error)
		fn31 func(input string) (string, *string, error)
	}{
		{
			name: "unmarshal/marshal symmetric across versions",
			fn30: func(input string) (string, *string, error) {
				var p spec30.Pet
				if err := json.Unmarshal([]byte(input), &p); err != nil {
					return "", nil, err
				}
				out, err := json.Marshal(p)
				return string(out), p.Nickname, err
			},
			fn31: func(input string) (string, *string, error) {
				var p spec31.Pet
				if err := json.Unmarshal([]byte(input), &p); err != nil {
					return "", nil, err
				}
				out, err := json.Marshal(p)
				return string(out), p.Nickname, err
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, in := range []string{withName, withBoth} {
				out30, n30, err30 := tc.fn30(in)
				require.NoError(t, err30)
				out31, n31, err31 := tc.fn31(in)
				require.NoError(t, err31)
				assert.JSONEq(t, in, out30, "3.0 round-trip should be lossless")
				assert.JSONEq(t, in, out31, "3.1 round-trip should be lossless")
				assert.JSONEq(t, out30, out31, "3.0 and 3.1 must marshal identically")
				if n30 == nil {
					assert.Nil(t, n31)
				} else {
					require.NotNil(t, n31)
					assert.Equal(t, *n30, *n31)
				}
			}
		})
	}
}
