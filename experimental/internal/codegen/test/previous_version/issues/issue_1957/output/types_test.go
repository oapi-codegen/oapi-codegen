package output

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// TestTypeWithOptionalFieldInstantiation verifies that TypeWithOptionalField
// uses uuid.UUID fields (via the googleuuid alias in generated code).
func TestTypeWithOptionalFieldInstantiation(t *testing.T) {
	id1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	id2 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

	tw := TypeWithOptionalField{
		At:         id1,
		AtRequired: id2,
	}

	if tw.At != id1 {
		t.Errorf("At = %v, want %v", tw.At, id1)
	}
	if tw.AtRequired != id2 {
		t.Errorf("AtRequired = %v, want %v", tw.AtRequired, id2)
	}
}

// TestTypeWithOptionalFieldJSONRoundTrip verifies JSON round-trip.
func TestTypeWithOptionalFieldJSONRoundTrip(t *testing.T) {
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	original := TypeWithOptionalField{
		At:         id,
		AtRequired: id,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded TypeWithOptionalField
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.At != id {
		t.Errorf("At = %v, want %v", decoded.At, id)
	}
	if decoded.AtRequired != id {
		t.Errorf("AtRequired = %v, want %v", decoded.AtRequired, id)
	}
}

// TestTypeWithAllOfInstantiation verifies TypeWithAllOf with its nested ID field.
func TestTypeWithAllOfInstantiation(t *testing.T) {
	idField := &TypeWithAllOfID{}
	tw := TypeWithAllOf{
		ID: idField,
	}

	if tw.ID == nil {
		t.Fatal("ID should not be nil")
	}
}

// TestTypeWithAllOfJSONRoundTrip verifies JSON round-trip for TypeWithAllOf.
func TestTypeWithAllOfJSONRoundTrip(t *testing.T) {
	original := TypeWithAllOf{
		ID: &TypeWithAllOfID{},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded TypeWithAllOf
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.ID == nil {
		t.Fatal("ID should not be nil after round trip")
	}
}

// TestTypeAliases verifies that UUID-related type aliases work correctly.
func TestTypeAliases(t *testing.T) {
	id := uuid.New()

	// ID is an alias for googleuuid.UUID (= uuid.UUID)
	var idAlias ID = id
	if idAlias != id {
		t.Errorf("ID alias = %v, want %v", idAlias, id)
	}

	// GetRootParameter is an alias for googleuuid.UUID
	var param GetRootParameter = id
	if param != id {
		t.Errorf("GetRootParameter alias = %v, want %v", param, id)
	}

	// TypeWithOptionalFieldAt is an alias for googleuuid.UUID
	var at TypeWithOptionalFieldAt = id
	if at != id {
		t.Errorf("TypeWithOptionalFieldAt alias = %v, want %v", at, id)
	}

	// TypeWithOptionalFieldAtRequired is an alias for googleuuid.UUID
	var atReq TypeWithOptionalFieldAtRequired = id
	if atReq != id {
		t.Errorf("TypeWithOptionalFieldAtRequired alias = %v, want %v", atReq, id)
	}

	// TypeWithAllOfIDAllOf0 is an alias for googleuuid.UUID
	var allOf0 TypeWithAllOfIDAllOf0 = id
	if allOf0 != id {
		t.Errorf("TypeWithAllOfIDAllOf0 alias = %v, want %v", allOf0, id)
	}
}

// TestApplyDefaults verifies ApplyDefaults does not panic on all types.
func TestApplyDefaults(t *testing.T) {
	(&TypeWithOptionalField{}).ApplyDefaults()
	(&TypeWithAllOf{}).ApplyDefaults()
	(&TypeWithAllOfID{}).ApplyDefaults()
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
