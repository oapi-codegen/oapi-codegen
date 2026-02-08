package codegen

import (
	"fmt"
)

// DefaultParamStyle returns the default style for a parameter location.
func DefaultParamStyle(location string) string {
	switch location {
	case "path", "header":
		return "simple"
	case "query", "cookie":
		return "form"
	default:
		return "form"
	}
}

// DefaultParamExplode returns the default explode value for a parameter location.
func DefaultParamExplode(location string) bool {
	switch location {
	case "path", "header":
		return false
	case "query", "cookie":
		return true
	default:
		return false
	}
}

// ValidateParamStyle validates that a style is supported for a location.
// Returns an error if the combination is invalid.
func ValidateParamStyle(style, location string) error {
	validStyles := map[string][]string{
		"path":   {"simple", "label", "matrix"},
		"query":  {"form", "spaceDelimited", "pipeDelimited", "deepObject"},
		"header": {"simple"},
		"cookie": {"form"},
	}

	allowed, ok := validStyles[location]
	if !ok {
		return fmt.Errorf("unknown parameter location: %s", location)
	}

	for _, s := range allowed {
		if s == style {
			return nil
		}
	}

	return fmt.Errorf("style '%s' is not valid for location '%s'; valid styles are: %v", style, location, allowed)
}
