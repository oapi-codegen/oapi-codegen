package codegen

// This file holds the allOf half of schema-merging-behavior v3, the version
// under development. Its anyOf/oneOf half is in union_v3.go. It started as a
// copy of v2's (merge_schemas_v2.go).

import (
	"cmp"
	"fmt"
	"iter"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

func mergeSchemasV3(ctx genContext, allOf []*openapi3.SchemaRef, path []string) (Schema, error) {
	n := len(allOf)

	// A child that is only an allOf of the parent that lists it is not the
	// parent's union: it is merged, which leaves the list out (see
	// listsFlattened).
	if n == 1 && !listsComposition(ctx, refSchemaFor(allOf[0])) {
		return generateGoSchema(ctx, allOf[0], path)
	}

	// A member whose schema the merge can't read can only be annotated by the
	// others, and then the composition is that member's type.
	opaque, err := opaqueMember(allOf)
	if err != nil {
		return Schema{}, err
	}
	if opaque != nil {
		return generateAnnotated(ctx, opaque, allOf, path)
	}

	// Likewise a $ref that the other members only annotate, the way OpenAPI
	// 3.0 puts anything next to a $ref: `allOf: [$ref A, {nullable: true}]`
	// is A, not a copy of A.
	if member := annotatedRefMember(ctx, allOf); member != nil {
		return generateAnnotated(ctx, member, allOf, path)
	}

	// The rest is a new type. Extensions that name a member's type, such as
	// a $ref'd schema's x-go-type-name, don't transfer to it, unless a member
	// that only annotates put them there for the composition, like
	// `allOf: [{properties: ...}, {x-go-type-name: Named}]`.
	decoratorIdiom := false
	for _, m := range allOf {
		if annotatesOnly(m) && len(m.Value.Extensions) > 0 {
			decoratorIdiom = true
			break
		}
	}

	merged := newAllOfMerge(ctx)
	merged.composition = &openapi3.Schema{AllOf: allOf}
	// The $refs of the members merged so far, so that a nested allOf that
	// refers back to one is not flattened into itself.
	seenTopLevel := make(map[string]bool)
	for _, member := range allOf {
		merged.markFlattening(member)
	}
	for i, member := range allOf {
		value, err := memberValue(member)
		if err != nil {
			return Schema{}, err
		}
		seen := maps.Clone(seenTopLevel)
		if member.Ref != "" {
			seen[member.Ref] = true
			seenTopLevel[member.Ref] = true
		}
		if err := merged.add(member, value, allOfMemberLabel(ctx, "", member, i), seen); err != nil {
			return Schema{}, err
		}
	}
	schema, err := merged.result()
	if err != nil {
		return Schema{}, err
	}

	if !decoratorIdiom {
		// Drop only the type-identity directives. Other extensions
		// (user-defined x-* metadata, etc.) are preserved — we only
		// have concrete evidence that the identity-bound ones cause
		// incorrect aliasing across composition.
		ext := maps.Clone(schema.Extensions)
		delete(ext, extGoTypeName)
		delete(ext, extPropGoImport)
		schema.Extensions = ext
	}

	if components := merged.components; len(components) > 1 {
		ctx.unionComponents[&schema] = components
	}
	return generateGoSchema(ctx, openapi3.NewSchemaRef("", &schema), path)
}

// memberValue returns a copy of an allOf member's schema, with extensions
// placed next to its $ref folded in (ref-side wins over value-side). This is
// what allows allOf members to carry per-use sibling directives without
// mutating the referenced schema.
//
// Reading a schema the merge can't read is an error (see isOpaqueSchema). Every
// read of a member goes through here, so however deep a merge flattens nested
// allOfs and array items, it never copies such a schema's body.
func memberValue(ref *openapi3.SchemaRef) (openapi3.Schema, error) {
	if isOpaqueSchema(ref) {
		return openapi3.Schema{}, opaqueMergeError(ref, ref, "other schemas")
	}
	schema := *ref.Value
	schema.Extensions = combinedSchemaExtensions(ref)
	return schema, nil
}

// isOpaqueSchema reports whether the merge can't read a schema:
//
//   - a schema in another document, which another generator run turns into a
//     Go type, with its own configuration; its body is not ours to copy, and
//     the refs inside it point into a document this run doesn't generate
//     (#2288). That includes a component of this document that is only a $ref
//     to one (see isRefInExternalDocument);
//   - a schema that x-go-type replaces with a Go type whose fields we can't
//     see.
//
// Such a schema can be annotated, but not merged with other schemas.
func isOpaqueSchema(ref *openapi3.SchemaRef) bool {
	if ref == nil {
		return false
	}
	if isRefInExternalDocument(ref.Ref) {
		return true
	}
	_, ok := combinedSchemaExtensions(ref)[extPropGoType]
	return ok
}

// isRefInExternalDocument reports whether ref points into another document.
// A local ref does when it names a component schema that is only a $ref into
// another document, directly or through other such components: kin-openapi
// hands us the other document's schema for it, with nothing to say where it
// came from.
//
// A $ref to a whole document ("./user.yaml", no fragment) doesn't point into
// one. oapi-codegen has no Go type for such a schema, and inlines it wherever
// it is used, so it is part of this run's output like any inline schema.
func isRefInExternalDocument(ref string) bool {
	seen := make(map[string]bool)
	for ref != "" && !seen[ref] {
		if ref[0] != '#' {
			return IsGoTypeReference(ref)
		}
		seen[ref] = true
		name, ok := strings.CutPrefix(ref, "#/components/schemas/")
		if !ok || strings.Contains(name, "/") || globalState.spec == nil || globalState.spec.Components == nil {
			return false
		}
		component := globalState.spec.Components.Schemas[name]
		if component == nil {
			return false
		}
		ref = component.Ref
	}
	return false
}

// opaqueSchemaFor returns the opaque schema (see isOpaqueSchema) an allOf member
// stands for, or nil (see standsFor).
func opaqueSchemaFor(ref *openapi3.SchemaRef) *openapi3.SchemaRef {
	return standsFor(ref, nil, isOpaqueSchema)
}

// refSchemaFor returns the $ref an allOf member stands for, or nil (see
// standsFor).
func refSchemaFor(ref *openapi3.SchemaRef) *openapi3.SchemaRef {
	return standsFor(ref, nil, func(r *openapi3.SchemaRef) bool { return r.Ref != "" })
}

// standsFor returns the schema an allOf member stands for, when is reports
// true for it, or nil. That is the member itself, or else, when the member is
// a composition, the member of that composition that its other members only
// annotate (see annotates): such a member generates as that schema's type.
func standsFor(ref *openapi3.SchemaRef, seen map[*openapi3.Schema]bool, is func(*openapi3.SchemaRef) bool) *openapi3.SchemaRef {
	if ref == nil {
		return nil
	}
	if is(ref) {
		return ref
	}
	s := ref.Value
	if s == nil || len(s.AllOf) == 0 || seen[s] {
		return nil
	}
	if seen == nil {
		seen = make(map[*openapi3.Schema]bool)
	}
	seen[s] = true
	members := compositionMembers(s)
	for _, m := range members {
		target := standsFor(m, seen, is)
		if target == nil {
			continue
		}
		for _, other := range members {
			if other != m && !annotates(other, target) {
				return nil
			}
		}
		return target
	}
	return nil
}

// compositionMembers returns what a schema with allOf merges: the allOf's
// members, and the schema's own keywords as one more member when it has any
// that shape a Go type. In JSON Schema a keyword next to allOf constrains the
// value like any member does: `{type: string, allOf: [A]}` is
// `{allOf: [A, {type: string}]}`.
func compositionMembers(s *openapi3.Schema) []*openapi3.SchemaRef {
	if !hasStructuralSiblingsV3(s) {
		return s.AllOf
	}
	own := *s
	own.AllOf = nil
	members := make([]*openapi3.SchemaRef, 0, len(s.AllOf)+1)
	members = append(members, s.AllOf...)
	return append(members, &openapi3.SchemaRef{Value: &own})
}

// annotatedRefMember returns the member of allOf that stands for a $ref (see
// refSchemaFor) when every other member only annotates that $ref (see
// annotates), or nil.
//
// At the top of a component schema the composition would then be defined as
// an alias of the $ref. When the $ref leads back to the component, through
// compositions that are aliases in turn, that alias would be an alias of
// itself, which Go rejects, so such a composition is merged instead.
func annotatedRefMember(ctx genContext, allOf []*openapi3.SchemaRef) *openapi3.SchemaRef {
	for _, m := range allOf {
		target := refSchemaFor(m)
		if target == nil {
			continue
		}
		for _, other := range allOf {
			if other != m && !annotates(other, target) {
				return nil
			}
		}
		if (ctx.rootPosition && aliasesBack(ctx, target)) || listsComposition(ctx, target) {
			return nil
		}
		return m
	}
	return nil
}

// listsComposition reports whether target's oneOf or anyOf lists the
// composition being merged: target is the parent of a child that is an allOf
// of it (see listsFlattened), so the child isn't target's type.
func listsComposition(ctx genContext, target *openapi3.SchemaRef) bool {
	if target == nil || target.Value == nil || ctx.composing == nil {
		return false
	}
	for _, b := range slices.Concat(target.Value.OneOf, target.Value.AnyOf) {
		if b != nil && b.Value == ctx.composing {
			return true
		}
	}
	return false
}

// annotates reports whether an allOf member only annotates target, the member
// the composition stands for: it only annotates (see annotatesOnly), it
// restates target's type (see restatesType), or it is the same $ref again.
func annotates(member, target *openapi3.SchemaRef) bool {
	// A oneOf or anyOf in the member may only add constraints to the target,
	// such as `oneOf: [{type: object, required: [email]}, ...]` next to an
	// object; the member's own lack of a type doesn't make them types.
	if member != nil && member.Ref == "" && member.Value != nil && target.Value != nil {
		if v := withoutConstraintOnlyUnions(member.Value, target.Value); v != member.Value {
			member = &openapi3.SchemaRef{Value: v}
		}
	}
	if annotatesOnly(member) || restatesType(member, target) {
		return true
	}
	return member != nil && member.Ref != "" && member.Ref == target.Ref
}

// restatesType reports whether member only restates target's type and format,
// as in `allOf: [$ref A, {type: object, description: ...}]`: it is inline,
// declares nothing else that shapes a Go type (nor x-go-type or enum names),
// and declares the same types and format as target. A target with properties
// or additionalProperties but no type is an object.
//
// For a schema in another document, only its type and format are read, to
// check the restatement; its body never is. When it doesn't say what it is at
// the top, as an allOf of its own doesn't, a restated type is taken at its
// word.
func restatesType(member, target *openapi3.SchemaRef) bool {
	if member == nil || member.Ref != "" || member.Value == nil {
		return false
	}
	for _, ext := range []string{extPropGoType, extEnumVarNames, extEnumNames} {
		if _, ok := member.Value.Extensions[ext]; ok {
			return false
		}
	}
	restated := member.Value
	for _, keyword := range typeKeywords(*restated) {
		if keyword != "type" && keyword != "format" {
			return false
		}
	}
	external := isRefInExternalDocument(target.Ref)
	declared := target.Value
	if declared == nil {
		return external
	}
	if restated.Format != "" && restated.Format != declared.Format {
		return false
	}
	types := declaredTypes(declared)
	if len(types) == 0 && external {
		return true
	}
	return sameTypes(types, nonNullTypes(restated.Type))
}

// declaredTypes returns the JSON types a schema says its values have: its
// type, or object when it has properties or additionalProperties, narrowed by
// the types its allOf members declare, the way the merge intersects them. It
// returns nil when the schema doesn't say, or says no value is allowed.
func declaredTypes(s *openapi3.Schema) []string {
	seen := make(map[*openapi3.Schema]bool)
	var declared func(s *openapi3.Schema) []string
	declared = func(s *openapi3.Schema) []string {
		if s == nil || seen[s] {
			return nil
		}
		seen[s] = true
		types := nonNullTypes(s.Type)
		if len(types) == 0 && (len(s.Properties) > 0 || s.AdditionalProperties.Schema != nil || s.AdditionalProperties.Has != nil) {
			types = []string{openapi3.TypeObject}
		}
		for _, m := range s.AllOf {
			if m == nil {
				continue
			}
			switch member := declared(m.Value); {
			case len(member) == 0:
			case len(types) == 0:
				types = member
			default:
				types = intersectTypes(types, member)
			}
		}
		return types
	}
	return declared(s)
}

// sameTypes reports whether two type lists allow the same values: the same
// types in any order, with [integer, number] the same as [number].
func sameTypes(a, b []string) bool {
	a, b = withoutIntegerUnderNumber(a), withoutIntegerUnderNumber(b)
	if len(a) != len(b) {
		return false
	}
	for _, t := range a {
		if !slices.Contains(b, t) {
			return false
		}
	}
	return true
}

// aliasesBack reports whether target, a $ref an alias is about to be defined
// as, is a composition being generated, or is an alias of one: a composition
// that stands for a $ref (see refSchemaFor). It follows up to 100 such
// aliases, far more than any spec chains, and treats a longer chain as a
// cycle, so the composition is merged rather than looping.
func aliasesBack(ctx genContext, target *openapi3.SchemaRef) bool {
	for range 100 {
		s := target.Value
		if s == nil {
			return false
		}
		if _, ok := ctx.inProgress[s]; ok {
			return true
		}
		if len(s.AllOf) == 0 {
			return false
		}
		next := refSchemaFor(&openapi3.SchemaRef{Value: s})
		if next == nil {
			return false
		}
		target = next
	}
	return true
}

// allOfTree yields ref, then the members of its allOf, all the way down, in
// the order the merge flattens them: each before the members of its own allOf.
// A schema reached again is yielded again but not descended into again, so a
// cycle ends. A ref with no schema is skipped. Break to stop early.
func allOfTree(ref *openapi3.SchemaRef) iter.Seq[*openapi3.SchemaRef] {
	return func(yield func(*openapi3.SchemaRef) bool) {
		seen := make(map[*openapi3.Schema]bool)
		var walk func(r *openapi3.SchemaRef) bool
		walk = func(r *openapi3.SchemaRef) bool {
			if r == nil || r.Value == nil {
				return true
			}
			if !yield(r) {
				return false
			}
			if seen[r.Value] {
				return true
			}
			seen[r.Value] = true
			for _, m := range r.Value.AllOf {
				if !walk(m) {
					return false
				}
			}
			return true
		}
		walk(ref)
	}
}

// opaqueSchemaWithin returns an opaque schema (see isOpaqueSchema) that
// merging ref would have to read, or nil: ref itself, or an opaque member of
// an allOf that the merge flattens out of ref, however deep.
func opaqueSchemaWithin(ref *openapi3.SchemaRef) *openapi3.SchemaRef {
	for r := range allOfTree(ref) {
		if isOpaqueSchema(r) {
			return r
		}
	}
	return nil
}

// opaqueMember looks for an allOf member that stands for an opaque
// schema (see opaqueSchemaFor). It returns nil when there is none. When every
// other member only annotates it, it returns that member, whose type is the
// composition's type. Any other combination would need fields the merge can't
// read, so it is an error, as is a member whose own allOf includes an opaque
// schema that merging it would flatten.
func opaqueMember(allOf []*openapi3.SchemaRef) (*openapi3.SchemaRef, error) {
	for _, m := range allOf {
		target := opaqueSchemaFor(m)
		if target == nil {
			continue
		}
		for _, other := range allOf {
			if other != m && !annotates(other, target) {
				return nil, opaqueMergeError(m, target, describeAllOfMember(other))
			}
		}
		return m, nil
	}
	for i, m := range allOf {
		if target := opaqueSchemaWithin(m); target != nil {
			other := allOf[0]
			if i == 0 {
				other = allOf[1]
			}
			return nil, opaqueMergeError(m, target, describeAllOfMember(other))
		}
	}
	return nil, nil
}

// generateAnnotated generates an allOf of a member, opaque or a $ref, and
// members that only annotate it: the member's type. Nullability needs nothing
// here, since schemaIsNullable finds a nullable member wherever the type is
// used. An annotating member's description describes the result.
func generateAnnotated(ctx genContext, member *openapi3.SchemaRef, allOf []*openapi3.SchemaRef, path []string) (Schema, error) {
	out, err := generateGoSchema(ctx, member, path)
	if err != nil {
		return Schema{}, err
	}
	var typeName string
	for _, m := range allOf {
		if m == member || m.Value == nil {
			continue
		}
		if m.Value.Description != "" {
			out.Description = m.Value.Description
		}
		// The idiom of issue #1957: a member that sets
		// x-go-type-skip-optional-pointer for the composed type.
		if ext, ok := m.Value.Extensions[extPropGoTypeSkipOptionalPointer]; ok {
			out.SkipOptionalPointer, err = extParsePropGoTypeSkipOptionalPointer(ext)
			if err != nil {
				return Schema{}, fmt.Errorf("invalid value for %q: %w", extPropGoTypeSkipOptionalPointer, err)
			}
		}
		if ext, ok := m.Value.Extensions[extGoTypeName]; ok {
			typeName, err = extTypeName(ext)
			if err != nil {
				return Schema{}, fmt.Errorf("invalid value for %q: %w", extGoTypeName, err)
			}
		}
	}
	if typeName != "" {
		// As x-go-type-name does anywhere else: define the named type, and
		// use it here. The named type is an alias of the member's type, whose
		// schema generatesMarshalJSON has to reach.
		if out.DefineViaAlias && out.OAPISchema != nil {
			out.aliasOf = out.OAPISchema
		}
		out = Schema{
			Description:         out.Description,
			GoType:              typeName,
			DefineViaAlias:      true,
			SkipOptionalPointer: out.SkipOptionalPointer,
			AdditionalTypes: append(out.AdditionalTypes, TypeDefinition{
				TypeName: typeName,
				JsonName: strings.Join(path, "."),
				Schema:   out,
			}),
		}
	}
	return out, nil
}

// annotatesOnly reports whether an allOf member only annotates the others:
// it is inline, has no x-go-type, doesn't name enum values (which makes a new
// enum type, with those names), and has no keyword that shapes a Go type (see
// typeKeywords).
func annotatesOnly(ref *openapi3.SchemaRef) bool {
	if ref == nil || ref.Ref != "" || ref.Value == nil {
		return false
	}
	for _, ext := range []string{extPropGoType, extEnumVarNames, extEnumNames} {
		if _, ok := ref.Value.Extensions[ext]; ok {
			return false
		}
	}
	return len(typeKeywords(*ref.Value)) == 0
}

// typeKeywords lists the keywords of a schema that shape a Go type, by their
// JSON names (see shallowTypeKeywords), leaving out a oneOf or anyOf whose
// branches only add constraints (see isConstraintOnlyUnionV3).
func typeKeywords(s openapi3.Schema) []string {
	return shallowTypeKeywords(*withoutConstraintOnlyUnions(&s, nil))
}

// shallowTypeKeywords lists the keywords of a schema that shape a Go type, by
// their JSON names, whatever the branches of its oneOf and anyOf are. Those
// are all of them but documentation, nullability, validation constraints
// oapi-codegen doesn't turn into Go types (minLength, pattern, maxItems, ...)
// and extensions. It clears those and lists what's left, so a keyword
// kin-openapi adds later counts as shaping the type until it's added here.
func shallowTypeKeywords(s openapi3.Schema) []string {
	// Documentation.
	s.Title, s.Description, s.Comment = "", "", ""
	s.Default, s.Example, s.Examples = nil, nil, nil
	s.ExternalDocs, s.XML = nil, nil
	s.Deprecated, s.ReadOnly, s.WriteOnly, s.AllowEmptyValue = false, false, false, false
	// Nullability. A 3.1 type array of only "null" says nullable.
	s.Nullable = false
	if len(nonNullTypes(s.Type)) == 0 {
		s.Type = nil
	}
	// Validation.
	s.Min, s.Max, s.MultipleOf = nil, nil, nil
	s.ExclusiveMin, s.ExclusiveMax = openapi3.ExclusiveBound{}, openapi3.ExclusiveBound{}
	s.MinLength, s.MaxLength, s.Pattern = 0, nil, ""
	s.MinItems, s.MaxItems, s.UniqueItems = 0, nil, false
	s.MinProps, s.MaxProps = 0, nil
	s.Not = nil
	// Extensions, and kin-openapi's source locations.
	s.Extensions, s.Origin = nil, nil

	var keywords []string
	v := reflect.ValueOf(s)
	for i := range v.NumField() {
		f := v.Field(i)
		if f.IsZero() || ((f.Kind() == reflect.Slice || f.Kind() == reflect.Map) && f.Len() == 0) {
			continue
		}
		name, _, _ := strings.Cut(v.Type().Field(i).Tag.Get("json"), ",")
		keywords = append(keywords, name)
	}
	return keywords
}

// opaqueMergeError reports an allOf that would merge member, which is or
// includes the opaque schema target, with what with describes.
func opaqueMergeError(member, target *openapi3.SchemaRef, with string) error {
	what := describeAllOfMember(member)
	if member != target {
		what += " (whose allOf includes " + describeAllOfMember(target) + ")"
	}
	// How the spec's OpenAPI version says nullable.
	annotations := "description, nullable, ..."
	if globalState.is31 {
		annotations = `description, type: "null", ...`
	}
	if isRefInExternalDocument(target.Ref) {
		return fmt.Errorf("allOf can't merge %s with %s: a reference to another document can't be merged "+
			"with other allOf members, only annotated (%s). "+
			"Define the schema in this document",
			what, with, annotations)
	}
	replaced := "an inline schema"
	if target.Ref != "" {
		replaced = target.Ref
	}
	return fmt.Errorf("allOf can't merge %s with %s: x-go-type replaces %s with %v, whose fields are unknown, "+
		"so it can only be annotated (%s). "+
		"Give the composition an x-go-type of its own",
		what, with, replaced, combinedSchemaExtensions(target)[extPropGoType], annotations)
}

// describeAllOfMember names an allOf member for an error message: its $ref,
// or what an inline member declares.
func describeAllOfMember(ref *openapi3.SchemaRef) string {
	if ref.Ref != "" {
		if ref.Ref[0] == '#' && isRefInExternalDocument(ref.Ref) {
			return ref.Ref + " (a $ref to another document)"
		}
		return ref.Ref
	}
	if ref.Value == nil {
		return "an inline schema"
	}
	declares := typeKeywords(*ref.Value)
	for _, ext := range []string{extEnumVarNames, extEnumNames} {
		if _, ok := ref.Value.Extensions[ext]; ok {
			declares = append(declares, ext)
		}
	}
	var with []string
	if goType, ok := ref.Value.Extensions[extPropGoType]; ok {
		with = append(with, fmt.Sprintf("x-go-type %v", goType))
	}
	if len(declares) > 0 {
		with = append(with, strings.Join(declares, ", "))
	}
	if len(with) == 0 {
		return "an inline schema"
	}
	return "an inline schema with " + strings.Join(with, " and ")
}

// allOfMemberLabel names an allOf member for error messages: the label the
// generator gave a member it made up, the member's $ref, or its place in the
// allOf of the member labeled parent ("" for the composition itself).
func allOfMemberLabel(ctx genContext, parent string, member *openapi3.SchemaRef, i int) string {
	if label, ok := ctx.memberLabels[member]; ok {
		return label
	}
	if member.Ref != "" {
		return member.Ref
	}
	return childLabel(parent, fmt.Sprintf("allOf/%d", i))
}

// allOfMerge merges the members of an allOf into one schema, keyword by
// keyword, keeping the keywords that shape a Go type:
//
//   - type: the types every member allows. integer narrows number, and "null"
//     is nullability, which any member can add (see nullable).
//   - format, const, discriminator: the one a member declares, or the same one
//     declared by several.
//   - enum: the values every member's enum allows, with their names
//     (x-enum-varnames). A composition that adds values to an enum can ask
//     for the values any member's enum allows instead, with
//     x-oapi-codegen-enum-merge: union.
//   - required: every member's.
//   - nullable, readOnly, writeOnly: set when any member sets them. Read
//     literally, a nullable member would change nothing unless every member
//     were nullable, but a member only says it to make the composed type
//     nullable, as in OpenAPI 3.0's `allOf: [$ref X, {nullable: true}]`
//     (issue #1898).
//   - properties, items, additionalProperties: a schema that only one member
//     declares, or that several declare alike, is kept as it is. Different
//     schemas that several members declare for the same position are merged
//     the way allOf merges them, as an allOf of the lot (issue #2107), so
//     `allOf: [$ref Base, {properties: {name: {nullable: true}}}]` makes
//     Base's name nullable instead of replacing it. additionalProperties:
//     false on any member closes the composed object: read literally, the
//     other members' properties could never appear.
//   - oneOf, anyOf: collected from every member.
//
// Documentation and validation constraints never conflict. Any other
// disagreement is an error naming both members: no Go type can hold what the
// composition describes, and picking one member's word would silently
// generate a type that disagrees with the other.
type allOfMerge struct {
	ctx    genContext
	schema openapi3.Schema
	// properties, items and additional collect each member's schema for
	// those positions, merged by result.
	properties    map[string][]labeledSchema
	items         []labeledSchema
	additional    []labeledSchema
	anyAdditional bool
	closed        bool
	// from records the member each keyword that can conflict came from.
	from map[string]string
	// enumNames are the names of schema.Enum's values, or nil.
	enumNames []string
	// renames are names from a member that has no enum of its own, such as
	// `allOf: [$ref Color, {x-enum-varnames: [...]}]`.
	renames []string
	// nullInType records a 3.1 "null" in a member's type array.
	nullInType bool
	// unionEnums merges enums into their union rather than their
	// intersection.
	unionEnums bool
	// madeUp is set when the composition is one the merge made for a
	// position several members declare, rather than one in the spec.
	madeUp bool
	// composition is the allOf being merged. A member's oneOf or anyOf may
	// restate the type another member declares (see isConstraintOnlyUnionV3).
	composition *openapi3.Schema
	// flattening holds the members' schemas the merge has reached, directly
	// or in a nested allOf (see listsFlattened).
	flattening map[*openapi3.Schema]bool
	// components are the oneOfs and anyOfs the members bring, each a union of
	// its own.
	components []unionComponent
}

// unionComponent is a oneOf or anyOf that an allOf's merge collects from one
// of its members. An allOf of several is a value that is one of each union's
// variants at once: `allOf: [$ref Payment, $ref Delivery]` is a card or a
// transfer, and a courier or a pickup. The generated type keeps one JSON
// value; each variant of each union parses it on demand, and setting a
// variant replaces only the keys its own union owns (see
// Schema.UnionOwnedKeys).
type unionComponent struct {
	branches      openapi3.SchemaRefs
	anyOf         bool
	discriminator *openapi3.Discriminator
	// ref is the member when it is a $ref to a schema that is only this
	// union, which then gets As* and From* of its own.
	ref *openapi3.SchemaRef
	// label names the member the union comes from, for errors.
	label string
}

// addComponent adds a union to the merge, unless one with the same branches
// is already there. Then the two are one union: a oneOf if either is, with
// the $ref and the discriminator either has.
func (m *allOfMerge) addComponent(c unionComponent) {
	for i, existing := range m.components {
		if sameBranches(existing.branches, c.branches) {
			if existing.anyOf && !c.anyOf {
				existing.anyOf, existing.label = false, c.label
			}
			existing.ref = cmp.Or(existing.ref, c.ref)
			existing.discriminator = cmp.Or(existing.discriminator, c.discriminator)
			m.components[i] = existing
			return
		}
	}
	m.components = append(m.components, c)
}

// placeDiscriminator decides which of several unions the composition's
// discriminator d tells apart: the one that declares it, or else, for one
// declared elsewhere (next to the allOf, or in a member of its own), the one
// whose variants its mapping names or, without a mapping, the one whose
// variants all have the property. A value has one discriminator, so it is an
// error for it to fit several unions, or none.
func (m *allOfMerge) placeDiscriminator(d *openapi3.Discriminator) error {
	var carriers, labels []string
	var candidates []int
	for i, c := range m.components {
		labels = append(labels, displayLabel(c.label))
		if c.discriminator != nil {
			carriers = append(carriers, displayLabel(c.label))
		}
		if unionFits(c.branches, d) {
			candidates = append(candidates, i)
		}
	}
	switch {
	case len(carriers) == 1:
		return nil
	case len(carriers) > 1:
		return fmt.Errorf("allOf can't use the discriminator %s for both %s and %s: a value has one discriminator",
			d.PropertyName, carriers[0], carriers[1])
	case len(candidates) != 1:
		return fmt.Errorf("allOf can't tell which of its unions (%s) the discriminator %s is for; declare it next to that union's oneOf or anyOf",
			strings.Join(labels, ", "), d.PropertyName)
	}
	m.components[candidates[0]].discriminator = d
	return nil
}

// unionFits reports whether the discriminator d can tell the union's branches
// apart: its mapping names one of them, or, without a mapping, every branch
// but a null one declares its property.
func unionFits(branches openapi3.SchemaRefs, d *openapi3.Discriminator) bool {
	if len(d.Mapping) > 0 {
		for _, b := range branches {
			for _, target := range d.Mapping {
				if b != nil && b.Ref != "" && b.Ref == target.Ref {
					return true
				}
			}
		}
		return false
	}
	declared := false
	for _, b := range branches {
		if b == nil || b.Value == nil || isNullTypeSchema(b.Value) {
			continue
		}
		if findProperty(b.Value, d.PropertyName, 0) == nil {
			return false
		}
		declared = true
	}
	return declared
}

// sameBranches reports whether two lists of union branches are the same.
func sameBranches(a, b openapi3.SchemaRefs) bool {
	return slices.EqualFunc(a, b, func(x, y *openapi3.SchemaRef) bool {
		return x == y || (x != nil && y != nil && sameSchema(x, y))
	})
}

// isUnionOnly reports whether a schema is only a oneOf or anyOf: a union type
// with nothing else that shapes a Go type but its discriminator and a
// restated `type: object`.
func isUnionOnly(v openapi3.Schema) bool {
	if (len(v.OneOf) == 0) == (len(v.AnyOf) == 0) {
		return false
	}
	own := v
	own.OneOf, own.AnyOf, own.Discriminator, own.Required = nil, nil, nil, nil
	for _, keyword := range shallowTypeKeywords(own) {
		if keyword != "type" || !sameTypes(nonNullTypes(own.Type), []string{openapi3.TypeObject}) {
			return false
		}
	}
	return true
}

// markFlattening records member, and the members of its allOf, all the way
// down, as schemas the merge flattens (see listsFlattened), before any is
// merged, so which unions are left out doesn't depend on the members' order.
func (m *allOfMerge) markFlattening(member *openapi3.SchemaRef) {
	for r := range allOfTree(member) {
		m.flattening[r.Value] = true
	}
}

// listsFlattened reports whether a oneOf or anyOf lists the composition being
// merged, or a schema the merge is flattening into it. That's the inheritance
// style some generators write, where a parent lists its children and each
// child is an allOf of the parent: `Pet: {oneOf: [Cat, Dog]}` and
// `Cat: {allOf: [$ref Pet, {...}]}`. Merging Pet into Cat would make every
// child a union of all of them, with Cat.AsDog(); but a Cat already is the Cat
// branch, so the list says nothing more about it and is left out.
func (m *allOfMerge) listsFlattened(branches openapi3.SchemaRefs) bool {
	for _, b := range branches {
		if b == nil || b.Value == nil {
			continue
		}
		if b.Value == m.ctx.composing || m.flattening[b.Value] {
			return true
		}
	}
	return false
}

// labeledSchema is a member's schema for a position, with its label.
type labeledSchema struct {
	ref   *openapi3.SchemaRef
	label string
}

func newAllOfMerge(ctx genContext) *allOfMerge {
	return &allOfMerge{
		ctx:        ctx,
		unionEnums: ctx.unionEnums,
		madeUp:     ctx.mergingMadeUp,
		schema:     openapi3.Schema{Extensions: map[string]any{}},
		from:       map[string]string{},
		properties: map[string][]labeledSchema{},
		flattening: map[*openapi3.Schema]bool{},
	}
}

// add merges one member's schema v, named by label for error messages. member
// is the member itself when the merge has one to hand, or nil: a member that
// is a $ref to a union gets As* and From* of its own. A member that is an
// allOf itself contributes its members, then its own keywords. seen holds the
// $refs being flattened, so that a cycle back into one is skipped.
func (m *allOfMerge) add(member *openapi3.SchemaRef, v openapi3.Schema, label string, seen map[string]bool) error {
	// A member that is an allOf is merged its own way first when that differs
	// from this composition's, so it has the values it has on its own:
	// decorating a composition that adds values to an enum doesn't take them
	// away again, and a union doesn't add values a nested composition rules out.
	if len(v.AllOf) > 0 {
		union := false
		if raw, ok := v.Extensions[extOapiCodegenEnumMerge]; ok {
			var err error
			if union, err = extParseEnumMerge(raw); err != nil {
				return fmt.Errorf("invalid value for %q in %s: %w", extOapiCodegenEnumMerge, displayLabel(label), err)
			}
		}
		if union != m.unionEnums {
			own := newAllOfMerge(m.ctx)
			own.unionEnums = union
			own.composition = m.composition
			own.flattening = m.flattening
			if err := own.add(member, v, label, seen); err != nil {
				return err
			}
			merged, err := own.result()
			if err != nil {
				return err
			}
			for _, c := range own.components {
				m.addComponent(c)
			}
			merged.OneOf, merged.AnyOf = nil, nil
			return m.add(nil, merged, label, seen)
		}
	}
	for j, inner := range v.AllOf {
		if inner.Ref != "" {
			if seen[inner.Ref] {
				continue
			}
			seen[inner.Ref] = true
		}
		iv, err := memberValue(inner)
		if err != nil {
			return err
		}
		if err := m.add(inner, iv, allOfMemberLabel(m.ctx, label, inner, j), seen); err != nil {
			return err
		}
	}

	for k, ext := range v.Extensions {
		if k != extEnumVarNames && k != extEnumNames {
			m.schema.Extensions[k] = ext
		}
	}
	// Each oneOf and anyOf is a union of its own (see unionComponent). One that
	// only adds constraints makes no union, and nor does one that lists a
	// schema this merge is part of (see listsFlattened). A list constrains the
	// member, or, when the member declares no type, the composition, whose
	// other members may declare the type its branches restate. A discriminator
	// goes with the list it tells apart, the oneOf when there are both: a
	// parent's discriminator says which child a value is, not which branch of
	// a union the child has of its own.
	if member == nil || member.Ref == "" || !isUnionOnly(v) {
		member = nil
	}
	owner := &v
	if len(declaredTypes(owner)) == 0 && m.composition != nil {
		owner = m.composition
	}
	oneOf := len(v.OneOf) > 0 && !isConstraintOnlyUnionV3(v.OneOf, owner) && !m.listsFlattened(v.OneOf)
	anyOf := len(v.AnyOf) > 0 && !isConstraintOnlyUnionV3(v.AnyOf, owner) && !m.listsFlattened(v.AnyOf)
	if (len(v.OneOf) > 0 && !oneOf) || (len(v.OneOf) == 0 && len(v.AnyOf) > 0 && !anyOf) {
		v.Discriminator = nil
	}
	anyOfDiscriminator := v.Discriminator
	if len(v.OneOf) > 0 {
		anyOfDiscriminator = nil
	}
	if anyOf {
		m.addComponent(unionComponent{branches: v.AnyOf, anyOf: true, discriminator: anyOfDiscriminator, ref: member, label: label})
	}
	if oneOf {
		m.addComponent(unionComponent{branches: v.OneOf, discriminator: v.Discriminator, ref: member, label: label})
	}

	if err := m.addType(v, label); err != nil {
		return err
	}
	if v.Format != "" {
		if m.schema.Format != "" && m.schema.Format != v.Format {
			return mergeConflict(m.from["format"], "format "+m.schema.Format, label, "format "+v.Format,
				"a value can't have both formats")
		}
		if m.schema.Format == "" {
			m.schema.Format, m.from["format"] = v.Format, label
		}
	}
	if err := m.addEnum(v, label); err != nil {
		return err
	}
	if v.Const != nil {
		if m.schema.Const != nil && !reflect.DeepEqual(m.schema.Const, v.Const) {
			return mergeConflict(m.from["const"], fmt.Sprintf("const %v", m.schema.Const), label,
				fmt.Sprintf("const %v", v.Const), "no value is both")
		}
		if m.schema.Const == nil {
			m.schema.Const, m.from["const"] = v.Const, label
		}
	}
	if v.Discriminator != nil {
		if m.schema.Discriminator != nil && !sameDiscriminator(m.schema.Discriminator, v.Discriminator) {
			return mergeConflict(m.from["discriminator"], "discriminator "+m.schema.Discriminator.PropertyName,
				label, "discriminator "+v.Discriminator.PropertyName, "a value has one discriminator")
		}
		if m.schema.Discriminator == nil {
			m.schema.Discriminator, m.from["discriminator"] = v.Discriminator, label
		}
	}

	// Annotations: a default is kept from the last member that has one.
	if v.Default != nil {
		m.schema.Default = v.Default
	}
	if schemaIsNullable(&v) {
		m.schema.Nullable = true
	}
	m.schema.ReadOnly = m.schema.ReadOnly || v.ReadOnly
	m.schema.WriteOnly = m.schema.WriteOnly || v.WriteOnly
	m.schema.AllowEmptyValue = m.schema.AllowEmptyValue || v.AllowEmptyValue

	for _, name := range v.Required {
		m.schema.Required = appendUnique(m.schema.Required, name)
	}

	for name, p := range v.Properties {
		m.properties[name] = append(m.properties[name], labeledSchema{p, childLabel(label, "properties/"+name)})
	}
	if v.Items != nil {
		m.items = append(m.items, labeledSchema{v.Items, childLabel(label, "items")})
	}
	switch {
	case isAdditionalPropertiesExplicitFalse(&v):
		m.closed = true
	case v.AdditionalProperties.Schema != nil:
		m.additional = append(m.additional, labeledSchema{v.AdditionalProperties.Schema, childLabel(label, "additionalProperties")})
	case v.AdditionalProperties.Has != nil:
		m.anyAdditional = true
	}
	return nil
}

// subschema merges the schemas members declare for one position (a property,
// items or additionalProperties): one schema, or several alike, is that
// schema, and different ones become an allOf of them, which generates the way
// any allOf does.
//
// The same schemas always make the same allOf. A schema that refers back to
// the composition merges the same schemas again, and generating the same allOf
// again is what genContext.inProgress recognises as recursion.
func (m *allOfMerge) subschema(schemas []labeledSchema) *openapi3.SchemaRef {
	switch len(schemas) {
	case 0:
		return nil
	case 1:
		return schemas[0].ref
	}
	// The result is kept by the schemas it's made of, so the same schemas
	// always give the same result, even when that is a schema made here.
	keys := make([]string, len(schemas))
	for i, s := range schemas {
		keys[i] = fmt.Sprintf("%p", s.ref)
	}
	key := strings.Join(keys, " ")
	if merged, ok := m.ctx.subschemas[key]; ok {
		return merged
	}

	var distinct []labeledSchema
	for _, s := range schemas {
		i := slices.IndexFunc(distinct, func(d labeledSchema) bool { return d.ref == s.ref || sameSchema(d.ref, s.ref) })
		switch {
		case i < 0:
			distinct = append(distinct, s)
		case s.ref.Ref != "" && len(s.ref.Extensions) > 0:
			// The same $ref, with extensions next to it, such as x-go-name.
			// It stays one $ref, keeping its type, with both members'
			// extensions; the later member's win, as they do in the merge.
			ext := maps.Clone(distinct[i].ref.Extensions)
			if ext == nil {
				ext = make(map[string]any, len(s.ref.Extensions))
			}
			maps.Copy(ext, s.ref.Extensions)
			distinct[i].ref = &openapi3.SchemaRef{Ref: s.ref.Ref, Value: s.ref.Value, Extensions: ext}
		}
	}
	if len(distinct) == 1 {
		m.ctx.subschemas[key] = distinct[0].ref
		return distinct[0].ref
	}

	merged := &openapi3.Schema{Extensions: map[string]any{}}
	for _, d := range distinct {
		// A copy of the member, so that its label belongs to this allOf.
		member := &openapi3.SchemaRef{Ref: d.ref.Ref, Value: d.ref.Value, Extensions: d.ref.Extensions}
		m.ctx.memberLabels[member] = d.label
		merged.AllOf = append(merged.AllOf, member)

		// What the generator reads from a property's own schema rather than
		// from its type: documentation, and field-level extensions such as
		// x-go-name. The members' types merge through the allOf instead.
		v := d.ref.Value
		if v == nil {
			continue
		}
		if v.Description != "" {
			merged.Description = v.Description
		}
		if v.Example != nil {
			merged.Example = v.Example
		}
		if len(v.Examples) > 0 {
			merged.Examples = v.Examples
		}
		merged.Deprecated = merged.Deprecated || v.Deprecated
		merged.ReadOnly = merged.ReadOnly || v.ReadOnly
		merged.WriteOnly = merged.WriteOnly || v.WriteOnly
		for k, ext := range combinedSchemaExtensions(d.ref) {
			switch k {
			case extPropGoType, extPropGoImport, extGoTypeName, extEnumVarNames, extEnumNames:
			default:
				merged.Extensions[k] = ext
			}
		}
	}
	ref := &openapi3.SchemaRef{Value: merged}
	m.ctx.subschemas[key] = ref
	m.ctx.madeUp[merged] = true
	return ref
}

// addType intersects the member's types with those merged so far.
func (m *allOfMerge) addType(v openapi3.Schema, label string) error {
	if slices.Contains(v.Type.Slice(), openapi3.TypeNull) {
		m.nullInType = true
	}
	types := nonNullTypes(v.Type)
	if len(types) == 0 {
		return nil
	}
	merged := m.schema.Type.Slice()
	if len(merged) == 0 {
		m.schema.Type, m.from["type"] = (*openapi3.Types)(&types), label
		return nil
	}
	both := intersectTypes(merged, types)
	if len(both) == 0 {
		return mergeConflict(m.from["type"], "type "+strings.Join(merged, ", "), label,
			"type "+strings.Join(types, ", "), "no value has both types")
	}
	m.schema.Type = (*openapi3.Types)(&both)
	return nil
}

// intersectTypes returns the JSON Schema types in both a and b. An integer is
// a number, so integer and number have integer in common, and a list with
// both says number.
func intersectTypes(a, b []string) []string {
	a, b = withoutIntegerUnderNumber(a), withoutIntegerUnderNumber(b)
	var both []string
	for _, t := range a {
		switch {
		case slices.Contains(b, t):
		case t == openapi3.TypeInteger && slices.Contains(b, openapi3.TypeNumber),
			t == openapi3.TypeNumber && slices.Contains(b, openapi3.TypeInteger):
			t = openapi3.TypeInteger
		default:
			continue
		}
		both = appendUnique(both, t)
	}
	return both
}

// withoutIntegerUnderNumber drops integer from a type list that also has
// number, which already allows every integer.
func withoutIntegerUnderNumber(types []string) []string {
	if !slices.Contains(types, openapi3.TypeNumber) || !slices.Contains(types, openapi3.TypeInteger) {
		return types
	}
	return slices.DeleteFunc(slices.Clone(types), func(t string) bool { return t == openapi3.TypeInteger })
}

// addEnum intersects the member's enum with the values merged so far, keeping
// each value's name.
func (m *allOfMerge) addEnum(v openapi3.Schema, label string) error {
	names := memberEnumNames(v)
	if len(v.Enum) == 0 {
		if names != nil {
			m.renames = names
		}
		return nil
	}
	if len(names) != len(v.Enum) {
		names = nil
	}
	if m.schema.Enum == nil {
		m.schema.Enum, m.enumNames, m.from["enum"] = v.Enum, names, label
		return nil
	}
	// A value without a name is named after itself, as it would be if no
	// member named any; the names are only kept when some member does.
	named := m.enumNames != nil || names != nil
	nameOf := func(names []string, i int, value any) string {
		if names != nil {
			return names[i]
		}
		return fmt.Sprintf("%v", value)
	}
	// A value both members name takes the later member's name, as later
	// members' annotations win elsewhere in the merge.
	nameOfBoth := func(i, j int, value any) string {
		if names != nil {
			return names[j]
		}
		return nameOf(m.enumNames, i, value)
	}
	var enum []any
	var enumNames []string
	if m.unionEnums {
		for i, value := range m.schema.Enum {
			enum = append(enum, value)
			if j := slices.IndexFunc(v.Enum, func(other any) bool { return reflect.DeepEqual(value, other) }); j >= 0 {
				enumNames = append(enumNames, nameOfBoth(i, j, value))
			} else {
				enumNames = append(enumNames, nameOf(m.enumNames, i, value))
			}
		}
		for j, value := range v.Enum {
			if !slices.ContainsFunc(enum, func(other any) bool { return reflect.DeepEqual(value, other) }) {
				enum = append(enum, value)
				enumNames = append(enumNames, nameOf(names, j, value))
			}
		}
	} else {
		for i, value := range m.schema.Enum {
			j := slices.IndexFunc(v.Enum, func(other any) bool { return reflect.DeepEqual(value, other) })
			if j < 0 {
				continue
			}
			enum = append(enum, value)
			enumNames = append(enumNames, nameOfBoth(i, j, value))
		}
	}
	if len(enum) == 0 {
		// The extension applies to the schema it is on. A property that
		// members declare separately is merged by an allOf the spec doesn't
		// spell out, so there it goes on the member's property.
		where := "the schema with this allOf"
		if m.madeUp {
			where = displayLabel(label)
		}
		return mergeConflict(m.from["enum"], fmt.Sprintf("enum %v", m.schema.Enum), label,
			fmt.Sprintf("enum %v", v.Enum), "no value is in both. To allow the values of either, "+
				"set x-oapi-codegen-enum-merge: union on "+where)
	}
	if !named {
		enumNames = nil
	}
	m.schema.Enum, m.enumNames = enum, enumNames
	return nil
}

// memberEnumNames returns the names a schema gives its enum values, or nil.
func memberEnumNames(v openapi3.Schema) []string {
	for _, key := range []string{extEnumVarNames, extEnumNames} {
		if ext, ok := v.Extensions[key]; ok {
			if names, err := extParseEnumVarNames(ext); err == nil {
				return names
			}
		}
	}
	return nil
}

// result returns the merged schema.
func (m *allOfMerge) result() (openapi3.Schema, error) {
	s := m.schema
	if len(m.properties) > 0 {
		s.Properties = make(openapi3.Schemas, len(m.properties))
		for name, schemas := range m.properties {
			s.Properties[name] = m.subschema(schemas)
		}
	}
	s.Items = m.subschema(m.items)
	switch {
	case m.closed:
		s.WithoutAdditionalProperties()
	case len(m.additional) > 0:
		s.AdditionalProperties = openapi3.AdditionalProperties{Schema: m.subschema(m.additional)}
	case m.anyAdditional:
		s.WithAnyAdditionalProperties()
	}
	if m.nullInType {
		types := append(slices.Clone(s.Type.Slice()), openapi3.TypeNull)
		s.Type = (*openapi3.Types)(&types)
	}
	// A oneOf or anyOf whose branches restate a type another member declares
	// only adds constraints too, which only the whole composition shows. One
	// of only 3.1 null branches says nullable, which is read where the schema
	// is used.
	m.components = slices.DeleteFunc(m.components, func(c unionComponent) bool {
		return isConstraintOnlyUnionV3(c.branches, &s) || !slices.ContainsFunc(c.branches, func(b *openapi3.SchemaRef) bool {
			return b == nil || b.Value == nil || !isNullTypeSchema(b.Value)
		})
	})
	if len(m.components) > 1 && s.Discriminator != nil {
		if err := m.placeDiscriminator(s.Discriminator); err != nil {
			return openapi3.Schema{}, err
		}
	}
	// A single union is generated as it always was. Several are one union
	// for the generator, whose variants unionComponents groups again.
	for _, c := range m.components {
		if c.anyOf && len(m.components) == 1 {
			s.AnyOf = c.branches
		} else {
			s.OneOf = append(slices.Clone(s.OneOf), c.branches...)
		}
	}
	if s.Const != nil && s.Enum != nil {
		i := slices.IndexFunc(s.Enum, func(value any) bool { return reflect.DeepEqual(value, s.Const) })
		if i < 0 {
			return openapi3.Schema{}, mergeConflict(m.from["const"], fmt.Sprintf("const %v", s.Const),
				m.from["enum"], fmt.Sprintf("enum %v", s.Enum), "the enum doesn't allow the const")
		}
		s.Enum = s.Enum[i : i+1]
		if m.enumNames != nil {
			m.enumNames = m.enumNames[i : i+1]
		}
	}
	names := m.enumNames
	if m.renames != nil && len(m.renames) == len(s.Enum) {
		names = m.renames
	}
	if names != nil {
		ext := make([]any, len(names))
		for i, name := range names {
			ext[i] = name
		}
		s.Extensions[extEnumVarNames] = ext
	}
	return s, nil
}

// sameDiscriminator reports whether two discriminators name the same property
// with the same mapping.
func sameDiscriminator(a, b *openapi3.Discriminator) bool {
	if a == b {
		return true
	}
	if a.PropertyName != b.PropertyName || len(a.Mapping) != len(b.Mapping) {
		return false
	}
	for key, ref := range a.Mapping {
		if other, ok := b.Mapping[key]; !ok || other.Ref != ref.Ref {
			return false
		}
	}
	return true
}

// childLabel names a position inside a member for error messages, relative to
// the composition: "allOf/1" and "properties/name" make "allOf/1/properties/name".
// The composition itself is "".
func childLabel(label, child string) string {
	if label == "" {
		return child
	}
	return label + "/" + child
}

// displayLabel shows a member's label in an error message.
func displayLabel(label string) string {
	if label == "" {
		return "the schema itself"
	}
	return label
}

// mergeConflict reports two members that no value can satisfy together.
func mergeConflict(labelA, a, labelB, b, why string) error {
	return fmt.Errorf("allOf can't merge %s (%s) with %s (%s): %s", displayLabel(labelA), a, displayLabel(labelB), b, why)
}

// hasStructuralSiblingsV3 reports whether a schema with allOf also has
// keywords of its own that shape a Go type (see typeKeywords), which are then
// merged as one more member (see compositionMembers). Documentation and
// nullability aren't, so `{allOf: [$ref X], nullable: true}` stays X; nor is
// a oneOf or anyOf whose branches only add constraints (see
// isConstraintOnlyUnionV3). A restated type is merged, and only annotates the
// $ref it restates, so `{type: object, allOf: [$ref X]}` stays X too.
func hasStructuralSiblingsV3(s *openapi3.Schema) bool {
	if s == nil {
		return false
	}
	// The schema's own oneOf or anyOf may only add constraints to what its
	// allOf members declare, so judge them against the whole schema.
	own := *withoutConstraintOnlyUnions(s, s)
	own.AllOf = nil
	return len(shallowTypeKeywords(own)) > 0
}

// isConstraintOnlyUnionV3 reports whether a oneOf or anyOf only adds
// constraints to owner, the schema it constrains, such as
// `oneOf: [{required: [email]}, {required: [phone]}]`, rather than naming
// alternative types. Every branch declares nothing that shapes a Go type but
// `required` and a boolean `additionalProperties`, and at most restates
// owner's type (see declaredTypes). v3 generates no union for such a list,
// wherever it is (issue #839): a union of `any` branches would make the type
// harder to use, not more correct.
//
// A local $ref branch is judged by the schema it refers to. A $ref into
// another document is a type, and so is a branch with x-go-type or
// x-go-type-name, which asks for a Go type. A 3.1 `{type: "null"}` branch
// says the schema is nullable, which is read from the schema where its type
// is used, so it is passed over; a list of nothing else isn't one of
// constraints.
func isConstraintOnlyUnionV3(branches openapi3.SchemaRefs, owner *openapi3.Schema) bool {
	constraints := 0
	for _, b := range branches {
		if b == nil || b.Value == nil || isRefInExternalDocument(b.Ref) {
			return false
		}
		if isNullTypeSchema(b.Value) {
			continue
		}
		// A branch that asks for a Go type, or names the one it gets, is a
		// type.
		extensions := combinedSchemaExtensions(b)
		if _, ok := extensions[extPropGoType]; ok {
			return false
		}
		if _, ok := extensions[extGoTypeName]; ok {
			return false
		}
		for _, keyword := range shallowTypeKeywords(*b.Value) {
			switch keyword {
			case "required":
			case "additionalProperties":
				if b.Value.AdditionalProperties.Schema != nil {
					return false
				}
			case "type":
				if !sameTypes(declaredTypes(owner), nonNullTypes(b.Value.Type)) {
					return false
				}
			default:
				return false
			}
		}
		constraints++
	}
	return constraints > 0
}

// withoutConstraintOnlyUnions returns s without a oneOf or anyOf that only
// adds constraints to owner, which is s when nil (see
// isConstraintOnlyUnionV3). v3 generates no union for such a list.
func withoutConstraintOnlyUnions(s, owner *openapi3.Schema) *openapi3.Schema {
	if owner == nil {
		owner = s
	}
	oneOf, anyOf := isConstraintOnlyUnionV3(s.OneOf, owner), isConstraintOnlyUnionV3(s.AnyOf, owner)
	if !oneOf && !anyOf {
		return s
	}
	c := *s
	if oneOf {
		c.OneOf = nil
	}
	if anyOf {
		c.AnyOf = nil
	}
	return &c
}

// generateAllOfV3 lowers a schema with allOf for generateGoSchema: the
// composition merged into one Go type by merge, with the parent's structural
// siblings merged in, its description kept, and a recursive composition named.
func generateAllOfV3(ctx genContext, schema *openapi3.Schema, path []string, extensions map[string]any, skipOptionalPointer bool) (Schema, error) {
	if alias, ok := ctx.inProgressAllOf(schema, skipOptionalPointer); ok {
		return alias, nil
	}
	var err error
	ctx.mergingMadeUp = ctx.madeUp[schema]
	ctx.composing = schema
	ctx.unionEnums = false
	if raw, ok := extensions[extOapiCodegenEnumMerge]; ok {
		if ctx.unionEnums, err = extParseEnumMerge(raw); err != nil {
			return Schema{}, fmt.Errorf("invalid value for %q: %w", extOapiCodegenEnumMerge, err)
		}
	}
	frame := &mergeFrame{typeName: ctx.typeName(path)}
	ctx.inProgress[schema] = frame
	defer delete(ctx.inProgress, schema)

	// The schema's own keywords next to allOf are merged as one more member
	// (see compositionMembers). v3 rejects old-allof-sibling-merging, which
	// discarded them. Issues #697, #931, #1710, #2102.
	members := compositionMembers(schema)
	if len(members) > len(schema.AllOf) {
		own := members[len(members)-1]
		ctx.memberLabels[own] = ""
		defer delete(ctx.memberLabels, own)
		// An opaque member can't be merged with the schema's own keywords
		// either, unless they only annotate it. Say so here, where they can
		// be named as the schema's own.
		for _, m := range schema.AllOf {
			target := opaqueSchemaWithin(m)
			if target == nil || (opaqueSchemaFor(m) == target && annotates(own, target)) {
				continue
			}
			return Schema{}, fmt.Errorf("error merging schemas: %w", opaqueMergeError(m, target,
				"the schema's own "+strings.Join(typeKeywords(*own.Value), ", ")))
		}
	}
	// A single member, such as `allOf: [$ref X]`, is that member's type.
	mergedSchema, err := mergeSchemasV3(ctx, members, path)
	if err != nil {
		return Schema{}, fmt.Errorf("error merging schemas: %w", err)
	}
	return finishAllOf(ctx, frame, schema, mergedSchema, schema.Description, extensions, skipOptionalPointer)
}
