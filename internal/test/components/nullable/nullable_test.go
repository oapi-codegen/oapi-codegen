package nullable

import (
	"encoding/json"
	"testing"

	nullablepkg "github.com/oapi-codegen/nullable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T {
	return &v
}

// From issue-1039: marshal with all fields present.
func TestNullableTypesMarshal(t *testing.T) {
	patchReq := PatchRequest{
		ComplexRequiredNullable: nullablepkg.NewNullableWithValue(ComplexRequiredNullable{
			Name: ptr("test-name"),
		}),
		SimpleOptionalNonNullable: ptr(SimpleOptionalNonNullable("bar")),
		ComplexOptionalNullable: nullablepkg.NewNullableWithValue(ComplexOptionalNullable{
			AliasName: nullablepkg.NewNullableWithValue("foo-alias"),
			Name:      ptr("foo"),
		}),
		SimpleOptionalNullable: nullablepkg.NewNullableWithValue(SimpleOptionalNullable(10)),
		SimpleRequiredNullable: nullablepkg.NewNullableWithValue(SimpleRequiredNullable(5)),
	}

	expected := `{"complex_optional_nullable":{"alias_name":"foo-alias","name":"foo"},"complex_required_nullable":{"name":"test-name"},"simple_optional_non_nullable":"bar","simple_optional_nullable":10,"simple_required_nullable":5}`

	actual, err := json.Marshal(patchReq)
	require.NoError(t, err)
	require.Equal(t, expected, string(actual))
}

// From issue-1039: marshal with some fields omitted.
func TestNullableTypesMarshalPartial(t *testing.T) {
	patchReq := PatchRequest{
		ComplexRequiredNullable: nullablepkg.NewNullableWithValue(ComplexRequiredNullable{
			Name: ptr("test-name"),
		}),
		ComplexOptionalNullable: nullablepkg.NewNullableWithValue(ComplexOptionalNullable{
			AliasName: nullablepkg.NewNullableWithValue("test-alias-name"),
			Name:      ptr("test-name"),
		}),
		SimpleOptionalNullable: nullablepkg.NewNullableWithValue(SimpleOptionalNullable(10)),
	}

	expected := `{"complex_optional_nullable":{"alias_name":"test-alias-name","name":"test-name"},"complex_required_nullable":{"name":"test-name"},"simple_optional_nullable":10,"simple_required_nullable":0}`

	actual, err := json.Marshal(patchReq)
	require.NoError(t, err)
	require.Equal(t, expected, string(actual))
}

// From issue-1039: unmarshal empty JSON.
func TestNullableTypesUnmarshalEmpty(t *testing.T) {
	var obj PatchRequest
	err := json.Unmarshal([]byte(`{}`), &obj)
	require.NoError(t, err)

	assert.False(t, obj.SimpleRequiredNullable.IsSpecified())
	assert.False(t, obj.SimpleRequiredNullable.IsNull())
	assert.False(t, obj.SimpleOptionalNullable.IsSpecified())
	assert.False(t, obj.SimpleOptionalNullable.IsNull())
	assert.False(t, obj.ComplexOptionalNullable.IsSpecified())
	assert.False(t, obj.ComplexOptionalNullable.IsNull())
	assert.False(t, obj.ComplexRequiredNullable.IsSpecified())
	assert.False(t, obj.ComplexRequiredNullable.IsNull())
	assert.Nil(t, obj.SimpleOptionalNonNullable)
}

// From issue-1039: unmarshal with empty complex_optional_nullable.
func TestNullableTypesUnmarshalEmptyComplex(t *testing.T) {
	var obj PatchRequest
	err := json.Unmarshal([]byte(`{"complex_optional_nullable":{}}`), &obj)
	require.NoError(t, err)

	assert.True(t, obj.ComplexOptionalNullable.IsSpecified())
	assert.False(t, obj.ComplexOptionalNullable.IsNull())

	assert.False(t, obj.SimpleRequiredNullable.IsSpecified())
	assert.False(t, obj.SimpleOptionalNullable.IsSpecified())
	assert.False(t, obj.ComplexRequiredNullable.IsSpecified())
	assert.Nil(t, obj.SimpleOptionalNonNullable)
}

