package output

import (
	"encoding/json"
	"testing"
)

func TestSchemaObjectInstantiation(t *testing.T) {
	s := SchemaObject{
		Role:                  "admin",
		FirstName:             "Alice",
		ReadOnlyRequiredProp:  "readonly",
		WriteOnlyRequiredProp: 42,
	}
	if s.Role != "admin" {
		t.Errorf("unexpected role: %s", s.Role)
	}
}

func TestSchemaObjectJSONRoundTrip(t *testing.T) {
	s := SchemaObject{Role: "user", FirstName: "Bob", ReadOnlyRequiredProp: "ro", WriteOnlyRequiredProp: 1}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded SchemaObject
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Role != "user" || decoded.FirstName != "Bob" {
		t.Errorf("round-trip failed: %+v", decoded)
	}
}

func TestAdditionalPropertiesObject1(t *testing.T) {
	opt := "opt"
	a := AdditionalPropertiesObject1{
		Name:                 "test",
		ID:                   1,
		Optional:             &opt,
		AdditionalProperties: map[string]int{"extra": 99},
	}
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded AdditionalPropertiesObject1
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Name != "test" || decoded.ID != 1 {
		t.Errorf("round-trip failed for known fields: %+v", decoded)
	}
	if decoded.AdditionalProperties["extra"] != 99 {
		t.Errorf("round-trip failed for additional properties: %v", decoded.AdditionalProperties)
	}
}

func TestFunnyValuesEnumConstants(t *testing.T) {
	// All FunnyValues constants should be unique and have expected values
	tests := []struct {
		name     string
		constant FunnyValues
		value    string
	}{
		{"star", FunnyValuesAsterisk, "*"},
		{"five", FunnyValuesN5, "5"},
		{"ampersand", FunnyValuesAnd, "&"},
		{"percent", FunnyValuesPercent, "%"},
		{"empty", FunnyValuesEmpty, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.value {
				t.Errorf("constant = %q, want %q", tt.constant, tt.value)
			}
		})
	}
}

func TestEnum1Constants(t *testing.T) {
	if Enum1One != "One" {
		t.Errorf("unexpected: %s", Enum1One)
	}
	if Enum1Two != "Two" {
		t.Errorf("unexpected: %s", Enum1Two)
	}
}

func TestOneOfObject1(t *testing.T) {
	// OneOfObject1 is a union type with variant fields
	v := OneOfVariant1{Name: "test"}
	o := OneOfObject1{OneOfVariant1: &v}
	data, err := o.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded OneOfObject1
	if err := decoded.UnmarshalJSON(data); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
}

func TestApplyDefaults(t *testing.T) {
	s := &SchemaObject{}
	s.ApplyDefaults()
	a := &AdditionalPropertiesObject1{}
	a.ApplyDefaults()
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
