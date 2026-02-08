package output

import (
	"encoding/json"
	"testing"
)

// TestWithAnyAdditional1JSONRoundTrip verifies custom Marshal/Unmarshal for
// types with additionalProperties: true (map[string]any).
func TestWithAnyAdditional1JSONRoundTrip(t *testing.T) {
	field1 := 42
	field2 := "hello"
	original := WithAnyAdditional1{
		Field1: &field1,
		Field2: &field2,
		AdditionalProperties: map[string]any{
			"extra": "value",
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded WithAnyAdditional1
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Field1 == nil || *decoded.Field1 != 42 {
		t.Errorf("Field1 = %v, want 42", decoded.Field1)
	}
	if decoded.Field2 == nil || *decoded.Field2 != "hello" {
		t.Errorf("Field2 = %v, want %q", decoded.Field2, "hello")
	}
	if decoded.AdditionalProperties["extra"] != "value" {
		t.Errorf("AdditionalProperties[extra] = %v, want %q", decoded.AdditionalProperties["extra"], "value")
	}
}

// TestWithAnyAdditional1AdditionalPropsNotInJSON verifies additional properties
// are separated from known fields during unmarshal.
func TestWithAnyAdditional1AdditionalPropsNotInJSON(t *testing.T) {
	input := `{"field1":10,"field2":"test","unknown_key":true}`

	var decoded WithAnyAdditional1
	if err := json.Unmarshal([]byte(input), &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Field1 == nil || *decoded.Field1 != 10 {
		t.Errorf("Field1 = %v, want 10", decoded.Field1)
	}
	if decoded.AdditionalProperties["unknown_key"] != true {
		t.Errorf("AdditionalProperties[unknown_key] = %v, want true", decoded.AdditionalProperties["unknown_key"])
	}
	// Known fields should NOT appear in additional properties
	if _, ok := decoded.AdditionalProperties["field1"]; ok {
		t.Error("field1 should not be in AdditionalProperties")
	}
}

// TestWithStringAdditional1JSONRoundTrip verifies custom Marshal/Unmarshal for
// types with additionalProperties: {type: string} (map[string]string).
func TestWithStringAdditional1JSONRoundTrip(t *testing.T) {
	field1 := 5
	field2 := "world"
	original := WithStringAdditional1{
		Field1: &field1,
		Field2: &field2,
		AdditionalProperties: map[string]string{
			"custom": "typed-value",
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded WithStringAdditional1
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Field1 == nil || *decoded.Field1 != 5 {
		t.Errorf("Field1 = %v, want 5", decoded.Field1)
	}
	if decoded.AdditionalProperties["custom"] != "typed-value" {
		t.Errorf("AdditionalProperties[custom] = %v, want %q", decoded.AdditionalProperties["custom"], "typed-value")
	}
}

// TestWithAnyAdditional2JSONRoundTrip verifies the second variant with map[string]any.
func TestWithAnyAdditional2JSONRoundTrip(t *testing.T) {
	fieldA := 99
	fieldB := "data"
	original := WithAnyAdditional2{
		FieldA: &fieldA,
		FieldB: &fieldB,
		AdditionalProperties: map[string]any{
			"num": float64(3.14),
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded WithAnyAdditional2
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.FieldA == nil || *decoded.FieldA != 99 {
		t.Errorf("FieldA = %v, want 99", decoded.FieldA)
	}
	if decoded.FieldB == nil || *decoded.FieldB != "data" {
		t.Errorf("FieldB = %v, want %q", decoded.FieldB, "data")
	}
	if decoded.AdditionalProperties["num"] != float64(3.14) {
		t.Errorf("AdditionalProperties[num] = %v, want 3.14", decoded.AdditionalProperties["num"])
	}
}

// TestWithStringAdditional2JSONRoundTrip verifies the second variant with map[string]string.
func TestWithStringAdditional2JSONRoundTrip(t *testing.T) {
	fieldA := 7
	fieldB := "str"
	original := WithStringAdditional2{
		FieldA: &fieldA,
		FieldB: &fieldB,
		AdditionalProperties: map[string]string{
			"key": "val",
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded WithStringAdditional2
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.AdditionalProperties["key"] != "val" {
		t.Errorf("AdditionalProperties[key] = %v, want %q", decoded.AdditionalProperties["key"], "val")
	}
}

// TestDefaultAdditionalPlainStruct verifies that DefaultAdditional types are
// plain structs without custom marshal/unmarshal.
func TestDefaultAdditionalPlainStruct(t *testing.T) {
	field1 := 1
	field2 := "plain"
	d := DefaultAdditional1{
		Field1: &field1,
		Field2: &field2,
	}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded DefaultAdditional1
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Field1 == nil || *decoded.Field1 != 1 {
		t.Errorf("Field1 = %v, want 1", decoded.Field1)
	}
	if decoded.Field2 == nil || *decoded.Field2 != "plain" {
		t.Errorf("Field2 = %v, want %q", decoded.Field2, "plain")
	}
}

// TestMergeTypeConstruction verifies that merged allOf types have all four fields.
func TestMergeTypeConstruction(t *testing.T) {
	field1 := 1
	field2 := "two"
	fieldA := 3
	fieldB := "four"
	m := MergeDefaultDefault{
		Field1: &field1,
		Field2: &field2,
		FieldA: &fieldA,
		FieldB: &fieldB,
	}

	if m.Field1 == nil || *m.Field1 != 1 {
		t.Errorf("Field1 = %v, want 1", m.Field1)
	}
	if m.Field2 == nil || *m.Field2 != "two" {
		t.Errorf("Field2 = %v, want %q", m.Field2, "two")
	}
	if m.FieldA == nil || *m.FieldA != 3 {
		t.Errorf("FieldA = %v, want 3", m.FieldA)
	}
	if m.FieldB == nil || *m.FieldB != "four" {
		t.Errorf("FieldB = %v, want %q", m.FieldB, "four")
	}
}

// TestMergeTypeJSONRoundTrip verifies JSON round-trip for a merge type.
func TestMergeTypeJSONRoundTrip(t *testing.T) {
	field1 := 10
	field2 := "x"
	fieldA := 20
	fieldB := "y"
	original := MergeWithAnyWithString{
		Field1: &field1,
		Field2: &field2,
		FieldA: &fieldA,
		FieldB: &fieldB,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded MergeWithAnyWithString
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Field1 == nil || *decoded.Field1 != 10 {
		t.Errorf("Field1 = %v, want 10", decoded.Field1)
	}
	if decoded.FieldB == nil || *decoded.FieldB != "y" {
		t.Errorf("FieldB = %v, want %q", decoded.FieldB, "y")
	}
}

// TestApplyDefaults verifies ApplyDefaults does not panic on all types.
func TestApplyDefaults(t *testing.T) {
	(&WithAnyAdditional1{}).ApplyDefaults()
	(&WithAnyAdditional2{}).ApplyDefaults()
	(&WithStringAdditional1{}).ApplyDefaults()
	(&WithStringAdditional2{}).ApplyDefaults()
	(&WithoutAdditional1{}).ApplyDefaults()
	(&WithoutAdditional2{}).ApplyDefaults()
	(&DefaultAdditional1{}).ApplyDefaults()
	(&DefaultAdditional2{}).ApplyDefaults()
	(&MergeWithoutWithout{}).ApplyDefaults()
	(&MergeWithoutWithString{}).ApplyDefaults()
	(&MergeWithoutWithAny{}).ApplyDefaults()
	(&MergeWithoutDefault{}).ApplyDefaults()
	(&MergeWithStringWithout{}).ApplyDefaults()
	(&MergeWithStringWithAny{}).ApplyDefaults()
	(&MergeWithStringDefault{}).ApplyDefaults()
	(&MergeWithAnyWithout{}).ApplyDefaults()
	(&MergeWithAnyWithString{}).ApplyDefaults()
	(&MergeWithAnyWithAny{}).ApplyDefaults()
	(&MergeWithAnyDefault{}).ApplyDefaults()
	(&MergeDefaultWithout{}).ApplyDefaults()
	(&MergeDefaultWithString{}).ApplyDefaults()
	(&MergeDefaultWithAny{}).ApplyDefaults()
	(&MergeDefaultDefault{}).ApplyDefaults()
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
