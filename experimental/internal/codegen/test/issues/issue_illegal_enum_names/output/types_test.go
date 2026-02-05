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
		{"empty string", Bar_Value, ""},
		{"Foo", Bar_Foo, "Foo"},
		{"Bar", Bar_Bar, "Bar"},
		{"Foo Bar (with space)", Bar_Foo_Bar, "Foo Bar"},
		{"Foo-Bar (with hyphen)", Bar_Foo_Bar_1, "Foo-Bar"},
		{"1Foo (leading digit)", Bar_Foo_1, "1Foo"},
		{" Foo (leading space)", Bar__Foo, " Foo"},
		{" Foo  (leading and trailing space)", Bar__Foo_, " Foo "},
		{"_Foo_ (underscores)", Bar__Foo__1, "_Foo_"},
		{"1 (just digit)", Bar_Value_1, "1"},
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
		Bar_Foo,
		Bar_Bar,
		Bar_Value, // empty string
	}

	if len(response) != 3 {
		t.Errorf("response length = %d, want 3", len(response))
	}
	if response[0] != "Foo" {
		t.Errorf("response[0] = %q, want %q", response[0], "Foo")
	}
}
