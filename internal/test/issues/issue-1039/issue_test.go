package issue1039

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/oapi-codegen/nullable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T {
	return &v
}

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
