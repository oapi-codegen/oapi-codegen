package codegen

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ResolvedName holds the final Go type name assigned to a gathered schema.
type ResolvedName struct {
	Schema    *GatheredSchema
	GoName    string // The resolved Go type name
	Candidate string // The initial candidate name before collision resolution
}

// ResolveNames takes the gathered schemas and assigns unique Go type names to each.
// It returns a map from the schema's path string to the resolved Go type name.
func ResolveNames(schemas []*GatheredSchema) map[string]string {
	// Step 1: Generate candidate names for all schemas
	candidates := make([]*ResolvedName, len(schemas))
	for i, s := range schemas {
		candidate := generateCandidateName(s)
		candidates[i] = &ResolvedName{
			Schema:    s,
			GoName:    candidate,
			Candidate: candidate,
		}
	}

	// Step 2: Resolve collisions iteratively
	resolveCollisions(candidates)

	// Step 3: Build the result map
	result := make(map[string]string, len(candidates))
	for _, c := range candidates {
		result[c.Schema.Path.String()] = c.GoName
	}
	return result
}

// generateCandidateName produces an initial Go type name candidate based on
// the schema's location and context in the OpenAPI document.
func generateCandidateName(s *GatheredSchema) string {
	switch s.Context {
	case ContextComponentSchema:
		return SchemaNameToTypeName(s.ComponentName)

	case ContextComponentParameter:
		return SchemaNameToTypeName(s.ComponentName)

	case ContextComponentResponse:
		return SchemaNameToTypeName(s.ComponentName)

	case ContextComponentRequestBody:
		return SchemaNameToTypeName(s.ComponentName)

	case ContextComponentHeader:
		return SchemaNameToTypeName(s.ComponentName)

	case ContextClientResponseWrapper:
		// Client response wrappers use: OperationId + responseTypeSuffix
		return fmt.Sprintf("%s%s", UppercaseFirstCharacter(s.OperationID), responseTypeSuffix)

	case ContextOperationParameter:
		if s.OperationID != "" {
			return SchemaNameToTypeName(s.OperationID) + "Parameter"
		}
		return SchemaNameToTypeName(s.ComponentName) + "Parameter"

	case ContextOperationRequestBody:
		if s.OperationID != "" {
			ct := contentTypeSuffix(s.ContentType)
			return SchemaNameToTypeName(s.OperationID) + ct + "Request"
		}
		return SchemaNameToTypeName(s.ComponentName) + "Request"

	case ContextOperationResponse:
		if s.OperationID != "" {
			ct := contentTypeSuffix(s.ContentType)
			return SchemaNameToTypeName(s.OperationID) + s.StatusCode + ct + "Response"
		}
		return SchemaNameToTypeName(s.ComponentName) + "Response"

	default:
		return SchemaNameToTypeName(s.ComponentName)
	}
}

// resolveCollisions detects and resolves naming collisions among the resolved names.
// It applies strategies in global phases of increasing aggressiveness:
// 1. Context suffix (Schema, Parameter, Response, etc.)
// 2. Per-schema disambiguation (content type, status code, etc.)
// 3. Numeric fallback
//
// Each strategy is applied to ALL colliding groups, then collisions are
// re-checked globally before moving to the next strategy. This prevents
// oscillation between strategies (e.g., context suffix and content type
// suffix repeatedly appending to the same names without resolution).
func resolveCollisions(names []*ResolvedName) {
	strategies := []func([]*ResolvedName) bool{
		strategyContextSuffix,
		strategyPerSchemaDisambiguate,
		strategyNumericFallback,
	}

	const maxIterations = 20

	for _, strategy := range strategies {
		for iter := 0; iter < maxIterations; iter++ {
			groups := groupByName(names)
			anyCollision := false
			anyProgress := false
			for _, group := range groups {
				if len(group) <= 1 {
					continue
				}
				anyCollision = true
				if strategy(group) {
					anyProgress = true
				}
			}
			if !anyCollision {
				return
			}
			if !anyProgress {
				break // This strategy can't help; try the next one
			}
		}
	}
}

// groupByName groups ResolvedNames by their current GoName.
func groupByName(names []*ResolvedName) map[string][]*ResolvedName {
	groups := make(map[string][]*ResolvedName)
	for _, n := range names {
		groups[n.GoName] = append(groups[n.GoName], n)
	}
	return groups
}

