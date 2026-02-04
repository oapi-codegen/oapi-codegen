package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToTypeName(t *testing.T) {
	c := NewNameConverter(DefaultNameMangling(), NameSubstitutions{})

	tests := []struct {
		input    string
		expected string
	}{
		// Basic conversions
		{"foo", "Foo"},
		{"fooBar", "FooBar"},
		{"foo_bar", "FooBar"},
		{"foo-bar", "FooBar"},
		{"foo.bar", "FooBar"},

		// Names starting with numbers
		{"123", "N123"},
		{"123foo", "N123Foo"},
		{"1param", "N1Param"},

		// Names starting with special characters
		{"$ref", "DollarSignRef"},
		{"$", "DollarSign"},
		{"-1", "Minus1"},
		{"+1", "Plus1"},
		{"&now", "AndNow"},
		{"#tag", "HashTag"},
		{".hidden", "DotHidden"},
		{"@timestamp", "AtTimestamp"},
		{"_private", "UnderscorePrivate"},

		// Initialisms
		{"userId", "UserID"},
		{"httpUrl", "HTTPURL"},
		{"apiId", "APIID"},
		{"jsonData", "JSONData"},
		{"xmlParser", "XMLParser"},
		{"getHttpResponse", "GetHTTPResponse"},

		// Go keywords (still PascalCase, just prefixed)
		{"type", "_Type"},
		{"interface", "_Interface"},
		{"map", "_Map"},
		{"chan", "_Chan"},

		// Predeclared identifiers (still PascalCase, just prefixed)
		{"string", "_String"},
		{"int", "_Int"},
		{"error", "_Error"},
		{"nil", "_Nil"},

		// Edge cases
		{"", "Empty"},
		{"a", "A"},
		{"A", "A"},
		{"ABC", "ABC"},
		{"myXMLParser", "MyXMLParser"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := c.ToTypeName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToPropertyName(t *testing.T) {
	c := NewNameConverter(DefaultNameMangling(), NameSubstitutions{})

	tests := []struct {
		input    string
		expected string
	}{
		{"user_id", "UserID"},
		{"created_at", "CreatedAt"},
		{"is_active", "IsActive"},
		{"123field", "N123Field"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := c.ToPropertyName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToVariableName(t *testing.T) {
	c := NewNameConverter(DefaultNameMangling(), NameSubstitutions{})

	tests := []struct {
		input    string
		expected string
	}{
		{"Foo", "foo"},
		{"FooBar", "fooBar"},
		{"user_id", "userID"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := c.ToVariableName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNameSubstitutions(t *testing.T) {
	c := NewNameConverter(DefaultNameMangling(), NameSubstitutions{
		TypeNames: map[string]string{
			"foo": "MyCustomFoo",
		},
		PropertyNames: map[string]string{
			"bar": "MyCustomBar",
		},
	})

	assert.Equal(t, "MyCustomFoo", c.ToTypeName("foo"))
	assert.Equal(t, "MyCustomBar", c.ToPropertyName("bar"))

	// Non-substituted names still work normally
	assert.Equal(t, "Baz", c.ToTypeName("baz"))
}

func TestCustomNameMangling(t *testing.T) {
	// Custom config that changes numeric prefix
	mangling := DefaultNameMangling()
	mangling.NumericPrefix = "Num"

	c := NewNameConverter(mangling, NameSubstitutions{})

	assert.Equal(t, "Num123", c.ToTypeName("123"))
	assert.Equal(t, "Num1Foo", c.ToTypeName("1foo"))
}

func TestMergeNameMangling(t *testing.T) {
	defaults := DefaultNameMangling()

	// User wants to change just the numeric prefix
	user := NameMangling{
		NumericPrefix: "Number",
	}

	merged := defaults.Merge(user)

	// User value overrides
	assert.Equal(t, "Number", merged.NumericPrefix)

	// Defaults preserved
	assert.Equal(t, defaults.KeywordPrefix, merged.KeywordPrefix)
	assert.Equal(t, defaults.WordSeparators, merged.WordSeparators)
	assert.Equal(t, len(defaults.CharacterSubstitutions), len(merged.CharacterSubstitutions))
}

func TestMergeCharacterSubstitutions(t *testing.T) {
	defaults := DefaultNameMangling()

	// User wants to override $ and add a new one
	user := NameMangling{
		CharacterSubstitutions: map[string]string{
			"$": "Dollar", // Override default "DollarSign"
			"€": "Euro",   // Add new
		},
	}

	merged := defaults.Merge(user)

	assert.Equal(t, "Dollar", merged.CharacterSubstitutions["$"])
	assert.Equal(t, "Euro", merged.CharacterSubstitutions["€"])
	assert.Equal(t, "Minus", merged.CharacterSubstitutions["-"]) // Default preserved
}

func TestRemoveCharacterSubstitution(t *testing.T) {
	defaults := DefaultNameMangling()

	// User wants to remove $ substitution (empty string = remove)
	user := NameMangling{
		CharacterSubstitutions: map[string]string{
			"$": "", // Remove
		},
	}

	merged := defaults.Merge(user)

	_, exists := merged.CharacterSubstitutions["$"]
	assert.False(t, exists)
}
