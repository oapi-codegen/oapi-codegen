package output

import (
	"encoding/json"
	"testing"
)

// TestProblemDetailsInstantiation verifies that ProblemDetails with additionalProperties: true
// generates correctly with both known fields and AdditionalProperties map.
// https://github.com/oapi-codegen/oapi-codegen/issues/1168
func TestProblemDetailsInstantiation(t *testing.T) {
	typVal := "https://example.com/error"
	title := "Not Found"
	status := int32(404)
	detail := "The requested resource was not found"
	instance := "/api/resource/123"

	pd := ProblemDetails{
		Type:     &typVal,
		Title:    &title,
		Status:   &status,
		Detail:   &detail,
		Instance: &instance,
	}

	if *pd.Type != "https://example.com/error" {
		t.Errorf("Type = %q, want %q", *pd.Type, "https://example.com/error")
	}
	if *pd.Title != "Not Found" {
		t.Errorf("Title = %q, want %q", *pd.Title, "Not Found")
	}
	if *pd.Status != 404 {
		t.Errorf("Status = %d, want %d", *pd.Status, 404)
	}
	if *pd.Detail != "The requested resource was not found" {
		t.Errorf("Detail = %q, want %q", *pd.Detail, "The requested resource was not found")
	}
	if *pd.Instance != "/api/resource/123" {
		t.Errorf("Instance = %q, want %q", *pd.Instance, "/api/resource/123")
	}
}

func TestProblemDetailsAdditionalProperties(t *testing.T) {
	pd := ProblemDetails{
		AdditionalProperties: map[string]any{
			"traceId":  "abc-123",
			"errorRef": float64(42),
		},
	}

	if pd.AdditionalProperties["traceId"] != "abc-123" {
		t.Errorf("AdditionalProperties[traceId] = %v, want %q",
			pd.AdditionalProperties["traceId"], "abc-123")
	}
	if pd.AdditionalProperties["errorRef"] != float64(42) {
		t.Errorf("AdditionalProperties[errorRef] = %v, want %v",
			pd.AdditionalProperties["errorRef"], float64(42))
	}
}

func TestProblemDetailsJSONRoundTrip(t *testing.T) {
	typVal := "https://example.com/validation"
	title := "Validation Error"
	status := int32(422)
	detail := "Field 'name' is required"
	instance := "/api/users"

	original := ProblemDetails{
		Type:     &typVal,
		Title:    &title,
		Status:   &status,
		Detail:   &detail,
		Instance: &instance,
		AdditionalProperties: map[string]any{
			"errors": []any{"name is required", "email is invalid"},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded ProblemDetails
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if *decoded.Type != *original.Type {
		t.Errorf("Type mismatch: got %q, want %q", *decoded.Type, *original.Type)
	}
	if *decoded.Title != *original.Title {
		t.Errorf("Title mismatch: got %q, want %q", *decoded.Title, *original.Title)
	}
	if *decoded.Status != *original.Status {
		t.Errorf("Status mismatch: got %d, want %d", *decoded.Status, *original.Status)
	}
	if *decoded.Detail != *original.Detail {
		t.Errorf("Detail mismatch: got %q, want %q", *decoded.Detail, *original.Detail)
	}
	if *decoded.Instance != *original.Instance {
		t.Errorf("Instance mismatch: got %q, want %q", *decoded.Instance, *original.Instance)
	}

	errors, ok := decoded.AdditionalProperties["errors"]
	if !ok {
		t.Fatal("AdditionalProperties should contain 'errors' after round trip")
	}
	errSlice, ok := errors.([]any)
	if !ok {
		t.Fatalf("errors should be a slice, got %T", errors)
	}
	if len(errSlice) != 2 {
		t.Errorf("errors length = %d, want 2", len(errSlice))
	}
}

func TestProblemDetailsApplyDefaults(t *testing.T) {
	pd := &ProblemDetails{}
	pd.ApplyDefaults()

	// Default value for Type should be "about:blank"
	if pd.Type == nil {
		t.Fatal("Type should not be nil after ApplyDefaults")
	}
	if *pd.Type != "about:blank" {
		t.Errorf("Type = %q, want %q", *pd.Type, "about:blank")
	}
}

func TestProblemDetailsApplyDefaultsDoesNotOverwrite(t *testing.T) {
	typVal := "https://example.com/custom"
	pd := &ProblemDetails{
		Type: &typVal,
	}
	pd.ApplyDefaults()

	// ApplyDefaults should not overwrite an already-set value
	if *pd.Type != "https://example.com/custom" {
		t.Errorf("Type = %q, want %q (should not be overwritten)", *pd.Type, "https://example.com/custom")
	}
}

func TestProblemDetailsJSONWithAdditionalPropertiesOnly(t *testing.T) {
	// Test marshaling with only additional properties set
	pd := ProblemDetails{
		AdditionalProperties: map[string]any{
			"custom": "value",
		},
	}

	data, err := json.Marshal(pd)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded ProblemDetails
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.AdditionalProperties["custom"] != "value" {
		t.Errorf("AdditionalProperties[custom] = %v, want %q",
			decoded.AdditionalProperties["custom"], "value")
	}
	// Known fields should be nil
	if decoded.Type != nil {
		t.Errorf("Type should be nil, got %q", *decoded.Type)
	}
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
