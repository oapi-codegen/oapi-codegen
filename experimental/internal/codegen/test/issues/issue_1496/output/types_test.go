package output

import (
	"encoding/json"
	"testing"
)

// TestValidIdentifiers verifies that all generated type names are valid Go identifiers.
// Issue 1496: Inline schemas in responses were generating identifiers starting with numbers.
func TestValidIdentifiers(t *testing.T) {
	// If this compiles, the identifiers are valid
	response := GetSomethingJSONResponse{
		Results: []GetSomething200ResponseJSON2{
			{
				GetSomething200ResponseJSONAnyOf0: &GetSomething200ResponseJSONAnyOf0{
					Order: ptr("order-123"),
				},
			},
			{
				GetSomething200ResponseJSONAnyOf11: &GetSomething200ResponseJSONAnyOf11{
					Error: &GetSomething200ResponseJSONAnyOf12{
						Code:    ptr(float32(400)),
						Message: ptr("Bad request"),
					},
				},
			},
		},
	}

	// Should be able to marshal
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	t.Logf("Marshaled response: %s", string(data))
}

func ptr[T any](v T) *T {
	return &v
}
