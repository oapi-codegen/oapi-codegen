package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsGoKeyword(t *testing.T) {
	keywords := []string{
		"break", "case", "chan", "const", "continue",
		"default", "defer", "else", "fallthrough", "for",
		"func", "go", "goto", "if", "import",
		"interface", "map", "package", "range", "return",
		"select", "struct", "switch", "type", "var",
	}

	for _, kw := range keywords {
		t.Run(kw, func(t *testing.T) {
			assert.True(t, IsGoKeyword(kw), "%s should be a keyword", kw)
		})
	}

	nonKeywords := []string{
		"user", "name", "id", "Type", "Interface", "Map",
		"string", "int", "bool", "error", // predeclared but not keywords
	}

	for _, nkw := range nonKeywords {
		t.Run(nkw+"_not_keyword", func(t *testing.T) {
			assert.False(t, IsGoKeyword(nkw), "%s should not be a keyword", nkw)
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"user", "User"},
		{"user_name", "UserName"},
		{"user-name", "UserName"},
		{"user.name", "UserName"},
		{"user name", "UserName"},
		{"USER", "USER"},
		{"USER_NAME", "USERNAME"},
		{"123", "123"},
		{"user123", "User123"},
		{"user123name", "User123name"},
		{"get-users-by-id", "GetUsersById"},
		{"__private", "Private"},
		{"a_b_c", "ABC"},
		{"already_CamelCase", "AlreadyCamelCase"},
		{"path/to/resource", "PathToResource"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := ToCamelCase(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestLowercaseFirstCharacter(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"User", "user"},
		{"UserName", "userName"},
		{"user", "user"},
		{"ABC", "aBC"},
		{"123", "123"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := LowercaseFirstCharacter(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestUppercaseFirstCharacter(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"user", "User"},
		{"userName", "UserName"},
		{"User", "User"},
		{"abc", "Abc"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := UppercaseFirstCharacter(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestToGoIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user", "User"},
		{"user_name", "UserName"},
		{"123abc", "N123abc"},
		{"type", "Type_"},
		{"map", "Map_"},
		{"interface", "Interface_"},
		{"", "Empty"},
		{"get-users", "GetUsers"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := ToGoIdentifier(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