// strategyContextSuffix attempts to resolve collisions by appending a suffix
// derived from the schema's context (Schema, Parameter, Response, etc.).
// Component schemas are "privileged" — if exactly one member is a component
// schema, it keeps the bare name and only the others get suffixed.
// Returns true if any name was modified, false if no progress was made.
func strategyContextSuffix(group []*ResolvedName) bool {
	// Count how many are component schemas (privileged)
	var componentSchemaCount int
	for _, n := range group {
		if n.Schema.IsComponentSchema() {
			componentSchemaCount++
		}
	}

	progress := false
	for _, n := range group {
		suffix := n.Schema.Context.Suffix()
		if suffix == "" {
			continue
		}

		// If exactly one is a component schema, it keeps the bare name
		if componentSchemaCount == 1 && n.Schema.IsComponentSchema() {
			continue
		}

		// Don't add suffix if name already ends with it
		if strings.HasSuffix(n.GoName, suffix) {
			continue
		}

		n.GoName = n.GoName + suffix
		progress = true
	}
	return progress
}

// strategyPerSchemaDisambiguate tries several per-schema disambiguation strategies.
// Returns true if any name was modified, false if no progress was made.
func strategyPerSchemaDisambiguate(group []*ResolvedName) bool {
	progress := tryContentTypeSuffix(group)
	if !progress && tryStatusCodeSuffix(group) {
		progress = true
	}
	if !progress && tryParamIndexSuffix(group) {
		progress = true
	}
	return progress
}

// tryContentTypeSuffix appends a content type discriminator when schemas
// differ by media type (e.g., JSON vs XML).
// Returns true if any name was modified, false if no progress was made.
func tryContentTypeSuffix(group []*ResolvedName) bool {
	// Check if any members have different content types
	contentTypes := make(map[string]bool)
	for _, n := range group {
		if n.Schema.ContentType != "" {
			contentTypes[n.Schema.ContentType] = true
		}
	}
	if len(contentTypes) <= 1 {
		return false
	}

	progress := false
	for _, n := range group {
		if n.Schema.ContentType == "" {
			continue
		}
		suffix := contentTypeSuffix(n.Schema.ContentType)
		if suffix != "" && !strings.HasSuffix(n.GoName, suffix) {
			n.GoName = n.GoName + suffix
			progress = true
		}
	}
	return progress
}

// tryStatusCodeSuffix appends the HTTP status code when schemas differ by status.
// Returns true if any name was modified, false if no progress was made.
func tryStatusCodeSuffix(group []*ResolvedName) bool {
	statusCodes := make(map[string]bool)
	for _, n := range group {
		if n.Schema.StatusCode != "" {
			statusCodes[n.Schema.StatusCode] = true
		}
	}
	if len(statusCodes) <= 1 {
		return false
	}

	progress := false
	for _, n := range group {
		if n.Schema.StatusCode != "" && !strings.HasSuffix(n.GoName, n.Schema.StatusCode) {
			n.GoName = n.GoName + n.Schema.StatusCode
			progress = true
		}
	}
	return progress
}

// tryParamIndexSuffix appends a parameter index when schemas differ by position.
// Returns true if any name was modified, false if no progress was made.
func tryParamIndexSuffix(group []*ResolvedName) bool {
	hasMultipleParams := false
	for i := 0; i < len(group); i++ {
		for j := i + 1; j < len(group); j++ {
			if group[i].Schema.ParamIndex != group[j].Schema.ParamIndex {
				hasMultipleParams = true
				break
			}
		}
		if hasMultipleParams {
			break
		}
	}
	if !hasMultipleParams {
		return false
	}

	progress := false
	for _, n := range group {
		suffix := strconv.Itoa(n.Schema.ParamIndex)
		if !strings.HasSuffix(n.GoName, suffix) {
			n.GoName = n.GoName + suffix
			progress = true
		}
	}
	return progress
}

// strategyNumericFallback is the last resort: append increasing numbers.
// Returns true if any name was modified (always true when group has 2+ members).
func strategyNumericFallback(group []*ResolvedName) bool {
	// Sort for determinism: component schemas first, then by path
	sort.Slice(group, func(i, j int) bool {
		if group[i].Schema.IsComponentSchema() != group[j].Schema.IsComponentSchema() {
			return group[i].Schema.IsComponentSchema()
		}
		return group[i].Schema.Path.String() < group[j].Schema.Path.String()
	})

	// First entry keeps name, rest get numeric suffix
	for i := 1; i < len(group); i++ {
		group[i].GoName = group[i].GoName + strconv.Itoa(i+1)
	}
	return len(group) > 1
}

// contentTypeSuffix returns a short suffix for a media type.
func contentTypeSuffix(ct string) string {
	if ct == "" {
		return ""
	}
	ct = strings.ToLower(ct)
	switch {
	case strings.Contains(ct, "json"):
		return "JSON"
	case strings.Contains(ct, "xml"):
		return "XML"
	case strings.Contains(ct, "form"):
		return "Form"
	case strings.Contains(ct, "text"):
		return "Text"
	case strings.Contains(ct, "octet"):
		return "Binary"
	case strings.Contains(ct, "yaml"):
		return "YAML"
	default:
		return mediaTypeToCamelCase(ct)
	}
}
