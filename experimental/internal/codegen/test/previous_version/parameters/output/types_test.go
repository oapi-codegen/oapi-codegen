package output

import (
	"encoding/json"
	"testing"
)

// TestParametersObjectType verifies the Object type has correctly generated
// fields with proper JSON tags.
// V2 test suite: internal/test/parameters
func TestParametersObjectType(t *testing.T) {
	obj := Object{
		Role:      "admin",
		FirstName: "Alice",
	}

	if obj.Role != "admin" {
		t.Errorf("Role = %q, want %q", obj.Role, "admin")
	}
	if obj.FirstName != "Alice" {
		t.Errorf("FirstName = %q, want %q", obj.FirstName, "Alice")
	}
}

// TestParametersComplexObjectType verifies the ComplexObject type has an
// embedded Object field along with ID and IsAdmin fields.
func TestParametersComplexObjectType(t *testing.T) {
	co := ComplexObject{
		Object: Object{
			Role:      "user",
			FirstName: "Bob",
		},
		ID:      42,
		IsAdmin: true,
	}

	if co.Object.Role != "user" {
		t.Errorf("Object.Role = %q, want %q", co.Object.Role, "user")
	}
	if co.Object.FirstName != "Bob" {
		t.Errorf("Object.FirstName = %q, want %q", co.Object.FirstName, "Bob")
	}
	if co.ID != 42 {
		t.Errorf("ID = %d, want %d", co.ID, 42)
	}
	if co.IsAdmin != true {
		t.Errorf("IsAdmin = %v, want true", co.IsAdmin)
	}
}

// TestParametersEnumType verifies the GetEnumsParameter enum type has the
// expected constants.
func TestParametersEnumType(t *testing.T) {
	if N100 != 100 {
		t.Errorf("N100 = %d, want %d", N100, 100)
	}
	if N200 != 200 {
		t.Errorf("N200 = %d, want %d", N200, 200)
	}
}

// TestParametersEnumTypeAssignment verifies that enum values can be assigned
// to the typed enum.
func TestParametersEnumTypeAssignment(t *testing.T) {
	var p GetEnumsParameter
	p = N100
	if p != 100 {
		t.Errorf("p = %d, want %d", p, 100)
	}

	p = N200
	if p != 200 {
		t.Errorf("p = %d, want %d", p, 200)
	}
}

// TestParametersObjectJSONRoundTrip verifies JSON round-trip for the Object
// type.
func TestParametersObjectJSONRoundTrip(t *testing.T) {
	original := Object{
		Role:      "editor",
		FirstName: "Charlie",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Object
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Role != original.Role {
		t.Errorf("Role mismatch: got %q, want %q", decoded.Role, original.Role)
	}
	if decoded.FirstName != original.FirstName {
		t.Errorf("FirstName mismatch: got %q, want %q", decoded.FirstName, original.FirstName)
	}
}

// TestParametersComplexObjectJSONRoundTrip verifies JSON round-trip for the
// ComplexObject type with its nested Object.
func TestParametersComplexObjectJSONRoundTrip(t *testing.T) {
	original := ComplexObject{
		Object: Object{
			Role:      "moderator",
			FirstName: "Diana",
		},
		ID:      99,
		IsAdmin: false,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded ComplexObject
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Object.Role != original.Object.Role {
		t.Errorf("Object.Role mismatch: got %q, want %q", decoded.Object.Role, original.Object.Role)
	}
	if decoded.Object.FirstName != original.Object.FirstName {
		t.Errorf("Object.FirstName mismatch: got %q, want %q", decoded.Object.FirstName, original.Object.FirstName)
	}
	if decoded.ID != original.ID {
		t.Errorf("ID mismatch: got %d, want %d", decoded.ID, original.ID)
	}
	if decoded.IsAdmin != original.IsAdmin {
		t.Errorf("IsAdmin mismatch: got %v, want %v", decoded.IsAdmin, original.IsAdmin)
	}
}

// TestParametersObjectJSONTags verifies that the JSON field names use the
// correct casing from the spec.
func TestParametersObjectJSONTags(t *testing.T) {
	obj := Object{
		Role:      "viewer",
		FirstName: "Eve",
	}

	data, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal into map failed: %v", err)
	}

	if _, ok := m["role"]; !ok {
		t.Error("expected 'role' key in JSON output")
	}
	if _, ok := m["firstName"]; !ok {
		t.Error("expected 'firstName' key in JSON output")
	}
}

// TestParametersComplexObjectJSONTags verifies that the ComplexObject JSON
// field names preserve the spec's casing.
func TestParametersComplexObjectJSONTags(t *testing.T) {
	co := ComplexObject{
		Object: Object{Role: "admin", FirstName: "Frank"},
		ID:     1,
	}

	data, err := json.Marshal(co)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal into map failed: %v", err)
	}

	if _, ok := m["Object"]; !ok {
		t.Error("expected 'Object' key in JSON output")
	}
	if _, ok := m["Id"]; !ok {
		t.Error("expected 'Id' key in JSON output")
	}
	if _, ok := m["IsAdmin"]; !ok {
		t.Error("expected 'IsAdmin' key in JSON output")
	}
}

// TestParametersApplyDefaults verifies that ApplyDefaults can be called on
// all types without panic.
func TestParametersApplyDefaults(t *testing.T) {
	obj := &Object{}
	obj.ApplyDefaults()

	co := &ComplexObject{}
	co.ApplyDefaults()
}
