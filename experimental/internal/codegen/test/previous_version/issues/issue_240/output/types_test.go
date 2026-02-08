package output

import (
	"encoding/json"
	"testing"
)

// TestClientTypeInstantiation verifies that ClientType is generated with the expected fields.
// https://github.com/oapi-codegen/oapi-codegen/issues/240
func TestClientTypeInstantiation(t *testing.T) {
	ct := ClientType{
		Name: "test-client",
	}

	if ct.Name != "test-client" {
		t.Errorf("Name = %q, want %q", ct.Name, "test-client")
	}
}

func TestClientTypeJSONRoundTrip(t *testing.T) {
	original := ClientType{
		Name: "my-client",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded ClientType
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %q, want %q", decoded.Name, original.Name)
	}
}

func TestClientTypeNameIsRequired(t *testing.T) {
	// Name is required (no omitempty), so empty struct should marshal with empty string
	ct := ClientType{}
	data, err := json.Marshal(ct)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := `{"name":""}`
	if string(data) != expected {
		t.Errorf("Marshal result = %s, want %s", string(data), expected)
	}
}

func TestUnreferencedType(t *testing.T) {
	// Unreferenced is generated because skip-prune is enabled in the config
	u := Unreferenced{
		ID: "some-id",
	}

	if u.ID != "some-id" {
		t.Errorf("ID = %v, want %q", u.ID, "some-id")
	}
}

func TestUpdateClient400JSONResponse(t *testing.T) {
	resp := UpdateClient400JSONResponse{
		Code: "invalid_request",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded UpdateClient400JSONResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Code != resp.Code {
		t.Errorf("Code mismatch: got %q, want %q", decoded.Code, resp.Code)
	}
}

func TestApplyDefaults(t *testing.T) {
	// ApplyDefaults should be callable on all types without panic
	ct := &ClientType{}
	ct.ApplyDefaults()

	u := &Unreferenced{}
	u.ApplyDefaults()

	resp := &UpdateClient400JSONResponse{}
	resp.ApplyDefaults()
}

func TestGetOpenAPISpecJSON(t *testing.T) {
	data, err := GetOpenAPISpecJSON()
	if err != nil {
		t.Fatalf("GetOpenAPISpecJSON failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GetOpenAPISpecJSON returned empty data")
	}
}
