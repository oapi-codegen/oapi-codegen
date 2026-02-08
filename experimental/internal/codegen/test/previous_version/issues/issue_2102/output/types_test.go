package output

import (
	"encoding/json"
	"testing"
)

// TestBarHasBothProperties verifies that Bar has both foo and bar properties.
// Issue 2102: When a schema has both properties and allOf at the same level,
// the properties were being ignored.
func TestBarHasBothProperties(t *testing.T) {
	// Bar should have both foo (from allOf ref to Foo) and bar (from direct properties)
	bar := Bar{
		Foo: "test-foo",
		Bar: "test-bar",
	}

	// Should be able to marshal/unmarshal
	data, err := json.Marshal(bar)
	if err != nil {
		t.Fatalf("Failed to marshal Bar: %v", err)
	}

	var unmarshaled Bar
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal Bar: %v", err)
	}

	if unmarshaled.Foo != "test-foo" {
		t.Errorf("Expected Foo to be 'test-foo', got %q", unmarshaled.Foo)
	}
	if unmarshaled.Bar != "test-bar" {
		t.Errorf("Expected Bar to be 'test-bar', got %q", unmarshaled.Bar)
	}
}

// TestBarRequiredFields verifies that bar is required (from allOf member's required array).
func TestBarRequiredFields(t *testing.T) {
	// Both foo and bar should be required (no omitempty), so an empty struct
	// should marshal with empty string values
	bar := Bar{}
	data, err := json.Marshal(bar)
	if err != nil {
		t.Fatalf("Failed to marshal empty Bar: %v", err)
	}

	// Both fields should be present in JSON
	expected := `{"bar":"","foo":""}`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}
}
