package output

import (
	"encoding/json"
	"testing"

	packageA "github.com/oapi-codegen/oapi-codegen-exp/experimental/internal/codegen/test/previous_version/issues/issue_1825/packageA/output"
)

func TestContainerInstantiation(t *testing.T) {
	name := "test"
	objA := &packageA.ObjectA{Name: &name}
	c := Container{
		ObjectA: objA,
	}
	if c.ObjectA == nil || c.ObjectA.Name == nil || *c.ObjectA.Name != "test" {
		t.Errorf("unexpected ObjectA: %v", c.ObjectA)
	}
}

func TestContainerJSONRoundTrip(t *testing.T) {
	name := "hello"
	objA := &packageA.ObjectA{Name: &name}
	c := Container{ObjectA: objA}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded Container
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.ObjectA == nil || decoded.ObjectA.Name == nil || *decoded.ObjectA.Name != "hello" {
		t.Errorf("round-trip failed: %v", decoded.ObjectA)
	}
}

func TestContainerObjectBField(t *testing.T) {
	// ObjectB is *any — should be settable to any JSON value
	val := any(map[string]any{"key": "value"})
	c := Container{ObjectB: &val}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded Container
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.ObjectB == nil {
		t.Fatal("ObjectB should not be nil")
	}
}

func TestApplyDefaults(t *testing.T) {
	c := &Container{}
	c.ApplyDefaults()
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
