package output

import (
	"encoding/json"
	"testing"
)

// TestAnyOfParamTestType verifies the Test anyOf union type with two
// inline object schemas.
// V2 test suite: internal/test/components/anyof/param
func TestAnyOfParamTestType(t *testing.T) {
	variant0 := TestAnyOf0{
		Item1: "value1",
		Item2: "value2",
	}

	test := Test{
		TestAnyOf0: &variant0,
	}

	if test.TestAnyOf0 == nil {
		t.Fatal("TestAnyOf0 should not be nil")
	}
	if test.TestAnyOf0.Item1 != "value1" {
		t.Errorf("Item1 = %q, want %q", test.TestAnyOf0.Item1, "value1")
	}
	if test.TestAnyOf0.Item2 != "value2" {
		t.Errorf("Item2 = %q, want %q", test.TestAnyOf0.Item2, "value2")
	}
	if test.TestAnyOf1 != nil {
		t.Error("TestAnyOf1 should be nil")
	}
}

// TestAnyOfParamTestAnyOf0 verifies the first anyOf variant with required
// fields.
func TestAnyOfParamTestAnyOf0(t *testing.T) {
	v := TestAnyOf0{
		Item1: "a",
		Item2: "b",
	}

	if v.Item1 != "a" {
		t.Errorf("Item1 = %q, want %q", v.Item1, "a")
	}
	if v.Item2 != "b" {
		t.Errorf("Item2 = %q, want %q", v.Item2, "b")
	}
}

// TestAnyOfParamTestAnyOf1 verifies the second anyOf variant with optional
// fields.
func TestAnyOfParamTestAnyOf1(t *testing.T) {
	item2 := "hello"
	item3 := "world"
	v := TestAnyOf1{
		Item2: &item2,
		Item3: &item3,
	}

	if *v.Item2 != "hello" {
		t.Errorf("Item2 = %q, want %q", *v.Item2, "hello")
	}
	if *v.Item3 != "world" {
		t.Errorf("Item3 = %q, want %q", *v.Item3, "world")
	}
}

// TestAnyOfParamTestMarshalJSON verifies that MarshalJSON merges fields from
// set anyOf members.
func TestAnyOfParamTestMarshalJSON(t *testing.T) {
	item2 := "shared"
	item3 := "extra"

	test := Test{
		TestAnyOf0: &TestAnyOf0{Item1: "first", Item2: "second"},
		TestAnyOf1: &TestAnyOf1{Item2: &item2, Item3: &item3},
	}

	data, err := json.Marshal(test)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal into map failed: %v", err)
	}

	// item1 and item3 should be present
	if m["item1"] != "first" {
		t.Errorf("item1 = %v, want %q", m["item1"], "first")
	}
	if m["item3"] != "extra" {
		t.Errorf("item3 = %v, want %q", m["item3"], "extra")
	}
}

// TestAnyOfParamTestUnmarshalJSON verifies that UnmarshalJSON populates
// matching anyOf members.
func TestAnyOfParamTestUnmarshalJSON(t *testing.T) {
	input := `{"item1":"a","item2":"b","item3":"c"}`

	var test Test
	if err := json.Unmarshal([]byte(input), &test); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Both variants should match since the JSON has fields from both
	if test.TestAnyOf0 == nil {
		t.Fatal("TestAnyOf0 should not be nil")
	}
	if test.TestAnyOf0.Item1 != "a" {
		t.Errorf("TestAnyOf0.Item1 = %q, want %q", test.TestAnyOf0.Item1, "a")
	}
	if test.TestAnyOf0.Item2 != "b" {
		t.Errorf("TestAnyOf0.Item2 = %q, want %q", test.TestAnyOf0.Item2, "b")
	}

	if test.TestAnyOf1 == nil {
		t.Fatal("TestAnyOf1 should not be nil")
	}
	if *test.TestAnyOf1.Item2 != "b" {
		t.Errorf("TestAnyOf1.Item2 = %q, want %q", *test.TestAnyOf1.Item2, "b")
	}
	if *test.TestAnyOf1.Item3 != "c" {
		t.Errorf("TestAnyOf1.Item3 = %q, want %q", *test.TestAnyOf1.Item3, "c")
	}
}

