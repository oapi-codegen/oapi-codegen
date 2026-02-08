package output

import (
	"encoding/json"
	"testing"
)

// TestFooInstantiation verifies that the Foo type can be created and fields accessed.
func TestFooInstantiation(t *testing.T) {
	val := "hello"
	foo := Foo{
		Field1: &val,
	}

	if foo.Field1 == nil {
		t.Fatal("Field1 should not be nil")
	}
	if *foo.Field1 != "hello" {
		t.Errorf("Field1 = %q, want %q", *foo.Field1, "hello")
	}
}

// TestBarInstantiation verifies that the Bar type can be created and fields accessed.
func TestBarInstantiation(t *testing.T) {
	val := "world"
	bar := Bar{
		Field2: &val,
	}

	if bar.Field2 == nil {
		t.Fatal("Field2 should not be nil")
	}
	if *bar.Field2 != "world" {
		t.Errorf("Field2 = %q, want %q", *bar.Field2, "world")
	}
}

// TestFooJSONRoundTrip verifies JSON marshal/unmarshal for Foo.
func TestFooJSONRoundTrip(t *testing.T) {
	val := "test"
	original := Foo{Field1: &val}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Foo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Field1 == nil {
		t.Fatal("Field1 should not be nil after round trip")
	}
	if *decoded.Field1 != "test" {
		t.Errorf("Field1 = %q, want %q", *decoded.Field1, "test")
	}
}

// TestBarJSONRoundTrip verifies JSON marshal/unmarshal for Bar.
func TestBarJSONRoundTrip(t *testing.T) {
	val := "test"
	original := Bar{Field2: &val}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Bar
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Field2 == nil {
		t.Fatal("Field2 should not be nil after round trip")
	}
	if *decoded.Field2 != "test" {
		t.Errorf("Field2 = %q, want %q", *decoded.Field2, "test")
	}
}

// TestApplyDefaults verifies ApplyDefaults does not panic on any type.
func TestApplyDefaults(t *testing.T) {
	foo := &Foo{}
	foo.ApplyDefaults()

	bar := &Bar{}
	bar.ApplyDefaults()
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
