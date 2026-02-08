package output

import (
	"encoding/json"
	"testing"
)

// TestNameNormalizerPetType verifies that the Pet type has correctly named
// fields with proper JSON tags.
// V2 test suite: internal/test/outputoptions/name_normalizer
func TestNameNormalizerPetType(t *testing.T) {
	pet := Pet{
		UUID: "550e8400-e29b-41d4-a716-446655440000",
		Name: "Buddy",
	}

	if pet.UUID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("UUID = %q, want the set value", pet.UUID)
	}
	if pet.Name != "Buddy" {
		t.Errorf("Name = %q, want %q", pet.Name, "Buddy")
	}
}

// TestNameNormalizerErrorType verifies the Error type has correctly named
// fields.
func TestNameNormalizerErrorType(t *testing.T) {
	err := Error{
		Code:    404,
		Message: "not found",
	}

	if err.Code != 404 {
		t.Errorf("Code = %d, want %d", err.Code, 404)
	}
	if err.Message != "not found" {
		t.Errorf("Message = %q, want %q", err.Message, "not found")
	}
}

// TestNameNormalizerOneOf2ThingsType verifies the OneOf2Things union type
// with two inline schemas and different ID types (int vs UUID).
func TestNameNormalizerOneOf2ThingsType(t *testing.T) {
	variant0 := OneOf2ThingsOneOf0{ID: 42}
	thing := OneOf2Things{
		OneOf2ThingsOneOf0: &variant0,
	}

	if thing.OneOf2ThingsOneOf0 == nil {
		t.Fatal("OneOf2ThingsOneOf0 should not be nil")
	}
	if thing.OneOf2ThingsOneOf0.ID != 42 {
		t.Errorf("ID = %d, want %d", thing.OneOf2ThingsOneOf0.ID, 42)
	}
	if thing.OneOf2ThingsOneOf1 != nil {
		t.Error("OneOf2ThingsOneOf1 should be nil")
	}
}

// TestNameNormalizerOneOf2ThingsOneOf0 verifies the first oneOf variant with
// an int ID.
func TestNameNormalizerOneOf2ThingsOneOf0(t *testing.T) {
	v := OneOf2ThingsOneOf0{ID: 100}
	if v.ID != 100 {
		t.Errorf("ID = %d, want %d", v.ID, 100)
	}
}

// TestNameNormalizerOneOf2ThingsOneOf1 verifies the second oneOf variant with
// a UUID ID.
func TestNameNormalizerOneOf2ThingsOneOf1(t *testing.T) {
	// UUID is a type alias for uuid.UUID
	var id UUID
	v := OneOf2ThingsOneOf1{ID: id}
	_ = v
}

// TestNameNormalizerOneOf2ThingsMarshalInt verifies marshaling the oneOf type
// with an int ID variant.
func TestNameNormalizerOneOf2ThingsMarshalInt(t *testing.T) {
	thing := OneOf2Things{
		OneOf2ThingsOneOf0: &OneOf2ThingsOneOf0{ID: 7},
	}

	data, err := json.Marshal(thing)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal into map failed: %v", err)
	}

	// The ID should be a number
	idVal, ok := m["id"]
	if !ok {
		t.Fatal("expected 'id' key in JSON output")
	}
	// JSON numbers unmarshal as float64
	if idVal != float64(7) {
		t.Errorf("id = %v, want %v", idVal, 7)
	}
}

// TestNameNormalizerPetJSONRoundTrip verifies JSON round-trip for the Pet
// type.
func TestNameNormalizerPetJSONRoundTrip(t *testing.T) {
	original := Pet{
		UUID: "abc-123",
		Name: "Felix",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Pet
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.UUID != original.UUID {
		t.Errorf("UUID mismatch: got %q, want %q", decoded.UUID, original.UUID)
	}
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %q, want %q", decoded.Name, original.Name)
	}
}

// TestNameNormalizerErrorJSONRoundTrip verifies JSON round-trip for the Error
// type.
func TestNameNormalizerErrorJSONRoundTrip(t *testing.T) {
	original := Error{
		Code:    500,
		Message: "internal server error",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Error
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Code != original.Code {
		t.Errorf("Code mismatch: got %d, want %d", decoded.Code, original.Code)
	}
	if decoded.Message != original.Message {
		t.Errorf("Message mismatch: got %q, want %q", decoded.Message, original.Message)
	}
}

// TestNameNormalizerApplyDefaults verifies that ApplyDefaults can be called
// on all types without panic.
func TestNameNormalizerApplyDefaults(t *testing.T) {
	pet := &Pet{}
	pet.ApplyDefaults()

	e := &Error{}
	e.ApplyDefaults()

	v0 := &OneOf2ThingsOneOf0{}
	v0.ApplyDefaults()

	v1 := &OneOf2ThingsOneOf1{}
	v1.ApplyDefaults()
}
