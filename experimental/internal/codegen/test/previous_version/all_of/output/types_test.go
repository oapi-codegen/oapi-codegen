package output

import (
	"encoding/json"
	"testing"
)

// TestAllOfPersonProperties verifies that the PersonProperties type has all
// optional fields generated from the allOf base schema.
// V2 test suite: internal/test/components/allof
func TestAllOfPersonProperties(t *testing.T) {
	firstName := "John"
	lastName := "Doe"
	govID := int64(123456)

	pp := PersonProperties{
		FirstName:          &firstName,
		LastName:           &lastName,
		GovernmentIDNumber: &govID,
	}

	if *pp.FirstName != "John" {
		t.Errorf("FirstName = %q, want %q", *pp.FirstName, "John")
	}
	if *pp.LastName != "Doe" {
		t.Errorf("LastName = %q, want %q", *pp.LastName, "Doe")
	}
	if *pp.GovernmentIDNumber != 123456 {
		t.Errorf("GovernmentIDNumber = %d, want %d", *pp.GovernmentIDNumber, 123456)
	}
}

// TestAllOfPerson verifies that the Person type has required first/last name
// fields (non-pointer) and optional GovernmentIDNumber (pointer), reflecting
// the allOf merge with a required-fields schema.
func TestAllOfPerson(t *testing.T) {
	govID := int64(999)
	p := Person{
		FirstName:          "Jane",
		LastName:           "Smith",
		GovernmentIDNumber: &govID,
	}

	if p.FirstName != "Jane" {
		t.Errorf("FirstName = %q, want %q", p.FirstName, "Jane")
	}
	if p.LastName != "Smith" {
		t.Errorf("LastName = %q, want %q", p.LastName, "Smith")
	}
	if *p.GovernmentIDNumber != 999 {
		t.Errorf("GovernmentIDNumber = %d, want %d", *p.GovernmentIDNumber, 999)
	}
}

// TestAllOfPersonWithID verifies the PersonWithID type which adds an ID field
// via allOf composition on top of PersonProperties.
func TestAllOfPersonWithID(t *testing.T) {
	firstName := "Alice"
	lastName := "Jones"
	govID := int64(555)

	pwid := PersonWithID{
		FirstName:          &firstName,
		LastName:           &lastName,
		GovernmentIDNumber: &govID,
		ID:                 42,
	}

	if *pwid.FirstName != "Alice" {
		t.Errorf("FirstName = %q, want %q", *pwid.FirstName, "Alice")
	}
	if pwid.ID != 42 {
		t.Errorf("ID = %d, want %d", pwid.ID, 42)
	}
}

// TestPersonJSONRoundTrip verifies that Person can be marshaled and
// unmarshaled via JSON with required and optional fields.
func TestPersonJSONRoundTrip(t *testing.T) {
	govID := int64(789)
	original := Person{
		FirstName:          "Bob",
		LastName:           "Brown",
		GovernmentIDNumber: &govID,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Person
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.FirstName != original.FirstName {
		t.Errorf("FirstName mismatch: got %q, want %q", decoded.FirstName, original.FirstName)
	}
	if decoded.LastName != original.LastName {
		t.Errorf("LastName mismatch: got %q, want %q", decoded.LastName, original.LastName)
	}
	if *decoded.GovernmentIDNumber != *original.GovernmentIDNumber {
		t.Errorf("GovernmentIDNumber mismatch: got %d, want %d", *decoded.GovernmentIDNumber, *original.GovernmentIDNumber)
	}
}

// TestPersonWithIDJSONRoundTrip verifies JSON round-trip for the composed
// PersonWithID type.
func TestPersonWithIDJSONRoundTrip(t *testing.T) {
	firstName := "Carol"
	lastName := "Davis"
	original := PersonWithID{
		FirstName: &firstName,
		LastName:  &lastName,
		ID:        100,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded PersonWithID
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if *decoded.FirstName != *original.FirstName {
		t.Errorf("FirstName mismatch: got %q, want %q", *decoded.FirstName, *original.FirstName)
	}
	if decoded.ID != original.ID {
		t.Errorf("ID mismatch: got %d, want %d", decoded.ID, original.ID)
	}
}

// TestPersonRequiredFieldsSerialization verifies that required fields appear in
// JSON even when zero-valued, while optional fields are omitted when nil.
func TestPersonRequiredFieldsSerialization(t *testing.T) {
	p := Person{}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal into map failed: %v", err)
	}

	// Required fields should be present
	if _, ok := m["FirstName"]; !ok {
		t.Error("expected FirstName key in JSON output")
	}
	if _, ok := m["LastName"]; !ok {
		t.Error("expected LastName key in JSON output")
	}

	// Optional field should be absent when nil
	if _, ok := m["GovernmentIDNumber"]; ok {
		t.Error("expected GovernmentIDNumber to be absent when nil")
	}
}

// TestApplyDefaults verifies that ApplyDefaults can be called on all types
// without panic.
func TestApplyDefaults(t *testing.T) {
	pp := &PersonProperties{}
	pp.ApplyDefaults()

	p := &Person{}
	p.ApplyDefaults()

	pwid := &PersonWithID{}
	pwid.ApplyDefaults()
}
