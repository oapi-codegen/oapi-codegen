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

		// Go keywords - PascalCase doesn't conflict with lowercase keywords
		{"type", "Type"},
		{"interface", "Interface"},
		{"map", "Map"},
		{"chan", "Chan"},

		// Predeclared identifiers - PascalCase doesn't conflict with lowercase identifiers
		{"string", "String"},
		{"int", "Int"},
		{"error", "Error"},
		{"nil", "Nil"},

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

func TestToEnumValueName(t *testing.T) {
	c := NewNameConverter(DefaultNameMangling(), NameSubstitutions{})

	tests := []struct {
		name     string
		value    string
		baseType string
		expected string
	}{
		// String enum: basic conversions (uses toGoIdentifier pipeline)
		{"camelCase", "fooBar", "string", "FooBar"},
		{"word separator hyphen", "foo-bar", "string", "FooBar"},
		{"word separator space", "foo bar", "string", "FooBar"},
		{"word separator dot", "foo.bar", "string", "FooBar"},

		// String enum: character substitutions at start
		{"dollar sign", "$ref", "string", "DollarSignRef"},
		{"underscore prefix", "_foo", "string", "UnderscoreFoo"},

		// String enum: initialisms
		{"initialism ID", "userId", "string", "UserID"},
		{"initialism HTTP", "httpUrl", "string", "HTTPURL"},

		// String enum: empty string
		{"empty string", "", "string", "Empty"},

		// String enum: numeric prefix
		{"leading digit string", "1foo", "string", "N1Foo"},
		{"all digits string", "123", "string", "N123"},

		// Integer enum: numeric values
		{"positive int", "42", "int", "N42"},
		{"negative int", "-5", "int", "Minus5"},
		{"large int", "1000", "int64", "N1000"},
		{"negative int32", "-100", "int32", "Minus100"},

		// Integer enum: zero
		{"zero", "0", "int", "N0"},

		// String enum: special characters become word boundaries
		// Leading space has no char substitution → generic "X" prefix
		{"spaces around", " Foo ", "string", "XFoo"},
		// Underscore at start gets substitution, trailing underscore is a word separator (skipped)
		{"underscores around", "_Foo_", "string", "UnderscoreFoo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.ToEnumValueName(tt.value, tt.baseType)
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
