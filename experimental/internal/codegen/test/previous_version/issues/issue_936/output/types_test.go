package output

import (
	"encoding/json"
	"testing"
)

// TestFilterColumnIncludesInstantiation verifies that recursive/circular schema references
// generate valid types.
// https://github.com/oapi-codegen/oapi-codegen/issues/936
func TestFilterColumnIncludesInstantiation(t *testing.T) {
	strVal := "hello"
	fv := FilterValue{
		String1: &strVal,
	}
	fp := FilterPredicate{
		FilterValue: &fv,
	}
	fci := FilterColumnIncludes{
		DollarSignIncludes: &fp,
	}

	if fci.DollarSignIncludes == nil {
		t.Fatal("DollarSignIncludes should not be nil")
	}
	if fci.DollarSignIncludes.FilterValue == nil {
		t.Fatal("FilterValue should not be nil")
	}
	if *fci.DollarSignIncludes.FilterValue.String1 != "hello" {
		t.Errorf("String1 = %q, want %q", *fci.DollarSignIncludes.FilterValue.String1, "hello")
	}
}

func TestFilterValueOneOfString(t *testing.T) {
	strVal := "test"
	fv := FilterValue{
		String1: &strVal,
	}

	data, err := json.Marshal(fv)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != `"test"` {
		t.Errorf("Marshal result = %s, want %q", string(data), `"test"`)
	}
}

func TestFilterValueOneOfFloat(t *testing.T) {
	floatVal := float32(3.14)
	fv := FilterValue{
		Float320: &floatVal,
	}

	data, err := json.Marshal(fv)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded float32
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if decoded != 3.14 {
		t.Errorf("decoded = %v, want %v", decoded, 3.14)
	}
}

func TestFilterValueOneOfBool(t *testing.T) {
	boolVal := true
	fv := FilterValue{
		Bool2: &boolVal,
	}

	data, err := json.Marshal(fv)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != "true" {
		t.Errorf("Marshal result = %s, want %s", string(data), "true")
	}
}

func TestFilterRangeValueOneOfString(t *testing.T) {
	strVal := "2023-01-01"
	frv := FilterRangeValue{
		String1: &strVal,
	}

	data, err := json.Marshal(frv)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != `"2023-01-01"` {
		t.Errorf("Marshal result = %s, want %q", string(data), `"2023-01-01"`)
	}
}

func TestFilterPredicateRangeOp(t *testing.T) {
	strVal := "100"
	fro := FilterPredicateRangeOp{
		DollarSignLt: &FilterRangeValue{
			String1: &strVal,
		},
	}

	data, err := json.Marshal(fro)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded FilterPredicateRangeOp
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.DollarSignLt == nil {
		t.Fatal("DollarSignLt should not be nil")
	}
	if *decoded.DollarSignLt.String1 != "100" {
		t.Errorf("String1 = %q, want %q", *decoded.DollarSignLt.String1, "100")
	}
}

func TestFilterColumnIncludesAdditionalProperties(t *testing.T) {
	fci := FilterColumnIncludes{
		AdditionalProperties: map[string]any{
			"customField": "customValue",
		},
	}

	data, err := json.Marshal(fci)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded FilterColumnIncludes
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.AdditionalProperties["customField"] != "customValue" {
		t.Errorf("AdditionalProperties[customField] = %v, want %q",
			decoded.AdditionalProperties["customField"], "customValue")
	}
}

func TestFilterColumnIncludesJSONRoundTrip(t *testing.T) {
	strVal := "match-me"
	original := FilterColumnIncludes{
		DollarSignIncludes: &FilterPredicate{
			FilterValue: &FilterValue{
				String1: &strVal,
			},
		},
		AdditionalProperties: map[string]any{
			"extra": "data",
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded FilterColumnIncludes
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.DollarSignIncludes == nil {
		t.Fatal("DollarSignIncludes should not be nil after round trip")
	}
	if decoded.AdditionalProperties["extra"] != "data" {
		t.Errorf("AdditionalProperties[extra] = %v, want %q",
			decoded.AdditionalProperties["extra"], "data")
	}
}

func TestFilterPredicateOpAdditionalProperties(t *testing.T) {
	op := FilterPredicateOp{
		AdditionalProperties: map[string]any{
			"$custom": "value",
		},
	}

	data, err := json.Marshal(op)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded FilterPredicateOp
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.AdditionalProperties["$custom"] != "value" {
		t.Errorf("AdditionalProperties[$custom] = %v, want %q",
			decoded.AdditionalProperties["$custom"], "value")
	}
}

func TestApplyDefaults(t *testing.T) {
	// ApplyDefaults should be callable on all types without panic
	fci := &FilterColumnIncludes{}
	fci.ApplyDefaults()

	fv := &FilterValue{}
	fv.ApplyDefaults()

	frv := &FilterRangeValue{}
	frv.ApplyDefaults()

	fp := &FilterPredicate{}
	fp.ApplyDefaults()

	fpo := &FilterPredicateOp{}
	fpo.ApplyDefaults()

	fpr := &FilterPredicateRangeOp{}
	fpr.ApplyDefaults()
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
