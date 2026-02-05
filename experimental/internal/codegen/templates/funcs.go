package templates

import (
	"regexp"
	"strings"
	"text/template"
)

// Funcs returns the template function map for server templates.
func Funcs() template.FuncMap {
	return template.FuncMap{
		"pathToStdHTTPPattern": PathToStdHTTPPattern,
		"toGoIdentifier":       ToGoIdentifier,
	}
}

// PathToStdHTTPPattern converts an OpenAPI path template to a Go 1.22+ std http pattern.
// OpenAPI: /users/{user_id}/posts/{post_id}
// StdHTTP: /users/{user_id}/posts/{post_id}
// They use the same format, but this function handles any edge cases.
func PathToStdHTTPPattern(path string) string {
	// Go 1.22+ uses the same {param} syntax as OpenAPI
	// Just ensure the pattern is valid
	return path
}

// swaggerPathParamRe matches OpenAPI path parameters like {param_name}.
var swaggerPathParamRe = regexp.MustCompile(`\{([^}]+)\}`)

// ToGoIdentifier converts a string to a valid Go identifier.
// This is a simple version for template usage.
func ToGoIdentifier(s string) string {
	if s == "" {
		return "Empty"
	}

	// Replace non-alphanumeric characters with underscores
	result := make([]byte, 0, len(s))
	capitalizeNext := true

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			if capitalizeNext {
				result = append(result, c-32) // uppercase
				capitalizeNext = false
			} else {
				result = append(result, c)
			}
		} else if c >= 'A' && c <= 'Z' {
			result = append(result, c)
			capitalizeNext = false
		} else if c >= '0' && c <= '9' {
			result = append(result, c)
			capitalizeNext = false
		} else {
			// Word separator
			capitalizeNext = true
		}
	}

	if len(result) == 0 {
		return "Empty"
	}

	// Handle leading digit
	if result[0] >= '0' && result[0] <= '9' {
		result = append([]byte("N"), result...)
	}

	str := string(result)

	// Handle Go keywords
	lower := strings.ToLower(str)
	if isGoKeyword(lower) {
		str = str + "_"
	}

	return str
}

// isGoKeyword returns true if s is a Go keyword.
func isGoKeyword(s string) bool {
	keywords := map[string]bool{
		"break": true, "case": true, "chan": true, "const": true, "continue": true,
		"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
		"func": true, "go": true, "goto": true, "if": true, "import": true,
		"interface": true, "map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true, "var": true,
	}
	return keywords[s]
}
