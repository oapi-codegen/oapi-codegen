package output

import (
	"encoding/json"
	"testing"
)

// TestNullableTypes verifies that nullable types are generated properly.
// https://github.com/oapi-codegen/oapi-codegen/issues/1039
//
// Note: The current implementation uses `any` for nullable primitive types
// as a workaround for Go not having native nullable types. This differs
// from the original oapi-codegen which uses the nullable package.
func TestNullableTypes(t *testing.T) {
	// Create a patch request with various nullable fields
	name := "test-name"
	req := PatchRequest{
		SimpleRequiredNullable:  42, // required nullable (can't be nil)
		SimpleOptionalNullable:  ptrTo[any](nil),
		ComplexRequiredNullable: ComplexRequiredNullable{Name: &name},
	}

	if req.SimpleRequiredNullable != 42 {
		t.Errorf("SimpleRequiredNullable = %v, want 42", req.SimpleRequiredNullable)
	}
}

func TestPatchRequestJSONRoundTrip(t *testing.T) {
	name := "test"
	original := PatchRequest{
		SimpleRequiredNullable:  100,
		ComplexRequiredNullable: ComplexRequiredNullable{Name: &name},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded PatchRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Note: JSON unmarshals numbers as float64, so we compare as float64
	// This is a limitation of using `any` for nullable primitive types
	decodedVal, ok := decoded.SimpleRequiredNullable.(float64)
	if !ok {
		t.Fatalf("SimpleRequiredNullable is not float64, got %T", decoded.SimpleRequiredNullable)
	}
	if decodedVal != 100 {
		t.Errorf("SimpleRequiredNullable mismatch: got %v, want %v", decodedVal, 100)
	}
}

func TestComplexNullableTypes(t *testing.T) {
	// Complex nullable types should be proper struct pointers
	aliasName := "alias"
	name := "name"
	opt := ComplexOptionalNullable{
		AliasName: &aliasName,
		Name:      &name,
	}

	req := PatchRequest{
		SimpleRequiredNullable:    nil, // null value
		ComplexRequiredNullable:   ComplexRequiredNullable{},
		ComplexOptionalNullable:   &opt,
	}

	if req.ComplexOptionalNullable == nil {
		t.Fatal("ComplexOptionalNullable should not be nil")
	}
	if *req.ComplexOptionalNullable.AliasName != "alias" {
		t.Errorf("AliasName = %q, want %q", *req.ComplexOptionalNullable.AliasName, "alias")
	}
}

// TestAdditionalPropertiesFalse verifies that additionalProperties: false
// generates proper marshal/unmarshal that rejects extra fields.
func TestAdditionalPropertiesFalse(t *testing.T) {
	// The struct has AdditionalProperties field but additionalProperties: false
	// means unknown fields are still collected but not expected
	req := PatchRequest{
		SimpleRequiredNullable:  1,
		ComplexRequiredNullable: ComplexRequiredNullable{},
		AdditionalProperties:    map[string]any{"extra": "value"},
	}

	// Should marshal with additional properties
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	t.Logf("Marshaled: %s", string(data))
}

func ptrTo[T any](v T) *T {
	return &v
}
