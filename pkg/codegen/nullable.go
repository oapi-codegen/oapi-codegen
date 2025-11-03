package codegen

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// IsSchemaNullable determines whether a schema permits JSON null.
//
//   - For OpenAPI 3.1 (JSON Schema 2020-12), a schema is nullable if, considering
//     base constraints (type/enum/not) and composition (oneOf/anyOf/allOf), the
//     overall schema would accept the JSON null value. Empty schemas `{}` are not
//     considered nullable here by design to avoid over-detection in generators.
//   - For OpenAPI 3.0, falls back to `schema.Value.Nullable`.
func IsSchemaNullable(specVersion string, sref *openapi3.SchemaRef) bool {
	if sref == nil || sref.Value == nil {
		return false
	}

	if strings.HasPrefix(specVersion, "3.1") {
		return permitsNull31(sref)
	}
	return sref.Value.Nullable
}

// permitsNull31 returns true if the schema (draft 2020-12 semantics) permits JSON null,
// accounting for base constraints (type/enum/not) and composed constraints (oneOf/anyOf/allOf).
// NOTE: Empty schemas are intentionally treated as non-nullable for generator purposes.
func permitsNull31(schemaRef *openapi3.SchemaRef) bool {
	if schemaRef == nil || schemaRef.Value == nil {
		return false
	}
	schema := schemaRef.Value

	// Base constraints that immediately rule out null
	if baseDisallowsNull(schema) {
		return false
	}
	// "not" forbids null if its subschema would accept null
	if notForbidsNull(schema.Not) {
		return false
	}

	// Composition constraints - AllOf
	if len(schema.AllOf) > 0 {
		for _, sub := range schema.AllOf {
			if !permitsNull31(sub) {
				return false
			}
		}
		return true
	}
	// Composition constraints - OneOf
	if len(schema.OneOf) > 0 {
		matches := 0
		for _, sub := range schema.OneOf {
			if permitsNull31(sub) {
				matches++
			}
		}
		return matches == 1
	}
	// Composition constraints - AnyOf
	if len(schema.AnyOf) > 0 {
		for _, sub := range schema.AnyOf {
			if permitsNull31(sub) {
				return true
			}
		}
		return false
	}

	// No composition: decide from type/enum
	if typeIncludesNull(schema.Type) {
		return true
	}
	if enumIncludesNull(schema.Enum) {
		return true
	}

	// Treat empty/unconstrained schemas as non-nullable for codegen
	return false
}

func baseDisallowsNull(s *openapi3.Schema) bool {
	// If a base type is present and does NOT include null, null cannot be valid overall
	if s.Type != nil && !typeIncludesNull(s.Type) {
		return true
	}
	// If enum exists and does NOT contain null, then null is disallowed
	if len(s.Enum) > 0 && !enumIncludesNull(s.Enum) {
		return true
	}
	return false
}

func typeIncludesNull(t *openapi3.Types) bool {
	if t == nil {
		return false
	}
	if t.Is(openapi3.TypeNull) {
		return true
	}
	for _, typ := range t.Slice() {
		if typ == openapi3.TypeNull {
			return true
		}
	}
	return false
}

func enumIncludesNull(values []any) bool {
	for _, v := range values {
		if v == nil {
			return true
		}
	}
	return false
}

func notForbidsNull(notRef *openapi3.SchemaRef) bool {
	if notRef == nil || notRef.Value == nil {
		return false
	}
	// If the negated schema would accept null, then the parent forbids null
	return permitsNull31(notRef)
}