// From issue-1039: unmarshal with nested nullable child fields.
func TestNullableTypesUnmarshalNestedFields(t *testing.T) {
	var obj PatchRequest
	err := json.Unmarshal([]byte(`{"complex_optional_nullable":{"name":"test-name"}}`), &obj)
	require.NoError(t, err)

	assert.True(t, obj.ComplexOptionalNullable.IsSpecified())
	assert.False(t, obj.ComplexOptionalNullable.IsNull())

	gotComplexObj, err := obj.ComplexOptionalNullable.Get()
	require.NoError(t, err)
	assert.Equal(t, "test-name", *gotComplexObj.Name)
	assert.False(t, gotComplexObj.AliasName.IsSpecified())
	assert.False(t, gotComplexObj.AliasName.IsNull())
}

// From issue-1039: unmarshal with explicit null child field.
func TestNullableTypesUnmarshalNullChild(t *testing.T) {
	var obj PatchRequest
	err := json.Unmarshal([]byte(`{"complex_optional_nullable":{"name":"test-name","alias_name":null}}`), &obj)
	require.NoError(t, err)

	gotComplexObj, err := obj.ComplexOptionalNullable.Get()
	require.NoError(t, err)
	assert.Equal(t, "test-name", *gotComplexObj.Name)
	assert.True(t, gotComplexObj.AliasName.IsSpecified())
	assert.True(t, gotComplexObj.AliasName.IsNull())
}

// From issue-1039: unmarshal with explicit null on required field.
func TestNullableTypesUnmarshalNullRequired(t *testing.T) {
	var obj PatchRequest
	err := json.Unmarshal([]byte(`{"simple_required_nullable":null}`), &obj)
	require.NoError(t, err)

	assert.True(t, obj.SimpleRequiredNullable.IsSpecified())
	assert.True(t, obj.SimpleRequiredNullable.IsNull())
}

// From issue-1039: unmarshal with null required and non-null complex.
func TestNullableTypesUnmarshalMixed(t *testing.T) {
	var obj PatchRequest
	err := json.Unmarshal([]byte(`{"complex_optional_nullable":{"name":"foo","alias_name":"bar"},"simple_required_nullable":null}`), &obj)
	require.NoError(t, err)

	assert.True(t, obj.SimpleRequiredNullable.IsSpecified())
	assert.True(t, obj.SimpleRequiredNullable.IsNull())

	assert.True(t, obj.ComplexOptionalNullable.IsSpecified())
	assert.False(t, obj.ComplexOptionalNullable.IsNull())

	gotComplexObj, err := obj.ComplexOptionalNullable.Get()
	require.NoError(t, err)
	assert.Equal(t, "foo", *gotComplexObj.Name)
	assert.True(t, gotComplexObj.AliasName.IsSpecified())
	assert.False(t, gotComplexObj.AliasName.IsNull())

	gotAliasName, err := gotComplexObj.AliasName.Get()
	require.NoError(t, err)
	assert.Equal(t, "bar", gotAliasName)
}

// From issue-2185: array of nullable items.
func TestContainerWithNullableArrayItems(t *testing.T) {
	c := Container{
		MayBeNull: []nullablepkg.Nullable[string]{
			nullablepkg.NewNullNullable[string](),
		},
	}

	require.Len(t, c.MayBeNull, 1)
	require.True(t, c.MayBeNull[0].IsNull())
}

func TestContainerWithMixedNullableItems(t *testing.T) {
	c := Container{
		MayBeNull: []nullablepkg.Nullable[string]{
			nullablepkg.NewNullableWithValue("hello"),
			nullablepkg.NewNullNullable[string](),
			nullablepkg.NewNullableWithValue("world"),
		},
	}

	require.Len(t, c.MayBeNull, 3)

	val, err := c.MayBeNull[0].Get()
	require.NoError(t, err)
	assert.Equal(t, "hello", val)

	assert.True(t, c.MayBeNull[1].IsNull())

	val, err = c.MayBeNull[2].Get()
	require.NoError(t, err)
	assert.Equal(t, "world", val)
}

func TestContainerNullableJSONRoundTrip(t *testing.T) {
	c := Container{
		MayBeNull: []nullablepkg.Nullable[string]{
			nullablepkg.NewNullableWithValue("hello"),
			nullablepkg.NewNullNullable[string](),
		},
	}

	data, err := json.Marshal(c)
	require.NoError(t, err)

	var roundTripped Container
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err)

	require.Len(t, roundTripped.MayBeNull, 2)
	val, err := roundTripped.MayBeNull[0].Get()
	require.NoError(t, err)
	assert.Equal(t, "hello", val)
	assert.True(t, roundTripped.MayBeNull[1].IsNull())
}
