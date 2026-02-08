package output

import (
	"encoding/json"
	"testing"
)

// TestResponseBodyOptionalFieldWithNoType verifies that an optional field with no type info
// is generated as *any.
// https://github.com/oapi-codegen/oapi-codegen/issues/609
func TestResponseBodyOptionalFieldWithNoType(t *testing.T) {
	// Unknown field should be *any (pointer to any), allowing any value or nil
	rb := ResponseBody{}
	if rb.Unknown != nil {
		t.Errorf("Unknown should be nil by default, got %v", rb.Unknown)
	}
}

func TestResponseBodyWithStringValue(t *testing.T) {
	val := any("hello")
	rb := ResponseBody{
		Unknown: &val,
	}

	if *rb.Unknown != "hello" {
		t.Errorf("Unknown = %v, want %q", *rb.Unknown, "hello")
	}
}

func TestResponseBodyWithNumericValue(t *testing.T) {
	val := any(42.0)
	rb := ResponseBody{
		Unknown: &val,
	}

	if *rb.Unknown != 42.0 {
		t.Errorf("Unknown = %v, want %v", *rb.Unknown, 42.0)
	}
}

func TestResponseBodyJSONRoundTrip(t *testing.T) {
	val := any("test-value")
	original := ResponseBody{
		Unknown: &val,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded ResponseBody
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Unknown == nil {
		t.Fatal("Unknown should not be nil after round trip")
	}
	if *decoded.Unknown != "test-value" {
		t.Errorf("Unknown = %v, want %q", *decoded.Unknown, "test-value")
	}
}

func TestResponseBodyJSONRoundTripOmitEmpty(t *testing.T) {
	// When Unknown is nil, it should be omitted from JSON (omitempty)
	original := ResponseBody{}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := `{}`
	if string(data) != expected {
		t.Errorf("Marshal result = %s, want %s", string(data), expected)
	}
}

func TestResponseBodyApplyDefaults(t *testing.T) {
	rb := &ResponseBody{}
	rb.ApplyDefaults()
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
