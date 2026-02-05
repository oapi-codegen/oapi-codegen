package output

import (
	"encoding/json"
	"testing"
)

// TestEnumGenerated verifies that the enum type is generated for properties inside anyOf.
// Issue 1429: enum type was not being generated when used inside anyOf.
func TestEnumGenerated(t *testing.T) {
	// The enum type should exist and have the expected constants
	_ = TestAnyOf1FieldA_foo
	_ = TestAnyOf1FieldA_bar

	// The alias should also exist
	_ = TestAnyOf1FieldA(TestAnyOf1FieldA_foo)
}

// TestAnyOfMarshal verifies that the anyOf type can be marshaled.
func TestAnyOfMarshal(t *testing.T) {
	test := Test{
		TestAnyOf1: &TestAnyOf1{
			FieldA: ptr("foo"),
		},
	}

	data, err := json.Marshal(test)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	t.Logf("Marshaled: %s", string(data))
}

func ptr[T any](v T) *T {
	return &v
}
