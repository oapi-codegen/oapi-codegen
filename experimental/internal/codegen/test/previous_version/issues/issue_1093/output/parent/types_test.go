package parent

import (
	"encoding/json"
	"testing"
)

// TestPetConstruction verifies that Pet can be constructed with its fields.
// https://github.com/oapi-codegen/oapi-codegen/issues/1093
func TestPetConstruction(t *testing.T) {
	tag := "friendly"
	pet := Pet{
		Name: "Fido",
		Tag:  &tag,
	}
	if pet.Name != "Fido" {
		t.Errorf("Pet.Name = %q, want %q", pet.Name, "Fido")
	}
	if pet.Tag == nil || *pet.Tag != "friendly" {
		t.Errorf("Pet.Tag = %v, want %q", pet.Tag, "friendly")
	}

	// Tag is optional (pointer)
	petNoTag := Pet{Name: "Rex"}
	if petNoTag.Tag != nil {
		t.Errorf("Pet.Tag should be nil, got %v", petNoTag.Tag)
	}
}

func TestPetJSONRoundTrip(t *testing.T) {
	tag := "indoor"
	original := Pet{Name: "Whiskers", Tag: &tag}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Pet
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Name != "Whiskers" {
		t.Errorf("decoded Name = %q, want %q", decoded.Name, "Whiskers")
	}
	if decoded.Tag == nil || *decoded.Tag != "indoor" {
		t.Errorf("decoded Tag = %v, want %q", decoded.Tag, "indoor")
	}
}

func TestPetJSONOmitsNilTag(t *testing.T) {
	pet := Pet{Name: "Buddy"}
	data, err := json.Marshal(pet)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal to map failed: %v", err)
	}

	if _, ok := m["tag"]; ok {
		t.Error("nil Tag should be omitted from JSON output")
	}
}

func TestPetApplyDefaults(t *testing.T) {
	pet := &Pet{Name: "Test"}
	pet.ApplyDefaults() // should not panic
}

func TestGetOpenAPISpecJSON(t *testing.T) {
	data, err := GetOpenAPISpecJSON()
	if err != nil {
		t.Fatalf("GetOpenAPISpecJSON() failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("GetOpenAPISpecJSON() returned empty data")
	}
}
