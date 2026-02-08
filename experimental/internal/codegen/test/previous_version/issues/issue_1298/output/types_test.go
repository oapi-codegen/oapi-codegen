package output

import (
	"encoding/json"
	"testing"
)

// TestTestTypeInstantiation verifies that the Test type can be created with
// required (non-pointer) string fields.
func TestTestTypeInstantiation(t *testing.T) {
	test := Test{
		Field1: "value1",
		Field2: "value2",
	}

	if test.Field1 != "value1" {
		t.Errorf("Field1 = %q, want %q", test.Field1, "value1")
	}
	if test.Field2 != "value2" {
		t.Errorf("Field2 = %q, want %q", test.Field2, "value2")
	}
}

// TestTestJSONRoundTrip verifies JSON marshal/unmarshal for the Test type.
func TestTestJSONRoundTrip(t *testing.T) {
	original := Test{
		Field1: "hello",
		Field2: "world",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Test
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Field1 != "hello" {
		t.Errorf("Field1 = %q, want %q", decoded.Field1, "hello")
	}
	if decoded.Field2 != "world" {
		t.Errorf("Field2 = %q, want %q", decoded.Field2, "world")
	}
}

// TestTestRequiredFieldsNoOmitempty verifies that required fields are serialized
// even when empty (no omitempty tag).
func TestTestRequiredFieldsNoOmitempty(t *testing.T) {
	test := Test{}

	data, err := json.Marshal(test)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Both fields should be present in JSON even with zero values
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal raw failed: %v", err)
	}

	if _, ok := raw["field1"]; !ok {
		t.Error("field1 should be present in JSON output for required field")
	}
	if _, ok := raw["field2"]; !ok {
		t.Error("field2 should be present in JSON output for required field")
	}
}

// TestApplyDefaults verifies ApplyDefaults does not panic.
func TestApplyDefaults(t *testing.T) {
	test := &Test{}
	test.ApplyDefaults()
}

// TestGetOpenAPISpecJSON verifies the embedded spec can be decoded.
func TestGetOpenAPISpecJSON(t *testing.T) {
	data, err := GetOpenAPISpecJSON()
	if err != nil {
		t.Fatalf("GetOpenAPISpecJSON failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetOpenAPISpecJSON returned empty data")
	}
}
