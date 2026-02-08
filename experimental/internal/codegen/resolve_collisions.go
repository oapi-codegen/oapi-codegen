package codegen

import (
	"fmt"
	"strings"
)

// schemaContextSuffix maps a SchemaContext to a disambiguation suffix.
func schemaContextSuffix(ctx SchemaContext) string {
	switch ctx {
	case ContextComponentSchema:
		return "Schema"
	case ContextParameter:
		return "Parameter"
	case ContextRequestBody:
		return "Request"
	case ContextResponse:
		return "Response"
	case ContextHeader:
		return "Header"
	case ContextCallback:
		return "Callback"
	case ContextWebhook:
		return "Webhook"
	default:
		return ""
	}
}

// collisionGroupStrategy attempts to resolve a naming collision within a group
// of schemas that share the same candidate name. It returns true if any name
// was changed.
type collisionGroupStrategy func(
	group []*SchemaDescriptor,
	candidates map[*SchemaDescriptor]string,
	converter *NameConverter,
) bool

// disambiguationStrategy attempts to produce a new, more specific name for a
// single schema. It returns the new name and true if it found a disambiguation,
// or the original name and false otherwise.
type disambiguationStrategy func(
	s *SchemaDescriptor,
	currentName string,
	converter *NameConverter,
) (string, bool)

// disambiguationStrategies is the ordered list of per-schema sub-strategies
// tried by strategyPerSchemaDisambiguate.
var disambiguationStrategies = []disambiguationStrategy{
	tryContentTypeSuffix,
	tryStatusCodeSuffix,
	tryParamIndexSuffix,
	tryCompositionTypeSuffix,
}

// collisionStrategies is the ordered list of group-level strategies tried by
// resolveCollisions for each conflicting bucket.
var collisionStrategies = []collisionGroupStrategy{
	strategyContextSuffix,
	strategyPerSchemaDisambiguate,
	strategyNumericFallback,
}

// resolveCollisions detects name collisions and makes them unique.
// Reference schemas are excluded from collision detection because they don't
// generate types — their names are only used for type resolution lookups.
//
// Resolution proceeds by trying one strategy at a time across all conflicting
// buckets, then re-bucketing. When a strategy makes no progress (no name
// changes across any bucket), the next strategy in the list is tried. The
// strategy list (collisionStrategies) is:
//  1. Context suffix — append a suffix derived from the schema's location.
//  2. Per-schema disambiguation — content type, status code, param index,
//     composition type, with numeric fallback per schema.
//  3. Numeric fallback — unconditionally append i+1 to every member.
func resolveCollisions(schemas []*SchemaDescriptor, candidates map[*SchemaDescriptor]string, converter *NameConverter) {
	// Filter out reference schemas — they don't generate types so their
	// short names can safely shadow non-ref names without causing a collision.
	var nonRefSchemas []*SchemaDescriptor
	for _, s := range schemas {
		if s.Ref == "" {
			nonRefSchemas = append(nonRefSchemas, s)
		}
	}

	maxIterations := 10 // Prevent infinite loops
	strategyIdx := 0

	for range maxIterations {
		// Group non-ref schemas by candidate name
		byName := make(map[string][]*SchemaDescriptor)
		for _, s := range nonRefSchemas {
			name := candidates[s]
			byName[name] = append(byName[name], s)
		}

		// Check if there are any collisions
		hasCollisions := false
		for _, group := range byName {
			if len(group) > 1 {
				hasCollisions = true
				break
			}
		}

		if !hasCollisions {
			return // All names are unique
		}

		if strategyIdx >= len(collisionStrategies) {
			return // Exhausted all strategies
		}

		// Apply the current strategy to all conflicting buckets.
		strategy := collisionStrategies[strategyIdx]
		anyChanged := false
		for _, group := range byName {
			if len(group) <= 1 {
				continue // No collision
			}
			if strategy(group, candidates, converter) {
				anyChanged = true
			}
		}

		// If the strategy made no progress, advance to the next one.
		// Otherwise, re-bucket with the same strategy index (the changes
		// may have created new collisions that the same strategy can fix).
		if !anyChanged {
			strategyIdx++
		}
	}
}

