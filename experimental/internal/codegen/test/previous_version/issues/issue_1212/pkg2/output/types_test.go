package output

import (
	"encoding/json"
	"testing"
)

// TestFooBarConstruction verifies that Foo and Bar types can be constructed
// and their optional fields accessed correctly.
// https://github.com/oapi-codegen/oapi-codegen/issues/1212
func TestFooBarConstruction(t *testing.T) {
	f1 := "value1"
	foo := Foo{Field1: &f1}
	if foo.Field1 == nil || *foo.Field1 != "value1" {
		t.Errorf("Foo.Field1 = %v, want %q", foo.Field1, "value1")
	}

	f2 := "value2"
	bar := Bar{Field2: &f2}
	if bar.Field2 == nil || *bar.Field2 != "value2" {
		t.Errorf("Bar.Field2 = %v, want %q", bar.Field2, "value2")
	}

	// Fields are optional (pointer)
	fooNil := Foo{}
	if fooNil.Field1 != nil {
		t.Errorf("Foo.Field1 should be nil, got %v", fooNil.Field1)
	}
}

func TestFooBarJSONRoundTrip(t *testing.T) {
	f1 := "hello"
	original := Foo{Field1: &f1}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal Foo failed: %v", err)
	}

	var decoded Foo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal Foo failed: %v", err)
	}

	if decoded.Field1 == nil || *decoded.Field1 != "hello" {
		t.Errorf("decoded Foo.Field1 = %v, want %q", decoded.Field1, "hello")
	}

	f2 := "world"
	originalBar := Bar{Field2: &f2}

	data, err = json.Marshal(originalBar)
	if err != nil {
		t.Fatalf("Marshal Bar failed: %v", err)
	}

	var decodedBar Bar
	if err := json.Unmarshal(data, &decodedBar); err != nil {
		t.Fatalf("Unmarshal Bar failed: %v", err)
	}

	if decodedBar.Field2 == nil || *decodedBar.Field2 != "world" {
		t.Errorf("decoded Bar.Field2 = %v, want %q", decodedBar.Field2, "world")
	}
}

func TestApplyDefaults(t *testing.T) {
	foo := &Foo{}
	foo.ApplyDefaults() // should not panic

	bar := &Bar{}
	bar.ApplyDefaults() // should not panic
}

func TestGetOpenAPISpecJSON(t *testing.T) {
	data, err := GetOpenAPISpecJSON()
	if err != nil {
		t.Fatalf("GetOpenAPISpecJSON() failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("GetOpenAPISpecJSON() returned empty data")
	}
}
