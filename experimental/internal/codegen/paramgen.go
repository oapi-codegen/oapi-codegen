package codegen

import (
	"fmt"
	"sort"

	"github.com/oapi-codegen/oapi-codegen/experimental/internal/codegen/templates"
)

// ParamUsageTracker tracks which parameter styling and binding functions
// are needed based on the OpenAPI spec being processed.
type ParamUsageTracker struct {
	// usedStyles tracks which style/explode combinations are used.
	// Keys are formatted as "style_{style}" or "style_{style}_explode" for serialization,
	// and "bind_{style}" or "bind_{style}_explode" for binding.
	usedStyles map[string]bool
}

// NewParamUsageTracker creates a new ParamUsageTracker.
func NewParamUsageTracker() *ParamUsageTracker {
	return &ParamUsageTracker{
		usedStyles: make(map[string]bool),
	}
}

// RecordStyleParam records that a style/explode combination is used for serialization.
// This is typically called when processing client parameters.
func (t *ParamUsageTracker) RecordStyleParam(style string, explode bool) {
	key := templates.ParamStyleKey("style_", style, explode)
	t.usedStyles[key] = true
}

// RecordBindParam records that a style/explode combination is used for binding.
// This is typically called when processing server parameters.
func (t *ParamUsageTracker) RecordBindParam(style string, explode bool) {
	key := templates.ParamStyleKey("bind_", style, explode)
	t.usedStyles[key] = true
}

// RecordParam records both style and bind usage for a parameter.
// Use this when generating both client and server code.
func (t *ParamUsageTracker) RecordParam(style string, explode bool) {
	t.RecordStyleParam(style, explode)
	t.RecordBindParam(style, explode)
}

// HasAnyUsage returns true if any parameter functions are needed.
func (t *ParamUsageTracker) HasAnyUsage() bool {
	return len(t.usedStyles) > 0
}

// GetRequiredTemplates returns the list of templates needed based on usage.
// The helpers template is always included first if any functions are needed.
func (t *ParamUsageTracker) GetRequiredTemplates() []templates.ParamTemplate {
	if !t.HasAnyUsage() {
		return nil
	}

	var result []templates.ParamTemplate

	// Always include helpers first
	result = append(result, templates.ParamHelpersTemplate)

	// Get all used style keys and sort them for deterministic output
	keys := make([]string, 0, len(t.usedStyles))
	for key := range t.usedStyles {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Add each required template
	for _, key := range keys {
		tmpl, ok := templates.ParamTemplates[key]
		if !ok {
			// This shouldn't happen if keys are properly validated
			continue
		}
		result = append(result, tmpl)
	}

	return result
}

// GetRequiredImports returns all imports needed for the used parameter functions.
// This aggregates imports from the helpers template and all used templates.
func (t *ParamUsageTracker) GetRequiredImports() []templates.Import {
	if !t.HasAnyUsage() {
		return nil
	}

	// Use a map to deduplicate imports
	importSet := make(map[string]templates.Import)

	// Add helpers imports
	for _, imp := range templates.ParamHelpersTemplate.Imports {
		importSet[imp.Path] = imp
	}

	// Add imports from each used template
	for key := range t.usedStyles {
		tmpl, ok := templates.ParamTemplates[key]
		if !ok {
			continue
		}
		for _, imp := range tmpl.Imports {
			importSet[imp.Path] = imp
		}
	}

	// Convert to sorted slice
	result := make([]templates.Import, 0, len(importSet))
	for _, imp := range importSet {
		result = append(result, imp)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Path < result[j].Path
	})

	return result
}

// GetUsedStyleKeys returns the sorted list of used style keys for debugging.
func (t *ParamUsageTracker) GetUsedStyleKeys() []string {
	keys := make([]string, 0, len(t.usedStyles))
	for key := range t.usedStyles {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

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
