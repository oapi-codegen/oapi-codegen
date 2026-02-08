package output

import (
	"encoding/json"
	"testing"
)

// TestPathWithColon verifies that paths with colons (like /pets:validate) generate properly.
// https://github.com/oapi-codegen/oapi-codegen/issues/312
func TestPathWithColonGeneratesTypes(t *testing.T) {
	// The path /pets:validate should generate a ValidatePetsJSONResponse type
	response := ValidatePetsJSONResponse{
		{Name: "Fluffy"},
		{Name: "Spot"},
	}

	if len(response) != 2 {
		t.Errorf("response length = %d, want 2", len(response))
	}
	if response[0].Name != "Fluffy" {
		t.Errorf("response[0].Name = %q, want %q", response[0].Name, "Fluffy")
	}
}

func TestPetSchema(t *testing.T) {
	pet := Pet{
		Name: "Max",
	}

	data, err := json.Marshal(pet)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := `{"name":"Max"}`
	if string(data) != expected {
		t.Errorf("Marshal result = %s, want %s", string(data), expected)
	}
}

func TestPetNamesSchema(t *testing.T) {
	petNames := PetNames{
		Names: []string{"Fluffy", "Spot", "Max"},
	}

	data, err := json.Marshal(petNames)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded PetNames
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(decoded.Names) != 3 {
		t.Errorf("Names length = %d, want 3", len(decoded.Names))
	}
}

func TestErrorSchema(t *testing.T) {
	err := Error{
		Code:    404,
		Message: "Not Found",
	}

	data, _ := json.Marshal(err)
	expected := `{"code":404,"message":"Not Found"}`
	if string(data) != expected {
		t.Errorf("Marshal result = %s, want %s", string(data), expected)
	}
}
