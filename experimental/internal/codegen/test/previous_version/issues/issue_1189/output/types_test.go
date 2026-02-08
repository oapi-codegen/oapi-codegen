package output

import (
	"encoding/json"
	"testing"
)

// TestAnyOfOneOfComposition verifies that anyOf/oneOf composition generates
// correct union types with proper Marshal/Unmarshal behavior.
// https://github.com/oapi-codegen/oapi-codegen/issues/1189
func TestTestTypeInstantiation(t *testing.T) {
	str := "hello"
	test := Test{
		FieldA: &TestFieldA{
			String0: &str,
		},
		FieldB: &TestFieldB{},
		FieldC: &TestFieldC{
			String0: &str,
		},
	}

	if test.FieldA == nil {
		t.Fatal("FieldA should not be nil")
	}
	if *test.FieldA.String0 != "hello" {
		t.Errorf("FieldA.String0 = %q, want %q", *test.FieldA.String0, "hello")
	}
	if test.FieldB == nil {
		t.Fatal("FieldB should not be nil")
	}
	if test.FieldC == nil {
		t.Fatal("FieldC should not be nil")
	}
}

func TestFieldAWithStringAnyOf(t *testing.T) {
	// FieldA is anyOf: string or enum
	str := "plain-string"
	fa := TestFieldA{
		String0: &str,
	}

	data, err := json.Marshal(fa)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != `"plain-string"` {
		t.Errorf("Marshal result = %s, want %q", string(data), `"plain-string"`)
	}
}

func TestFieldAWithEnumAnyOf(t *testing.T) {
	// FieldA is anyOf: string or enum with values foo/bar
	enumVal := TestFieldAAnyOf1Foo
	fa := TestFieldA{
		TestFieldAAnyOf1: &enumVal,
	}

	data, err := json.Marshal(fa)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded map[string]any
	// For the enum variant, it marshals as a map (struct-like)
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		// If it's not a map, it might be a plain string
		var strVal string
		if err2 := json.Unmarshal(data, &strVal); err2 != nil {
			t.Fatalf("Cannot unmarshal as map or string: %v / %v", err, err2)
		}
	}
}

func TestFieldAAnyOf1EnumConstants(t *testing.T) {
	// Verify enum constants exist and have correct values
	if string(TestFieldAAnyOf1Foo) != "foo" {
		t.Errorf("TestFieldAAnyOf1Foo = %q, want %q", TestFieldAAnyOf1Foo, "foo")
	}
	if string(TestFieldAAnyOf1Bar) != "bar" {
		t.Errorf("TestFieldAAnyOf1Bar = %q, want %q", TestFieldAAnyOf1Bar, "bar")
	}
}

func TestFieldCOneOfString(t *testing.T) {
	// FieldC is oneOf: string or enum
	str := "one-of-string"
	fc := TestFieldC{
		String0: &str,
	}

	data, err := json.Marshal(fc)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != `"one-of-string"` {
		t.Errorf("Marshal result = %s, want %q", string(data), `"one-of-string"`)
	}
}

func TestFieldCOneOf1EnumConstants(t *testing.T) {
	// Verify enum constants exist and have correct values
	if string(TestFieldCOneOf1Foo) != "foo" {
		t.Errorf("TestFieldCOneOf1Foo = %q, want %q", TestFieldCOneOf1Foo, "foo")
	}
	if string(TestFieldCOneOf1Bar) != "bar" {
		t.Errorf("TestFieldCOneOf1Bar = %q, want %q", TestFieldCOneOf1Bar, "bar")
	}
}

func TestFieldBEmptyStruct(t *testing.T) {
	// FieldB is an empty struct
	fb := TestFieldB{}
	data, err := json.Marshal(fb)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := `{}`
	if string(data) != expected {
		t.Errorf("Marshal result = %s, want %s", string(data), expected)
	}
}

func TestTestJSONRoundTrip(t *testing.T) {
	str := "round-trip"
	original := Test{
		FieldA: &TestFieldA{
			String0: &str,
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Test
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.FieldA == nil {
		t.Fatal("FieldA should not be nil after round trip")
	}
}

func TestApplyDefaults(t *testing.T) {
	// ApplyDefaults should be callable on all types without panic
	test := &Test{}
	test.ApplyDefaults()

	fa := &TestFieldA{}
	fa.ApplyDefaults()

	fb := &TestFieldB{}
	fb.ApplyDefaults()

	fc := &TestFieldC{}
	fc.ApplyDefaults()
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
