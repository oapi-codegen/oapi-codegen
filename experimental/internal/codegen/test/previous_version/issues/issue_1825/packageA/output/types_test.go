package output

import (
	"encoding/json"
	"testing"
)

func TestObjectAInstantiation(t *testing.T) {
	name := "test"
	a := ObjectA{Name: &name}
	if a.Name == nil || *a.Name != "test" {
		t.Errorf("unexpected name: %v", a.Name)
	}
}

func TestObjectAJSONRoundTrip(t *testing.T) {
	name := "hello"
	a := ObjectA{Name: &name}
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded ObjectA
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Name == nil || *decoded.Name != "hello" {
		t.Errorf("round-trip failed: %v", decoded.Name)
	}
}

func TestApplyDefaults(t *testing.T) {
	a := &ObjectA{}
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