// TestAnyOfParamTest2Type verifies the Test2 anyOf union type with primitive
// (int and string) members.
func TestAnyOfParamTest2Type(t *testing.T) {
	intVal := 42
	t2 := Test2{
		Int0: &intVal,
	}

	if t2.Int0 == nil {
		t.Fatal("Int0 should not be nil")
	}
	if *t2.Int0 != 42 {
		t.Errorf("Int0 = %d, want %d", *t2.Int0, 42)
	}
	if t2.String1 != nil {
		t.Error("String1 should be nil")
	}
}

// TestAnyOfParamTest2MarshalInt verifies Test2 marshals when only the int
// member is set.
func TestAnyOfParamTest2MarshalInt(t *testing.T) {
	intVal := 99
	t2 := Test2{Int0: &intVal}

	data, err := json.Marshal(t2)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != "99" {
		t.Errorf("Marshal result = %s, want %q", string(data), "99")
	}
}

// TestAnyOfParamTest2MarshalString verifies Test2 marshals when only the
// string member is set.
func TestAnyOfParamTest2MarshalString(t *testing.T) {
	strVal := "hello"
	t2 := Test2{String1: &strVal}

	data, err := json.Marshal(t2)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != `"hello"` {
		t.Errorf("Marshal result = %s, want %q", string(data), `"hello"`)
	}
}

// TestAnyOfParamTest2MarshalBothSetError verifies that marshaling Test2 with
// both members set returns an error (exactly one must be set).
func TestAnyOfParamTest2MarshalBothSetError(t *testing.T) {
	intVal := 1
	strVal := "one"
	t2 := Test2{Int0: &intVal, String1: &strVal}

	_, err := json.Marshal(t2)
	if err == nil {
		t.Error("expected error when both members are set, got nil")
	}
}

// TestAnyOfParamTest2UnmarshalInt verifies that unmarshaling an integer value
// populates the Int0 member.
func TestAnyOfParamTest2UnmarshalInt(t *testing.T) {
	input := `42`

	var t2 Test2
	if err := json.Unmarshal([]byte(input), &t2); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if t2.Int0 == nil {
		t.Fatal("Int0 should not be nil")
	}
	if *t2.Int0 != 42 {
		t.Errorf("Int0 = %d, want %d", *t2.Int0, 42)
	}
}

// TestAnyOfParamTest2UnmarshalString verifies that unmarshaling a string value
// populates the String1 member.
func TestAnyOfParamTest2UnmarshalString(t *testing.T) {
	input := `"world"`

	var t2 Test2
	if err := json.Unmarshal([]byte(input), &t2); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if t2.String1 == nil {
		t.Fatal("String1 should not be nil")
	}
	if *t2.String1 != "world" {
		t.Errorf("String1 = %q, want %q", *t2.String1, "world")
	}
}

// TestAnyOfParamGetTestParameterAlias verifies the GetTestParameter type alias
// is a slice of Test2.
func TestAnyOfParamGetTestParameterAlias(t *testing.T) {
	intVal := 10
	var params GetTestParameter
	params = append(params, Test2{Int0: &intVal})

	if len(params) != 1 {
		t.Fatalf("params length = %d, want 1", len(params))
	}
	if *params[0].Int0 != 10 {
		t.Errorf("params[0].Int0 = %d, want %d", *params[0].Int0, 10)
	}
}

// TestAnyOfParamApplyDefaults verifies that ApplyDefaults can be called on
// all types without panic.
func TestAnyOfParamApplyDefaults(t *testing.T) {
	test := &Test{}
	test.ApplyDefaults()

	v0 := &TestAnyOf0{}
	v0.ApplyDefaults()

	v1 := &TestAnyOf1{}
	v1.ApplyDefaults()

	t2 := &Test2{}
	t2.ApplyDefaults()
}