// strategyContextSuffix attempts to disambiguate colliding schemas by
// appending a suffix derived from their path context (e.g. "Request",
// "Response"). If exactly one member is a component schema, it keeps the
// bare name and only the others are suffixed.
func strategyContextSuffix(group []*SchemaDescriptor, candidates map[*SchemaDescriptor]string, _ *NameConverter) bool {
	// Count how many are from components/schemas
	var componentSchemaCount int
	for _, s := range group {
		ctx, _ := parsePathContext(s.Path)
		if ctx == ContextComponentSchema {
			componentSchemaCount++
		}
	}

	// If exactly one is from components/schemas, it is "privileged" and keeps
	// the bare name.
	privileged := componentSchemaCount == 1

	changed := false
	for _, s := range group {
		ctx, _ := parsePathContext(s.Path)

		// Privileged component schema keeps the bare name
		if privileged && ctx == ContextComponentSchema {
			continue
		}

		suffix := schemaContextSuffix(ctx)
		if suffix != "" {
			name := candidates[s]
			if !strings.HasSuffix(name, suffix) {
				candidates[s] = name + suffix
				changed = true
			}
		}
		// If suffix is empty (unknown context), leave unchanged for later
		// strategies to handle.
	}
	return changed
}

// strategyPerSchemaDisambiguate tries per-schema sub-strategies
// (disambiguationStrategies) in order for each member of the group. If no
// sub-strategy matches a given schema, it falls back to a numeric suffix
// (index+1). Returns true if any name was changed.
func strategyPerSchemaDisambiguate(group []*SchemaDescriptor, candidates map[*SchemaDescriptor]string, converter *NameConverter) bool {
	changed := false
	for i, s := range group {
		currentName := candidates[s]
		resolved := false
		for _, sub := range disambiguationStrategies {
			if newName, ok := sub(s, currentName, converter); ok {
				candidates[s] = newName
				changed = true
				resolved = true
				break
			}
		}
		if !resolved {
			// Numeric fallback per schema
			candidates[s] = fmt.Sprintf("%s%d", currentName, i+1)
			changed = true
		}
	}
	return changed
}

// strategyNumericFallback unconditionally appends i+1 to every schema in the
// group. This is the last-resort strategy that always succeeds.
func strategyNumericFallback(group []*SchemaDescriptor, candidates map[*SchemaDescriptor]string, _ *NameConverter) bool {
	for i, s := range group {
		candidates[s] = fmt.Sprintf("%s%d", candidates[s], i+1)
	}
	return true
}

// tryContentTypeSuffix checks for "content/{type}" in the schema path and
// appends a content-type suffix (JSON, XML, Form, Text, Binary, or the
// normalized content type).
func tryContentTypeSuffix(s *SchemaDescriptor, currentName string, converter *NameConverter) (string, bool) {
	for i, part := range s.Path {
		if part == "content" && i+1 < len(s.Path) {
			contentType := s.Path[i+1]
			var suffix string
			switch {
			case strings.Contains(contentType, "json"):
				suffix = "JSON"
			case strings.Contains(contentType, "xml"):
				suffix = "XML"
			case strings.Contains(contentType, "form"):
				suffix = "Form"
			case strings.Contains(contentType, "text"):
				suffix = "Text"
			case strings.Contains(contentType, "binary"):
				suffix = "Binary"
			default:
				suffix = converter.ToTypeName(strings.ReplaceAll(contentType, "/", "_"))
			}
			// Use Contains since the suffix might be embedded before "Response" or "Request"
			if !strings.Contains(currentName, suffix) {
				return currentName + suffix, true
			}
		}
	}
	return currentName, false
}

// tryStatusCodeSuffix checks for "responses/{code}" in the schema path and
// appends the status code.
func tryStatusCodeSuffix(s *SchemaDescriptor, currentName string, _ *NameConverter) (string, bool) {
	for i, part := range s.Path {
		if part == "responses" && i+1 < len(s.Path) {
			code := s.Path[i+1]
			if !strings.Contains(currentName, code) {
				return currentName + code, true
			}
		}
	}
	return currentName, false
}

// tryParamIndexSuffix checks for "parameters/{idx}" in the schema path and
// appends the parameter index.
func tryParamIndexSuffix(s *SchemaDescriptor, currentName string, _ *NameConverter) (string, bool) {
	for i, part := range s.Path {
		if part == "parameters" && i+1 < len(s.Path) {
			idx := s.Path[i+1]
			if !strings.HasSuffix(currentName, idx) {
				return currentName + idx, true
			}
		}
	}
	return currentName, false
}

// tryCompositionTypeSuffix checks for allOf/anyOf/oneOf in the schema path
// (searching from the end) and appends the composition type and index.
func tryCompositionTypeSuffix(s *SchemaDescriptor, currentName string, converter *NameConverter) (string, bool) {
	for i := len(s.Path) - 1; i >= 0; i-- {
		part := s.Path[i]
		switch part {
		case "allOf", "anyOf", "oneOf":
			suffix := converter.ToTypeName(part)
			if i+1 < len(s.Path) {
				suffix += s.Path[i+1] // Add index
			}
			if !strings.Contains(currentName, suffix) {
				return currentName + suffix, true
			}
		}
	}
	return currentName, false
}
