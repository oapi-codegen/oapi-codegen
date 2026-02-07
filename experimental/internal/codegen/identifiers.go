package codegen

import (
	"strings"
	"unicode"
)

// Go keywords that can't be used as identifiers
var goKeywords = map[string]bool{
	"break":       true,
	"case":        true,
	"chan":        true,
	"const":       true,
	"continue":    true,
	"default":     true,
	"defer":       true,
	"else":        true,
	"fallthrough": true,
	"for":         true,
	"func":        true,
	"go":          true,
	"goto":        true,
	"if":          true,
	"import":      true,
	"interface":   true,
	"map":         true,
	"package":     true,
	"range":       true,
	"return":      true,
	"select":      true,
	"struct":      true,
	"switch":      true,
	"type":        true,
	"var":         true,
}

// IsGoKeyword returns true if s is a Go keyword.
func IsGoKeyword(s string) bool {
	return goKeywords[s]
}

// ToCamelCase converts a string to CamelCase (PascalCase).
// It treats hyphens, underscores, spaces, and other non-alphanumeric characters as word separators.
// Example: "user-name" -> "UserName", "user_id" -> "UserId"
func ToCamelCase(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder
	capitalizeNext := true

	for _, r := range s {
		if isWordSeparator(r) {
			capitalizeNext = true
			continue
		}

		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			capitalizeNext = true
			continue
		}

		if capitalizeNext {
			result.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// LowercaseFirstCharacter lowercases only the first character of a string.
// Example: "UserName" -> "userName"
func LowercaseFirstCharacter(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// UppercaseFirstCharacter uppercases only the first character of a string.
// Example: "userName" -> "UserName"
func UppercaseFirstCharacter(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// isWordSeparator returns true if the rune is a word separator.
func isWordSeparator(r rune) bool {
	return r == '-' || r == '_' || r == ' ' || r == '.' || r == '/'
}

// ToGoIdentifier converts a string to a valid Go identifier.
// It converts to CamelCase, handles leading digits, and avoids Go keywords.
func ToGoIdentifier(s string) string {
	result := ToCamelCase(s)

	// Handle empty result
	if result == "" {
		return "Empty"
	}

	// Handle leading digits
	if result[0] >= '0' && result[0] <= '9' {
		result = "N" + result
	}

	// Handle Go keywords - check both the original input and lowercase result
	// "type" -> "Type" but we still want to avoid "Type" being used as-is
	// since user might write it as lowercase in code
	if IsGoKeyword(s) || IsGoKeyword(strings.ToLower(result)) {
		result = result + "_"
	}

	return result
}
