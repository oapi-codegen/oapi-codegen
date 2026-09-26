package codegen

// This file holds the anyOf/oneOf code that every schema-merging-behavior
// version shares: the dispatch on the version in effect, the collapse of a
// nullable union into its one branch, and the helpers that read a union's
// branches. The versions' own code is in union_v2.go and union_v3.go. A
// change here changes what every version generates, so the versions move in
// step.

import (
	"maps"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// effectiveBranches counts the branches of a oneOf or anyOf other than a bare
// `{type: "null"}`. In OpenAPI 3.1 such a branch is a nullability marker, not
// a real union variant: there is no Go type that corresponds to "only the JSON
// value null". The parent schema's nullability is captured by
// schemaIsNullable, which inspects anyOf/oneOf for the same idiom and wraps the
// result in a pointer at the call site. It returns the count, whether there
// was a null branch, and the first effective branch.
func effectiveBranches(elements openapi3.SchemaRefs) (count int, hadNull bool, sole *openapi3.SchemaRef) {
	for _, e := range elements {
		if e != nil && isNullTypeSchema(e.Value) {
			hadNull = true
			continue
		}
		count++
		if sole == nil {
			sole = e
		}
	}
	return count, hadNull, sole
}

// collapseNullableUnion generates a oneOf or anyOf whose only effective branch
// is sole (see effectiveBranches) as that branch: the schema is semantically
// equivalent to the single branch, made nullable by the null branch. This
// produces the same Go shape the type-array idiom would: `anyOf: [{type:
// string}, {type: "null"}]` must generate the same `*string` field as `type:
// ["string", "null"]`. Without this, the single remaining branch would be
// wrapped in a one-variant union type, exposing a needless `FromX`/`AsX`
// accessor API.
//
// outSchema inherits the branch's underlying representation. The caller will
// apply nullability (schemaIsNullable returns true because the original
// anyOf/oneOf contained a null branch).
func collapseNullableUnion(ctx genContext, outSchema *Schema, sole *openapi3.SchemaRef, path []string) error {
	elementSchema, err := generateGoSchema(ctx.at(path), sole, path)
	if err != nil {
		return err
	}
	outSchema.GoType = elementSchema.GoType
	outSchema.RefType = elementSchema.RefType
	outSchema.DefineViaAlias = elementSchema.DefineViaAlias
	outSchema.Properties = elementSchema.Properties
	outSchema.HasAdditionalProperties = elementSchema.HasAdditionalProperties
	outSchema.AdditionalPropertiesType = elementSchema.AdditionalPropertiesType
	outSchema.ArrayType = elementSchema.ArrayType
	outSchema.SkipOptionalPointer = elementSchema.SkipOptionalPointer
	outSchema.AdditionalTypes = append(outSchema.AdditionalTypes, elementSchema.AdditionalTypes...)
	return nil
}

// generateUnions generates the anyOf and oneOf of an object schema into
// outSchema's union members, the way the schema-merging-behavior in effect
// does. v1 and v2 share v2's code.
func generateUnions(ctx genContext, outSchema *Schema, schema *openapi3.Schema, path []string) error {
	if ctx.run.merging == SchemaMergingV3 {
		return generateUnionsV3(ctx, outSchema, schema, path)
	}
	return generateUnionsV2(ctx, outSchema, schema, path)
}

// unionTextKinds returns the sorted JSON types of a union's branches when
// every branch (other than a 3.1 null branch) is a single scalar type, or is
// itself a union of scalar types, such as a `$ref` to one; and nil otherwise.
func unionTextKinds(branches openapi3.SchemaRefs) []string {
	return unionTextKindsAt(branches, 0)
}

func unionTextKindsAt(branches openapi3.SchemaRefs, depth int) []string {
	if depth > 8 {
		// A union that contains itself; it can't be bound from text.
		return nil
	}
	var kinds []string
	for _, b := range branches {
		if b == nil || b.Value == nil {
			return nil
		}
		v := b.Value
		if isNullTypeSchema(v) {
			continue
		}
		if len(v.Properties) > 0 || v.Items != nil || len(v.AllOf) > 0 || v.AdditionalProperties.Schema != nil {
			return nil
		}
		if len(v.AnyOf) > 0 || len(v.OneOf) > 0 {
			if v.Type.Slice() != nil {
				return nil
			}
			nested := unionTextKindsAt(slices.Concat(v.AnyOf, v.OneOf), depth+1)
			if nested == nil {
				return nil
			}
			for _, kind := range nested {
				kinds = appendUnique(kinds, kind)
			}
			continue
		}
		types := nonNullTypes(v.Type)
		if len(types) != 1 {
			return nil
		}
		switch types[0] {
		case openapi3.TypeBoolean, openapi3.TypeInteger, openapi3.TypeNumber, openapi3.TypeString:
			kinds = appendUnique(kinds, types[0])
		default:
			return nil
		}
	}
	slices.Sort(kinds)
	return kinds
}

// AdoptedProperties returns the properties of a union that its From* and
// Merge* helpers fill from the variant: those that MarshalJSON always writes,
// so that a zero value would overwrite the variant's. The rest are written only
// when set, and otherwise the union data's value is written as it is.
func (s Schema) AdoptedProperties() []Property {
	var out []Property
	for _, p := range s.Properties {
		if !p.RequiresNilCheck() {
			out = append(out, p)
		}
	}
	return out
}

// UnionTextNonStringKinds names the JSON types other than string that a
// scalar union accepts, for its UnmarshalText doc comment: "boolean or
// integer"; see UnionTextKinds.
func (s Schema) UnionTextNonStringKinds() string {
	var kinds []string
	for _, kind := range s.UnionTextKinds {
		if kind != openapi3.TypeString {
			kinds = append(kinds, kind)
		}
	}
	return strings.Join(kinds, " or ")
}

// UnionTextAccepts reports whether a scalar union has a branch of the given
// JSON type; see UnionTextKinds.
func (s Schema) UnionTextAccepts(kind string) bool {
	return slices.Contains(s.UnionTextKinds, kind)
}

// discriminatorValueType returns the JSON type of a union's discriminator
// property when it is boolean, integer or number, and "" (a string) otherwise.
// The union's own declaration of the property wins; failing that, every
// element that declares it must agree.
func discriminatorValueType(outSchema *Schema, elements openapi3.SchemaRefs, property string) string {
	for _, p := range outSchema.Properties {
		if p.JsonFieldName == property {
			return scalarType(p.Schema.OAPISchema)
		}
	}
	found := ""
	for _, e := range elements {
		if e == nil || e.Value == nil || isNullTypeSchema(e.Value) {
			continue
		}
		prop := findProperty(e.Value, property, 0)
		if prop == nil {
			continue
		}
		typ := scalarType(prop)
		if found != "" && typ != found {
			return ""
		}
		found = typ
	}
	return found
}

// findProperty looks up a property on a schema, including properties it
// gets from allOf members, as a variant that extends a base schema does.
func findProperty(s *openapi3.Schema, name string, depth int) *openapi3.Schema {
	if s == nil || depth > 8 {
		return nil
	}
	if p, ok := s.Properties[name]; ok && p != nil {
		return p.Value
	}
	for _, m := range s.AllOf {
		if m == nil {
			continue
		}
		if p := findProperty(m.Value, name, depth+1); p != nil {
			return p
		}
	}
	return nil
}

// propertyNames returns the property names a schema declares, including
// those it gets from allOf members.
func propertyNames(s *openapi3.Schema, depth int) []string {
	if s == nil || depth > 8 {
		return nil
	}
	names := slices.Collect(maps.Keys(s.Properties))
	for _, m := range s.AllOf {
		if m != nil {
			names = append(names, propertyNames(m.Value, depth+1)...)
		}
	}
	return names
}

// scalarType returns "boolean", "integer" or "number" for a schema of that
// single (non-null) type, and "" for anything else.
func scalarType(s *openapi3.Schema) string {
	if s == nil {
		return ""
	}
	types := nonNullTypes(s.Type)
	if len(types) != 1 {
		return ""
	}
	switch types[0] {
	case openapi3.TypeBoolean, openapi3.TypeInteger, openapi3.TypeNumber:
		return types[0]
	}
	return ""
}
