package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeduplicateNames(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no duplicates",
			input:    []string{"Red", "Green", "Blue"},
			expected: []string{"Red", "Green", "Blue"},
		},
		{
			name:     "two duplicates",
			input:    []string{"Foo", "Bar", "Foo"},
			expected: []string{"Foo0", "Bar", "Foo1"},
		},
		{
			name:     "three duplicates",
			input:    []string{"X", "X", "X"},
			expected: []string{"X0", "X1", "X2"},
		},
		{
			name:     "empty list",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "single element",
			input:    []string{"Solo"},
			expected: []string{"Solo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deduplicateNames(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestComputeEnumConstantNames(t *testing.T) {
	converter := NewNameConverter(DefaultNameMangling(), NameSubstitutions{})

	t.Run("string enum with NameConverter pipeline", func(t *testing.T) {
		info := &EnumInfo{
			TypeName: "Color",
			BaseType: "string",
			Values:   []string{"red", "green", "blue"},
		}
		computeEnumConstantNames([]*EnumInfo{info}, converter)
		assert.Equal(t, []string{"Red", "Green", "Blue"}, info.SanitizedNames)
	})

	t.Run("integer enum with numeric prefix", func(t *testing.T) {
		info := &EnumInfo{
			TypeName: "Priority",
			BaseType: "int",
			Values:   []string{"1", "2", "3"},
		}
		computeEnumConstantNames([]*EnumInfo{info}, converter)
		assert.Equal(t, []string{"N1", "N2", "N3"}, info.SanitizedNames)
	})

	t.Run("within-enum deduplication", func(t *testing.T) {
		info := &EnumInfo{
			TypeName: "Example",
			BaseType: "string",
			// "Foo Bar" and "Foo-Bar" both sanitize to "FooBar"
			Values: []string{"Foo Bar", "Foo-Bar"},
		}
		computeEnumConstantNames([]*EnumInfo{info}, converter)
		assert.Equal(t, []string{"FooBar0", "FooBar1"}, info.SanitizedNames)
	})

	t.Run("custom names override sanitization", func(t *testing.T) {
		info := &EnumInfo{
			TypeName:    "Status",
			BaseType:    "string",
			Values:      []string{"active", "inactive"},
			CustomNames: []string{"On", "Off"},
		}
		computeEnumConstantNames([]*EnumInfo{info}, converter)
		assert.Equal(t, []string{"On", "Off"}, info.SanitizedNames)
	})

	t.Run("partial custom names", func(t *testing.T) {
		info := &EnumInfo{
			TypeName:    "Status",
			BaseType:    "string",
			Values:      []string{"active", "inactive"},
			CustomNames: []string{"On"}, // only first is custom
		}
		computeEnumConstantNames([]*EnumInfo{info}, converter)
		assert.Equal(t, []string{"On", "Inactive"}, info.SanitizedNames)
	})
}

func TestResolveEnumCollisions(t *testing.T) {
	t.Run("no collisions - no prefix", func(t *testing.T) {
		infos := []*EnumInfo{
			{TypeName: "Color", SanitizedNames: []string{"Red", "Green", "Blue"}},
			{TypeName: "Size", SanitizedNames: []string{"Small", "Medium", "Large"}},
		}
		resolveEnumCollisions(infos, map[string]bool{}, false)
		assert.False(t, infos[0].PrefixTypeName)
		assert.False(t, infos[1].PrefixTypeName)
	})

	t.Run("cross-enum collision - both prefixed", func(t *testing.T) {
		infos := []*EnumInfo{
			{TypeName: "Status1", SanitizedNames: []string{"Active", "Inactive"}},
			{TypeName: "Status2", SanitizedNames: []string{"Active", "Pending"}},
		}
		resolveEnumCollisions(infos, map[string]bool{}, false)
		assert.True(t, infos[0].PrefixTypeName, "Status1 should be prefixed due to 'Active' collision")
		assert.True(t, infos[1].PrefixTypeName, "Status2 should be prefixed due to 'Active' collision")
	})

	t.Run("type-name collision - constant matches non-enum type", func(t *testing.T) {
		infos := []*EnumInfo{
			{TypeName: "Role", SanitizedNames: []string{"Admin", "User"}},
		}
		allTypeNames := map[string]bool{"Admin": true} // Admin is a struct type
		resolveEnumCollisions(infos, allTypeNames, false)
		assert.True(t, infos[0].PrefixTypeName, "Role should be prefixed because 'Admin' is a type name")
	})

	t.Run("self-collision - constant matches own type name", func(t *testing.T) {
		infos := []*EnumInfo{
			{TypeName: "Bar", SanitizedNames: []string{"Foo", "Bar", "Baz"}},
		}
		resolveEnumCollisions(infos, map[string]bool{}, false)
		assert.True(t, infos[0].PrefixTypeName, "Bar should be prefixed because 'Bar' is its own type name")
	})

	t.Run("alwaysPrefix forces all enums prefixed", func(t *testing.T) {
		infos := []*EnumInfo{
			{TypeName: "Color", SanitizedNames: []string{"Red", "Green", "Blue"}},
			{TypeName: "Size", SanitizedNames: []string{"Small", "Medium", "Large"}},
		}
		resolveEnumCollisions(infos, map[string]bool{}, true)
		assert.True(t, infos[0].PrefixTypeName)
		assert.True(t, infos[1].PrefixTypeName)
	})

	t.Run("post-prefix collision cascades", func(t *testing.T) {
		// Enum1 has values "One","Two","Three" -> One, Two, Three
		// Enum2 has values "Two","Three","Four" -> Two, Three, Four
		// Enum3 has values "Enum1One","Foo","Bar" -> Enum1One, Foo, Bar
		//
		// Pass 1: Enum1 and Enum2 collide on "Two","Three" -> both get prefixed
		//   Enum1: Enum1One, Enum1Two, Enum1Three
		//   Enum2: Enum2Two, Enum2Three, Enum2Four
		//   Enum3: Enum1One, Foo, Bar (unprefixed)
		// Pass 2: Enum1's "Enum1One" collides with Enum3's "Enum1One" -> Enum3 gets prefixed
		//   Enum3: Enum3Enum1One, Enum3Foo, Enum3Bar
		infos := []*EnumInfo{
			{TypeName: "Enum1", SanitizedNames: []string{"One", "Two", "Three"}},
			{TypeName: "Enum2", SanitizedNames: []string{"Two", "Three", "Four"}},
			{TypeName: "Enum3", SanitizedNames: []string{"Enum1One", "Foo", "Bar"}},
		}
		resolveEnumCollisions(infos, map[string]bool{}, false)
		assert.True(t, infos[0].PrefixTypeName, "Enum1 should be prefixed (shares Two/Three with Enum2)")
		assert.True(t, infos[1].PrefixTypeName, "Enum2 should be prefixed (shares Two/Three with Enum1)")
		assert.True(t, infos[2].PrefixTypeName, "Enum3 should be prefixed (Enum1One collides with Enum1's prefixed name)")
	})

	t.Run("unrelated enum not affected by collision", func(t *testing.T) {
		infos := []*EnumInfo{
			{TypeName: "Status1", SanitizedNames: []string{"Active", "Inactive"}},
			{TypeName: "Status2", SanitizedNames: []string{"Active", "Pending"}},
			{TypeName: "Color", SanitizedNames: []string{"Red", "Green", "Blue"}},
		}
		resolveEnumCollisions(infos, map[string]bool{}, false)
		assert.True(t, infos[0].PrefixTypeName)
		assert.True(t, infos[1].PrefixTypeName)
		assert.False(t, infos[2].PrefixTypeName, "Color should not be affected by Status collision")
	})
}
