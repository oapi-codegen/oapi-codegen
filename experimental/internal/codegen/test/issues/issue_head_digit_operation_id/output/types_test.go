package output

import (
	"encoding/json"
	"testing"
)

// TestOperationIdStartingWithDigit verifies that operation IDs starting with
// digits generate valid Go identifiers with an N prefix.
func TestOperationIdStartingWithDigit(t *testing.T) {
	// The operationId is "3GPPFoo" which should generate N3GPPFooJSONResponse
	// (N prefix added to make it a valid Go identifier)
	value := "test"
	response := N3GPPFooJSONResponse{
		Value: &value,
	}

	if *response.Value != "test" {
		t.Errorf("Value = %q, want %q", *response.Value, "test")
	}
}

func TestN3GPPFooJSONRoundTrip(t *testing.T) {
	value := "test-value"
	original := N3GPPFooJSONResponse{
		Value: &value,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded N3GPPFooJSONResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if *decoded.Value != *original.Value {
		t.Errorf("Value = %q, want %q", *decoded.Value, *original.Value)
	}
}

// TestTypeNameIsValid ensures the type name is a valid Go identifier
func TestTypeNameIsValid(t *testing.T) {
	// This test passes if it compiles - the type N3GPPFooJSONResponse
	// must be a valid Go identifier
	var _ N3GPPFooJSONResponse
	var _ N3GPPFooJSONResponse
}
