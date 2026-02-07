package output

import (
	"encoding/json"
	"testing"
)

// TestXHasAllFields verifies that schema X has properties from both its own
// definition AND from the allOf reference to YBase.
// https://github.com/oapi-codegen/oapi-codegen/issues/697
func TestXHasAllFields(t *testing.T) {
	a := "a-value"
	b := 42
	baseField := "base-value"

	x := X{
		A:         &a,
		B:         &b,
		BaseField: &baseField,
	}

	// Verify all fields are accessible
	if *x.A != "a-value" {
		t.Errorf("X.A = %q, want %q", *x.A, "a-value")
	}
	if *x.B != 42 {
		t.Errorf("X.B = %d, want %d", *x.B, 42)
	}
	if *x.BaseField != "base-value" {
		t.Errorf("X.BaseField = %q, want %q", *x.BaseField, "base-value")
	}
}

func TestXJSONRoundTrip(t *testing.T) {
	a := "a-value"
	b := 42
	baseField := "base-value"

	original := X{
		A:         &a,
		B:         &b,
		BaseField: &baseField,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded X
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if *decoded.A != *original.A || *decoded.B != *original.B || *decoded.BaseField != *original.BaseField {
		t.Errorf("Round trip failed: got %+v, want %+v", decoded, original)
	}
}
