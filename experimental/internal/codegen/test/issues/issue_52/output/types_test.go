package output

import (
	"encoding/json"
	"testing"
)

// TestRecursiveTypes verifies that recursive type definitions work correctly.
// https://github.com/oapi-codegen/oapi-codegen/issues/52
func TestRecursiveTypes(t *testing.T) {
	// Value references ArrayValue which is []Value - recursive
	str := "test"
	val := Value{
		StringValue: &str,
		ArrayValue: &ArrayValue{
			{StringValue: &str},
		},
	}

	if *val.StringValue != "test" {
		t.Errorf("StringValue = %q, want %q", *val.StringValue, "test")
	}
	if len(*val.ArrayValue) != 1 {
		t.Errorf("ArrayValue length = %d, want 1", len(*val.ArrayValue))
	}
}

func TestRecursiveJSONRoundTrip(t *testing.T) {
	str := "test"
	nested := "nested"
	original := Value{
		StringValue: &str,
		ArrayValue: &ArrayValue{
			{StringValue: &nested},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Value
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if *decoded.StringValue != *original.StringValue {
		t.Errorf("StringValue mismatch: got %q, want %q", *decoded.StringValue, *original.StringValue)
	}
}

func TestDocumentWithRecursiveFields(t *testing.T) {
	// Document.Fields is map[string]any (due to additionalProperties: $ref Value)
	doc := Document{
		Fields: map[string]any{
			"key1": "value1",
		},
	}

	if doc.Fields["key1"] != "value1" {
		t.Errorf("Fields[key1] = %v, want %q", doc.Fields["key1"], "value1")
	}
}
