package output

import (
	"testing"
)

// TestIllegalEnumNames verifies that enum constants with various edge case values
// are generated with valid Go identifiers.
// The enum type "Bar" has a value "Bar" which matches its own type name,
// triggering the self-collision rule. This causes all constants to be prefixed
// with the type name.
func TestIllegalEnumNames(t *testing.T) {
	// All these enum constants should exist and have valid Go names.
	// Because "Bar" appears as both a type name and an enum value,
	// all constants are prefixed with "Bar" (the type name).
	tests := []struct {
		name     string
		constant Bar
		value    string
	}{
		{"empty string", BarEmpty, ""},
		{"Foo", BarFoo, "Foo"},
		{"Bar (self-collision)", BarBar, "Bar"},
		{"Foo Bar (with space)", BarFooBar0, "Foo Bar"},
		{"Foo-Bar (with hyphen)", BarFooBar1, "Foo-Bar"},
		{"1Foo (leading digit)", BarN1Foo, "1Foo"},
		{" Foo (leading space)", BarXFoo0, " Foo"},
		{" Foo  (leading and trailing space)", BarXFoo1, " Foo "},
		{"_Foo_ (underscores)", BarUnderscoreFoo, "_Foo_"},
		{"1 (just digit)", BarN1, "1"},
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
		BarFoo,
		BarBar,
		BarEmpty, // empty string
	}

	if len(response) != 3 {
		t.Errorf("response length = %d, want 3", len(response))
	}
	if response[0] != "Foo" {
		t.Errorf("response[0] = %q, want %q", response[0], "Foo")
	}
}
