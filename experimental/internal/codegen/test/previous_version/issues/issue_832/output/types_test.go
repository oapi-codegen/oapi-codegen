package output

import (
	"encoding/json"
	"testing"
)

// TestEnumTypeGeneration verifies that enum types in properties are generated.
// https://github.com/oapi-codegen/oapi-codegen/issues/832
//
// Note: The x-go-type-name extension is not currently supported. The enum type
// is generated with a name derived from the property path rather than the
// specified x-go-type-name.
func TestEnumTypeGeneration(t *testing.T) {
	// Enum constants should exist (no collision, so no type prefix)
	_ = One
	_ = Two
	_ = Three
	_ = Four

	if string(One) != "one" {
		t.Errorf("one = %q, want %q", One, "one")
	}
}

func TestDocumentWithStatus(t *testing.T) {
	name := "test"
	status := "one"
	doc := Document{
		Name:   &name,
		Status: &status,
	}

	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Document
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if *decoded.Name != *doc.Name {
		t.Errorf("Name = %q, want %q", *decoded.Name, *doc.Name)
	}
	if *decoded.Status != *doc.Status {
		t.Errorf("Status = %q, want %q", *decoded.Status, *doc.Status)
	}
}

func TestDocumentStatusSchema(t *testing.T) {
	// There's also a DocumentStatus schema (separate from the enum property)
	value := "test-value"
	ds := DocumentStatus{
		Value: &value,
	}

	if *ds.Value != "test-value" {
		t.Errorf("Value = %q, want %q", *ds.Value, "test-value")
	}
}
