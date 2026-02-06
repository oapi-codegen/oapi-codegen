package templates

import (
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var titleCaser = cases.Title(language.English)

// pathParamRE matches OpenAPI path parameters including styled variants.
// Matches: {param}, {param*}, {.param}, {.param*}, {;param}, {;param*}, {?param}, {?param*}
var pathParamRE = regexp.MustCompile(`{[.;?]?([^{}*]+)\*?}`)

// Funcs returns the template function map for server templates.
func Funcs() template.FuncMap {
	return template.FuncMap{
		"pathToStdHTTPPattern":  PathToStdHTTPPattern,
		"pathToChiPattern":      PathToChiPattern,
		"pathToEchoPattern":     PathToEchoPattern,
		"pathToGinPattern":      PathToGinPattern,
		"pathToGorillaPattern":  PathToGorillaPattern,
		"pathToFiberPattern":    PathToFiberPattern,
		"pathToIrisPattern":     PathToIrisPattern,
		"toGoIdentifier":        ToGoIdentifier,
		"lower":                 strings.ToLower,
		"title":                 titleCaser.String,
	}
}

// PathToStdHTTPPattern converts an OpenAPI path template to a Go 1.22+ std http pattern.
// OpenAPI: /users/{user_id}/posts/{post_id}
// StdHTTP: /users/{user_id}/posts/{post_id}
// Special case: "/" becomes "/{$}" to match only the root path.
func PathToStdHTTPPattern(path string) string {
	// https://pkg.go.dev/net/http#hdr-Patterns-ServeMux
	// The special wildcard {$} matches only the end of the URL.
	if path == "/" {
		return "/{$}"
	}
	return pathParamRE.ReplaceAllString(path, "{$1}")
}

// PathToChiPattern converts an OpenAPI path template to a Chi-compatible pattern.
// OpenAPI: /users/{user_id}/posts/{post_id}
// Chi: /users/{user_id}/posts/{post_id}
func PathToChiPattern(path string) string {
	return pathParamRE.ReplaceAllString(path, "{$1}")
}

// PathToEchoPattern converts an OpenAPI path template to an Echo-compatible pattern.
// OpenAPI: /users/{user_id}/posts/{post_id}
// Echo: /users/:user_id/posts/:post_id
func PathToEchoPattern(path string) string {
	return pathParamRE.ReplaceAllString(path, ":$1")
}

// PathToGinPattern converts an OpenAPI path template to a Gin-compatible pattern.
// OpenAPI: /users/{user_id}/posts/{post_id}
// Gin: /users/:user_id/posts/:post_id
func PathToGinPattern(path string) string {
	return pathParamRE.ReplaceAllString(path, ":$1")
}

// PathToGorillaPattern converts an OpenAPI path template to a Gorilla Mux-compatible pattern.
// OpenAPI: /users/{user_id}/posts/{post_id}
// Gorilla: /users/{user_id}/posts/{post_id}
func PathToGorillaPattern(path string) string {
	return pathParamRE.ReplaceAllString(path, "{$1}")
}

// PathToFiberPattern converts an OpenAPI path template to a Fiber-compatible pattern.
// OpenAPI: /users/{user_id}/posts/{post_id}
// Fiber: /users/:user_id/posts/:post_id
func PathToFiberPattern(path string) string {
	return pathParamRE.ReplaceAllString(path, ":$1")
}

// PathToIrisPattern converts an OpenAPI path template to an Iris-compatible pattern.
// OpenAPI: /users/{user_id}/posts/{post_id}
// Iris: /users/:user_id/posts/:post_id
func PathToIrisPattern(path string) string {
	return pathParamRE.ReplaceAllString(path, ":$1")
}

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
