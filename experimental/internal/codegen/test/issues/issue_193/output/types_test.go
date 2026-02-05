package output

import (
	"encoding/json"
	"testing"
)

// TestAllOfWithAdditionalProperties verifies that allOf with additionalProperties: true
// merges fields correctly from multiple allOf members.
// https://github.com/oapi-codegen/oapi-codegen/issues/193
func TestAllOfWithAdditionalProperties(t *testing.T) {
	name := "John"
	age := float32(30)

	person := Person{
		Metadata: "some-metadata",
		Name:     &name,
		Age:      &age,
	}

	// All fields from both allOf members should be present
	if person.Metadata != "some-metadata" {
		t.Errorf("Metadata = %q, want %q", person.Metadata, "some-metadata")
	}
	if *person.Name != "John" {
		t.Errorf("Name = %q, want %q", *person.Name, "John")
	}
	if *person.Age != 30 {
		t.Errorf("Age = %v, want %v", *person.Age, 30)
	}
}

func TestPersonJSONRoundTrip(t *testing.T) {
	name := "Jane"
	age := float32(25)
	original := Person{
		Metadata: "meta",
		Name:     &name,
		Age:      &age,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Person
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Metadata != original.Metadata {
		t.Errorf("Metadata mismatch: got %q, want %q", decoded.Metadata, original.Metadata)
	}
	if *decoded.Name != *original.Name {
		t.Errorf("Name mismatch: got %q, want %q", *decoded.Name, *original.Name)
	}
	if *decoded.Age != *original.Age {
		t.Errorf("Age mismatch: got %v, want %v", *decoded.Age, *original.Age)
	}
}

func TestMetadataIsRequired(t *testing.T) {
	// Metadata is required (no omitempty), so empty struct should marshal with empty string
	person := Person{}
	data, err := json.Marshal(person)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Should contain "metadata" even if empty
	expected := `{"metadata":""}`
	if string(data) != expected {
		t.Errorf("Marshal result = %s, want %s", string(data), expected)
	}
}
