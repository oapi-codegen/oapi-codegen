package output

import (
	"encoding/json"
	"testing"
)

// TestWhateverInstantiation verifies that the Whatever type is generated correctly
// when multiple content types reference the same schema.
// https://github.com/oapi-codegen/oapi-codegen/issues/1127
func TestWhateverInstantiation(t *testing.T) {
	prop := "test-value"
	w := Whatever{
		SomeProperty: &prop,
	}

	if *w.SomeProperty != "test-value" {
		t.Errorf("SomeProperty = %q, want %q", *w.SomeProperty, "test-value")
	}
}

func TestWhateverNilProperty(t *testing.T) {
	w := Whatever{}
	if w.SomeProperty != nil {
		t.Errorf("SomeProperty should be nil by default, got %v", *w.SomeProperty)
	}
}

func TestWhateverJSONRoundTrip(t *testing.T) {
	prop := "round-trip-value"
	original := Whatever{
		SomeProperty: &prop,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Whatever
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.SomeProperty == nil {
		t.Fatal("SomeProperty should not be nil after round trip")
	}
	if *decoded.SomeProperty != *original.SomeProperty {
		t.Errorf("SomeProperty mismatch: got %q, want %q", *decoded.SomeProperty, *original.SomeProperty)
	}
}

func TestWhateverOmitEmpty(t *testing.T) {
	// SomeProperty is optional (omitempty), so empty struct should marshal to {}
	w := Whatever{}
	data, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := `{}`
	if string(data) != expected {
		t.Errorf("Marshal result = %s, want %s", string(data), expected)
	}
}

func TestWhateverApplyDefaults(t *testing.T) {
	w := &Whatever{}
	w.ApplyDefaults()
	// Should not panic
}

func TestGetOpenAPISpecJSON(t *testing.T) {
	data, err := GetOpenAPISpecJSON()
	if err != nil {
		t.Fatalf("GetOpenAPISpecJSON failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetOpenAPISpecJSON returned empty data")
	}
}
