package output

import (
	"testing"
)

// TestIllegalEnumNames verifies that enum constants with various edge case values
// are generated with valid Go identifiers.
func TestIllegalEnumNames(t *testing.T) {
	// All these enum constants should exist and have valid Go names
	tests := []struct {
		name     string
		constant BarSchemaComponent
		value    string
	}{
		{"empty string", BarSchemaComponent_Value, ""},
		{"Foo", BarSchemaComponent_Foo, "Foo"},
		{"Bar", BarSchemaComponent_Bar, "Bar"},
		{"Foo Bar (with space)", BarSchemaComponent_Foo_Bar, "Foo Bar"},
		{"Foo-Bar (with hyphen)", BarSchemaComponent_Foo_Bar_1, "Foo-Bar"},
		{"1Foo (leading digit)", BarSchemaComponent_Foo_1, "1Foo"},
		{" Foo (leading space)", BarSchemaComponent__Foo, " Foo"},
		{" Foo  (leading and trailing space)", BarSchemaComponent__Foo_, " Foo "},
		{"_Foo_ (underscores)", BarSchemaComponent__Foo__1, "_Foo_"},
		{"1 (just digit)", BarSchemaComponent_Value_1, "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.value {
				t.Errorf("constant %q = %q, want %q", tt.name, tt.constant, tt.value)
			}
		})
	}
}

func TestBarCanBeUsedInSlice(t *testing.T) {
	// The response type is []Bar
	response := GetFooJSONResponse{
		BarSchemaComponent_Foo,
		BarSchemaComponent_Bar,
		BarSchemaComponent_Value, // empty string
	}

	if len(response) != 3 {
		t.Errorf("response length = %d, want 3", len(response))
	}
	if response[0] != "Foo" {
		t.Errorf("response[0] = %q, want %q", response[0], "Foo")
	}
}
