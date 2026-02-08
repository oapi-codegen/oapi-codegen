package output

import (
	"encoding/json"
	"testing"
)

// TestArrayContainerInstantiation verifies the ArrayContainer type with its
// values slice and additional properties map.
func TestArrayContainerInstantiation(t *testing.T) {
	ac := ArrayContainer{
		Values: []string{"a", "b", "c"},
		AdditionalProperties: map[string]any{
			"extra": "data",
		},
	}

	if len(ac.Values) != 3 {
		t.Errorf("Values len = %d, want 3", len(ac.Values))
	}
	if ac.Values[0] != "a" {
		t.Errorf("Values[0] = %q, want %q", ac.Values[0], "a")
	}
	if ac.AdditionalProperties["extra"] != "data" {
		t.Errorf("AdditionalProperties[extra] = %v, want %q", ac.AdditionalProperties["extra"], "data")
	}
}

// TestArrayContainerJSONRoundTrip verifies custom MarshalJSON/UnmarshalJSON
// correctly handles both the values array and additional properties.
func TestArrayContainerJSONRoundTrip(t *testing.T) {
	original := ArrayContainer{
		Values: []string{"x", "y"},
		AdditionalProperties: map[string]any{
			"num":  float64(42),
			"flag": true,
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded ArrayContainer
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(decoded.Values) != 2 {
		t.Fatalf("Values len = %d, want 2", len(decoded.Values))
	}
	if decoded.Values[0] != "x" || decoded.Values[1] != "y" {
		t.Errorf("Values = %v, want [x y]", decoded.Values)
	}
	if decoded.AdditionalProperties["num"] != float64(42) {
		t.Errorf("AdditionalProperties[num] = %v, want 42", decoded.AdditionalProperties["num"])
	}
	if decoded.AdditionalProperties["flag"] != true {
		t.Errorf("AdditionalProperties[flag] = %v, want true", decoded.AdditionalProperties["flag"])
	}
}

// TestArrayContainerAdditionalPropsNotMixed verifies that known fields do not
// appear in the additional properties map after unmarshal.
func TestArrayContainerAdditionalPropsNotMixed(t *testing.T) {
	input := `{"values":["a"],"unknown_key":"surprise"}`

	var decoded ArrayContainer
	if err := json.Unmarshal([]byte(input), &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(decoded.Values) != 1 || decoded.Values[0] != "a" {
		t.Errorf("Values = %v, want [a]", decoded.Values)
	}
	if decoded.AdditionalProperties["unknown_key"] != "surprise" {
		t.Errorf("AdditionalProperties[unknown_key] = %v, want %q", decoded.AdditionalProperties["unknown_key"], "surprise")
	}
	if _, ok := decoded.AdditionalProperties["values"]; ok {
		t.Error("values should not be in AdditionalProperties")
	}
}

// TestArrayContainerEmptyValues verifies marshaling with nil/empty values slice.
func TestArrayContainerEmptyValues(t *testing.T) {
	ac := ArrayContainer{}

	data, err := json.Marshal(ac)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// The values field should still be present (marshaled as null)
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal raw failed: %v", err)
	}

	if _, ok := raw["values"]; !ok {
		t.Error("values key should be present in marshaled JSON")
	}
}

// TestApplyDefaults verifies ApplyDefaults does not panic.
func TestApplyDefaults(t *testing.T) {
	ac := &ArrayContainer{}
	ac.ApplyDefaults()
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
