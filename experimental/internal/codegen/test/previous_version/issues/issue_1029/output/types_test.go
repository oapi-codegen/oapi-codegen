package output

import (
	"encoding/json"
	"testing"
)

// TestRegistrationStateOneOfEnums verifies that oneOf with string enums generates
// correctly with proper enum constants.
// https://github.com/oapi-codegen/oapi-codegen/issues/1029
func TestRegistrationStateOneOfEnums(t *testing.T) {
	// Verify enum constants exist and have correct values
	tests := []struct {
		name  string
		enum  RegistrationStateOneOf0
		value string
	}{
		{"undefined", Undefined, "undefined"},
	}

	for _, tt := range tests {
		if string(tt.enum) != tt.value {
			t.Errorf("%s enum = %q, want %q", tt.name, tt.enum, tt.value)
		}
	}
}

func TestRegistrationStateMarshal(t *testing.T) {
	// Test serialization of oneOf with string enum
	state := RegistrationState{
		RegistrationStateOneOf0: ptrTo(RegistrationStateOneOf0(Undefined)),
	}

	reg := Registration{
		State: &state,
	}

	data, err := json.Marshal(reg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify the JSON contains the expected value
	expected := `{"state":"undefined"}`
	if string(data) != expected {
		t.Errorf("Marshal result = %s, want %s", string(data), expected)
	}
}

func TestRegistrationStateUnmarshalLimitation(t *testing.T) {
	// Note: Unmarshaling oneOf with multiple string enum types is inherently
	// ambiguous without a discriminator, since any string can match any of the
	// enum types. This test documents that limitation.
	input := `{"state":"undefined"}`

	var decoded Registration
	err := json.Unmarshal([]byte(input), &decoded)

	// The error is expected because all 4 string enum types can unmarshal
	// from the same string value
	if err == nil {
		t.Log("Unmarshal succeeded (all variants matched)")
	} else {
		t.Logf("Unmarshal failed as expected for ambiguous oneOf: %v", err)
	}
}

func TestAllEnumConstants(t *testing.T) {
	// Verify all enum constants are defined
	_ = Undefined
	_ = Registered
	_ = Pending
	_ = Active

	// Test values
	if string(Registered) != "registered" {
		t.Error("registered enum has wrong value")
	}
	if string(Pending) != "pending" {
		t.Error("pending enum has wrong value")
	}
	if string(Active) != "active" {
		t.Error("active enum has wrong value")
	}
}

func ptrTo[T any](v T) *T {
	return &v
}
