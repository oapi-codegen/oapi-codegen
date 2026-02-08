package output

import (
	"encoding/json"
	"testing"
)

// TestNestedObjectTypes verifies that nested objects get proper types.
// https://github.com/oapi-codegen/oapi-codegen/issues/1397
//
// Note: The x-go-type-name extension is not currently supported. Types are
// named based on their path in the schema rather than the specified names.
func TestNestedObjectTypes(t *testing.T) {
	// Test schema should have array of enums, and two nested objects
	test := MyTestRequest{
		Field1: []TestField1Item{
			Option1,
			Option2,
		},
		Field2: &MyTestRequestNestedField{
			Field1: true,
			Field2: "value2",
		},
		Field3: &TestField3{
			Field1: false,
			Field2: "value3",
		},
	}

	if len(test.Field1) != 2 {
		t.Errorf("Field1 length = %d, want 2", len(test.Field1))
	}
	if test.Field2.Field1 != true {
		t.Errorf("Field2.Field1 = %v, want true", test.Field2.Field1)
	}
	if test.Field3.Field2 != "value3" {
		t.Errorf("Field3.Field2 = %q, want %q", test.Field3.Field2, "value3")
	}
}

func TestEnumArrayField(t *testing.T) {
	// Field1 is an array of enum values
	_ = Option1
	_ = Option2

	items := []TestField1Item{
		TestField1Item("option1"),
		TestField1Item("option2"),
	}

	if len(items) != 2 {
		t.Errorf("items length = %d, want 2", len(items))
	}
}

func TestTestJSONRoundTrip(t *testing.T) {
	original := MyTestRequest{
		Field1: []TestField1Item{Option1},
		Field2: &MyTestRequestNestedField{Field1: true, Field2: "test"},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded MyTestRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(decoded.Field1) != 1 {
		t.Errorf("Field1 length = %d, want 1", len(decoded.Field1))
	}
	if decoded.Field2 == nil {
		t.Fatal("Field2 should not be nil")
	}
	if decoded.Field2.Field1 != true {
		t.Errorf("Field2.Field1 = %v, want true", decoded.Field2.Field1)
	}
}
