package spec_ext

import (
	"encoding/json"
	"testing"
)

func TestCamelSchemaInstantiation(t *testing.T) {
	id := "abc"
	s := CamelSchema{ID: &id}
	if s.ID == nil || *s.ID != "abc" {
		t.Errorf("unexpected ID: %v", s.ID)
	}
}

func TestPascalSchemaInstantiation(t *testing.T) {
	id := "xyz"
	s := PascalSchema{ID: &id}
	if s.ID == nil || *s.ID != "xyz" {
		t.Errorf("unexpected ID: %v", s.ID)
	}
}

func TestFooInstantiation(t *testing.T) {
	attr := "internal"
	camel := &CamelSchema{ID: nil}
	pascal := &PascalSchema{ID: nil}
	f := Foo{
		InternalAttr: &attr,
		CamelSchema:  camel,
		PascalSchema: pascal,
	}
	if f.InternalAttr == nil || *f.InternalAttr != "internal" {
		t.Errorf("unexpected InternalAttr: %v", f.InternalAttr)
	}
}

func TestFooJSONRoundTrip(t *testing.T) {
	id := "123"
	f := Foo{
		CamelSchema: &CamelSchema{ID: &id},
	}
	data, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded Foo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.CamelSchema == nil || decoded.CamelSchema.ID == nil || *decoded.CamelSchema.ID != "123" {
		t.Errorf("round-trip failed")
	}
}

func TestApplyDefaults(t *testing.T) {
	c := &CamelSchema{}
	c.ApplyDefaults()
	p := &PascalSchema{}
	p.ApplyDefaults()
	f := &Foo{}
	f.ApplyDefaults()
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
