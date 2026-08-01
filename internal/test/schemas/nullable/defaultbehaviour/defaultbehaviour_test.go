// Package defaultbehaviour tests the nullable-type:false (default) behaviour for
// issue #1039: without nullable-type:true, nullable fields are generated as *T
// pointers instead of nullable.Nullable[T].
package defaultbehaviour

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T {
	return &v
}

// TestNullableDisabled asserts that with nullable-type disabled (the default),
// nullable fields are generated as plain *T pointers.
func TestNullableDisabled(t *testing.T) {
	// include all fields in patch request
	patchReq := PatchRequest{
		ComplexRequiredNullable: &ComplexRequiredNullable{
			Name: ptr("test-name"),
		},
		SimpleOptionalNonNullable: ptr(SimpleOptionalNonNullable("bar")),
		ComplexOptionalNullable: &ComplexOptionalNullable{
			AliasName: ptr("foo-alias"),
			Name:      ptr("foo"),
		},
		SimpleOptionalNullable: ptr(SimpleOptionalNullable(10)),
		SimpleRequiredNullable: ptr(SimpleRequiredNullable(5)),
	}

	expected := []byte(`{"complex_optional_nullable":{"alias_name":"foo-alias","name":"foo"},"complex_required_nullable":{"name":"test-name"},"simple_optional_non_nullable":"bar","simple_optional_nullable":10,"simple_required_nullable":5}`)

	actual, err := json.Marshal(patchReq)
	require.NoError(t, err)
	require.Equal(t, string(expected), string(actual))

	// omit some fields
	patchReq = PatchRequest{
		ComplexRequiredNullable: &ComplexRequiredNullable{
			Name: ptr("test-name"),
		},
		// SimpleOptionalNonNullable is omitted
		ComplexOptionalNullable: &ComplexOptionalNullable{
			AliasName: ptr("test-alias-name"),
			Name:      ptr("test-name"),
		},
		SimpleOptionalNullable: ptr(SimpleOptionalNullable(10)),
		// SimpleRequiredNullable is omitted
	}

	expected = []byte(`{"complex_optional_nullable":{"alias_name":"test-alias-name","name":"test-name"},"complex_required_nullable":{"name":"test-name"},"simple_optional_nullable":10,"simple_required_nullable":null}`)

	actual, err = json.Marshal(patchReq)
	require.NoError(t, err)
	require.Equal(t, string(expected), string(actual))
}
