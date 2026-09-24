package codegen

// This file holds the allOf half of schema-merging-behavior v3, the version
// under development. Its anyOf/oneOf half is in union_v3.go. It started as a
// copy of v2's (merge_schemas_v2.go).

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

func mergeSchemasV3(ctx genContext, allOf []*openapi3.SchemaRef, path []string) (Schema, error) {
	n := len(allOf)

	if n == 1 {
		return generateGoSchema(ctx, allOf[0], path)
	}

	// A member whose schema the merge can't read can only be annotated by the
	// others, and then the composition is that member's type.
	opaque, err := opaqueMember(allOf)
	if err != nil {
		return Schema{}, err
	}
	if opaque != nil {
		return generateAnnotatedOpaque(ctx, opaque, allOf, path)
	}

	// Distinguish two uses of allOf:
	//
	//   1. Decorator idiom — at least one INLINE member (Ref == "") is
	//      "extension-only" (carries no structural content). This is a
	//      workaround for OpenAPI 3.0's $ref-sibling restriction: users
	//      wrap a $ref in allOf to attach extensions like
	//      x-go-type-skip-optional-pointer (see issue #1957). Here
	//      extensions are meant to flow through to the result.
	//
	//   2. Real composition — every member either contributes structural
	//      content or is a $ref contributing the referenced schema. The
	//      result is a NEW distinct type, and extensions that name a source
	//      schema's type, such as x-go-type-name, do NOT transfer. (A member
	//      with x-go-type never gets this far: see opaqueMember.)
	//
	// A $ref member is excluded from the decorator check because it is by
	// construction delivering the referenced schema, not "decorating"
	// siblings — even if the referenced schema happens to carry only
	// extensions, that's a property of the target, not an intent on this
	// composition.
	decoratorIdiom := false
	for _, m := range allOf {
		if m.Ref == "" && isExtensionOnlySchemaV3(m.Value) {
			decoratorIdiom = true
			break
		}
	}

	merged := newAllOfMerge(ctx)
	// The $refs of the members merged so far, so that a nested allOf that
	// refers back to one is not flattened into itself.
	seenTopLevel := make(map[string]bool)
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
		if err := merged.add(value, allOfMemberLabel(ctx, member, i), seen); err != nil {
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
		//
		ext := maps.Clone(schema.Extensions)
		delete(ext, extGoTypeName)
		delete(ext, extPropGoImport)
		schema.Extensions = ext
	}

	return generateGoSchema(ctx, openapi3.NewSchemaRef("", &schema), path)
}

// isExtensionOnlySchemaV3 reports whether a schema carries only extensions,
// with no structural or constraint-bearing content. Used to detect the
// "$ref + sibling extension" idiom: allOf wrappers whose purpose is
// attaching extensions to a $ref (since OpenAPI 3.0 disallows sibling
// keys next to $ref).
//
// Implementation: zero out every field that doesn't affect the generated
// Go type, then compare to the zero Schema. Anything left over — a Type,
// Properties, Pattern, MinLength, etc. — disqualifies the schema from
// being treated as a pure decorator. This formulation defaults to safe
// behavior if kin-openapi gains new structural fields: they'd be non-zero
// by default and correctly disqualify.
func isExtensionOnlySchemaV3(s *openapi3.Schema) bool {
	if s == nil || len(s.Extensions) == 0 {
		return false
	}
	tmp := *s
	tmp.Extensions = nil
	// Source-tracking metadata from kin-openapi; always non-nil for
	// schemas parsed from a file.
	tmp.Origin = nil
	// Purely documentary / metadata fields. These don't affect the
	// generated Go type, so a schema carrying only these plus extensions
	// still behaves as a decorator.
	tmp.Title = ""
	tmp.Description = ""
	tmp.Default = nil
	tmp.Example = nil
	tmp.ExternalDocs = nil
	tmp.Deprecated = false
	tmp.ReadOnly = false
	tmp.WriteOnly = false
	tmp.AllowEmptyValue = false
	tmp.XML = nil
	return reflect.DeepEqual(tmp, openapi3.Schema{})
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
// stands for, or nil. That is the member itself when it is opaque, and
// otherwise the opaque member of an allOf that the member is, when that allOf
// only annotates it: such a member generates as the opaque schema's type, so
// it is just as opaque.
func opaqueSchemaFor(ref *openapi3.SchemaRef, seen map[*openapi3.Schema]bool) *openapi3.SchemaRef {
	if ref == nil {
		return nil
	}
	if isOpaqueSchema(ref) {
		return ref
	}
	s := ref.Value
	if s == nil || len(s.AllOf) == 0 || seen[s] || hasStructuralSiblingsV3(s) {
		return nil
	}
	if seen == nil {
		seen = make(map[*openapi3.Schema]bool)
	}
	seen[s] = true
	for _, m := range s.AllOf {
		target := opaqueSchemaFor(m, seen)
		if target == nil {
			continue
		}
		for _, other := range s.AllOf {
			if other != m && !annotatesOnly(other) {
				return nil
			}
		}
		return target
	}
	return nil
}

// opaqueSchemaWithin returns an opaque schema (see isOpaqueSchema) that
// merging ref would have to read, or nil: ref itself, or an opaque member of
// an allOf that the merge flattens out of ref, however deep.
func opaqueSchemaWithin(ref *openapi3.SchemaRef, seen map[*openapi3.Schema]bool) *openapi3.SchemaRef {
	if ref == nil {
		return nil
	}
	if isOpaqueSchema(ref) {
		return ref
	}
	s := ref.Value
	if s == nil || seen[s] {
		return nil
	}
	if seen == nil {
		seen = make(map[*openapi3.Schema]bool)
	}
	seen[s] = true
	for _, m := range s.AllOf {
		if target := opaqueSchemaWithin(m, seen); target != nil {
			return target
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
		target := opaqueSchemaFor(m, nil)
		if target == nil {
			continue
		}
		for _, other := range allOf {
			if other != m && !annotatesOnly(other) {
				return nil, opaqueMergeError(m, target, describeAllOfMember(other))
			}
		}
		return m, nil
	}
	for i, m := range allOf {
		if target := opaqueSchemaWithin(m, nil); target != nil {
			other := allOf[0]
			if i == 0 {
				other = allOf[1]
			}
			return nil, opaqueMergeError(m, target, describeAllOfMember(other))
		}
	}
	return nil, nil
}

// generateAnnotatedOpaque generates an allOf of an opaque member and members that
// only annotate it: the opaque member's type. Nullability needs nothing here,
// since schemaIsNullable finds a nullable member wherever the type is used.
func generateAnnotatedOpaque(ctx genContext, opaque *openapi3.SchemaRef, allOf []*openapi3.SchemaRef, path []string) (Schema, error) {
	out, err := generateGoSchema(ctx, opaque, path)
	if err != nil {
		return Schema{}, err
	}
	var typeName string
	for _, m := range allOf {
		if m == opaque {
			continue
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
		// use it here.
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
// it is inline, has no x-go-type, and has no keyword that shapes a Go type
// (see typeKeywords).
func annotatesOnly(ref *openapi3.SchemaRef) bool {
	if ref == nil || ref.Ref != "" || ref.Value == nil {
		return false
	}
	if _, ok := ref.Value.Extensions[extPropGoType]; ok {
		return false
	}
	return len(typeKeywords(*ref.Value)) == 0
}

// typeKeywords lists the keywords of a schema that shape a Go type, by their
// JSON names. Those are all of them but documentation, nullability, validation
// constraints oapi-codegen doesn't turn into Go types (minLength, pattern,
// maxItems, ...) and extensions. It clears those and lists what's left, so a
// keyword kin-openapi adds later counts as shaping the type until it's added
// here.
func typeKeywords(s openapi3.Schema) []string {
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
	var declares []string
	if goType, ok := ref.Value.Extensions[extPropGoType]; ok {
		declares = append(declares, fmt.Sprintf("x-go-type %v", goType))
	}
	declares = append(declares, typeKeywords(*ref.Value)...)
	if len(declares) == 0 {
		return "an inline schema"
	}
	return "an inline schema with " + strings.Join(declares, ", ")
}

// allOfMemberLabel names an allOf member for error messages: the label the
// generator gave a member it made up, the member's $ref, or its place in the
// allOf.
func allOfMemberLabel(ctx genContext, member *openapi3.SchemaRef, i int) string {
	if label, ok := ctx.memberLabels[member]; ok {
		return label
	}
	if member.Ref != "" {
		return member.Ref
	}
	return fmt.Sprintf("allOf/%d", i)
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
		schema:     openapi3.Schema{Extensions: map[string]any{}},
		from:       map[string]string{},
		properties: map[string][]labeledSchema{},
	}
}

// add merges one member, named by label for error messages. A member that is
// an allOf itself contributes its members, then its own keywords. seen holds
// the $refs being flattened, so that a cycle back into one is skipped.
func (m *allOfMerge) add(v openapi3.Schema, label string, seen map[string]bool) error {
	// A member that is an allOf merging its enums its own way is merged that
	// way first, so decorating a composition that adds values to an enum
	// doesn't take them away again.
	if raw, ok := v.Extensions[extOapiCodegenEnumMerge]; ok && len(v.AllOf) > 0 {
		union, err := extParseEnumMerge(raw)
		if err != nil {
			return fmt.Errorf("invalid value for %q in %s: %w", extOapiCodegenEnumMerge, displayLabel(label), err)
		}
		if union != m.unionEnums {
			own := newAllOfMerge(m.ctx)
			own.unionEnums = union
			if err := own.add(v, label, seen); err != nil {
				return err
			}
			merged, err := own.result()
			if err != nil {
				return err
			}
			return m.add(merged, label, seen)
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
		innerLabel, ok := m.ctx.memberLabels[inner]
		switch {
		case ok:
		case inner.Ref != "":
			innerLabel = inner.Ref
		default:
			innerLabel = childLabel(label, fmt.Sprintf("allOf/%d", j))
		}
		if err := m.add(iv, innerLabel, seen); err != nil {
			return err
		}
	}

	for k, ext := range v.Extensions {
		if k != extEnumVarNames && k != extEnumNames {
			m.schema.Extensions[k] = ext
		}
	}
	m.schema.OneOf = append(m.schema.OneOf, v.OneOf...)
	m.schema.AnyOf = append(m.schema.AnyOf, v.AnyOf...)

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
	m.schema.UniqueItems = m.schema.UniqueItems || v.UniqueItems
	if !m.schema.ExclusiveMin.IsSet() {
		m.schema.ExclusiveMin = v.ExclusiveMin
	}
	if !m.schema.ExclusiveMax.IsSet() {
		m.schema.ExclusiveMax = v.ExclusiveMax
	}

	for _, name := range v.Required {
		if !slices.Contains(m.schema.Required, name) {
			m.schema.Required = append(m.schema.Required, name)
		}
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
		i := slices.IndexFunc(distinct, func(d labeledSchema) bool { return d.ref == s.ref || sameSchemaV3(d.ref, s.ref) })
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
		if !slices.Contains(both, t) {
			both = append(both, t)
		}
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
	var enum []any
	var enumNames []string
	if m.unionEnums {
		for i, value := range m.schema.Enum {
			enum = append(enum, value)
			enumNames = append(enumNames, nameOf(m.enumNames, i, value))
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
			if m.enumNames != nil {
				enumNames = append(enumNames, m.enumNames[i])
			} else {
				enumNames = append(enumNames, nameOf(names, j, value))
			}
		}
	}
	if len(enum) == 0 {
		return mergeConflict(m.from["enum"], fmt.Sprintf("enum %v", m.schema.Enum), label,
			fmt.Sprintf("enum %v", v.Enum), "no value is in both. To allow the values of either, "+
				"set x-oapi-codegen-enum-merge: union on the composition")
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

// sameSchemaV3 reports whether two schema positions describe the same schema:
// the same $ref, or inline schemas with the same content. kin-openapi's
// source-location metadata is not part of the JSON encoding, so two identical
// schemas declared in different places compare equal.
func sameSchemaV3(r1, r2 *openapi3.SchemaRef) bool {
	if r1.Ref != "" || r2.Ref != "" {
		return r1.Ref == r2.Ref
	}
	b1, err1 := json.Marshal(r1.Value)
	b2, err2 := json.Marshal(r2.Value)
	return err1 == nil && err2 == nil && bytes.Equal(b1, b2)
}

// hasStructuralSiblingsV3 reports whether a schema with allOf also carries
// fields outside allOf that materially affect the generated Go type.
// Such fields must be merged with the allOf members rather than discarded.
//
// Description and Title are excluded — they are metadata, not structural,
// and the caller propagates them separately. Nullable/ReadOnly/WriteOnly
// are also excluded: merging would turn a pure wrapper such as
// {allOf: [$ref X], nullable: true}, which aliases X, into a copy of X.
// `type` is excluded for the same reason: {type: object, allOf: [$ref X]}
// must stay an alias of X.
//
// A oneOf or anyOf next to allOf adds a union to the merged type, unless its
// branches only add constraints (see isConstraintOnlyUnionV3).
func hasStructuralSiblingsV3(s *openapi3.Schema) bool {
	if s == nil {
		return false
	}
	return len(s.Properties) > 0 ||
		len(s.Required) > 0 ||
		s.AdditionalProperties.Has != nil ||
		s.AdditionalProperties.Schema != nil ||
		(len(s.OneOf) > 0 && !isConstraintOnlyUnionV3(s.OneOf)) ||
		(len(s.AnyOf) > 0 && !isConstraintOnlyUnionV3(s.AnyOf))
}

// isConstraintOnlyUnionV3 reports whether every branch of a oneOf/anyOf only
// adds constraints, such as `oneOf: [{required: [email]}, {required: [phone]}]`,
// rather than naming alternative types. Such a list next to allOf used to be
// dropped; it keeps being dropped, since generating a union of `any` branches
// would turn the merged type into something harder to use, not more correct.
// A `$ref` branch is judged by the schema it refers to.
func isConstraintOnlyUnionV3(branches openapi3.SchemaRefs) bool {
	for _, b := range branches {
		if b == nil || b.Value == nil {
			return false
		}
		v := b.Value
		if v.Type.Slice() != nil || len(v.Properties) > 0 || v.Items != nil ||
			len(v.AllOf) > 0 || len(v.AnyOf) > 0 || len(v.OneOf) > 0 ||
			len(v.Enum) > 0 || v.AdditionalProperties.Schema != nil {
			return false
		}
	}
	return true
}

// generateAllOfV3 lowers a schema with allOf for generateGoSchema: the
// composition merged into one Go type by merge, with the parent's structural
// siblings merged in, its description kept, and a recursive composition named.
func generateAllOfV3(ctx genContext, schema *openapi3.Schema, path []string, extensions map[string]any, skipOptionalPointer bool, merge schemaMerger) (Schema, error) {
	// An enclosing frame is already generating this composition. Refer to
	// the type it is building instead of inlining the body a second time,
	// which is what used to recurse until the stack ran out (issue #2542).
	if frame, ok := ctx.inProgress[schema]; ok {
		frame.consulted = true
		return Schema{
			GoType:              frame.typeName,
			RefType:             frame.typeName,
			DefineViaAlias:      true,
			SkipOptionalPointer: skipOptionalPointer,
			OAPISchema:          schema,
		}, nil
	}
	var err error
	ctx.unionEnums = false
	if raw, ok := extensions[extOapiCodegenEnumMerge]; ok {
		if ctx.unionEnums, err = extParseEnumMerge(raw); err != nil {
			return Schema{}, fmt.Errorf("invalid value for %q: %w", extOapiCodegenEnumMerge, err)
		}
	}
	frame := &mergeFrame{typeName: ctx.typeName(path)}
	ctx.inProgress[schema] = frame
	defer delete(ctx.inProgress, schema)

	// The parent's structural siblings are always merged: v3 rejects
	// old-allof-sibling-merging, which discarded them.
	var mergedSchema Schema
	if hasStructuralSiblingsV3(schema) {
		// Inject the parent (with AllOf cleared) as the final allOf
		// member so its structural siblings — Properties, Required,
		// AdditionalProperties — are merged with the allOf members
		// rather than discarded. Issues #697, #931, #1710, #2102.
		//
		// Allocate a fresh slice rather than appending to schema.AllOf
		// directly: if kin-openapi gave us a slice with spare capacity,
		// `append` would write the new element into the shared backing
		// array, mutating any other view that has been extended past
		// len(schema.AllOf).
		s := *schema
		s.AllOf = nil
		own := &openapi3.SchemaRef{Value: &s}
		ctx.memberLabels[own] = ""
		defer delete(ctx.memberLabels, own)
		// An opaque member can't be merged with the siblings either. Say
		// so here, where they can be named as the schema's own.
		for _, m := range schema.AllOf {
			if target := opaqueSchemaWithin(m, nil); target != nil {
				return Schema{}, fmt.Errorf("error merging schemas: %w", opaqueMergeError(m, target,
					"the schema's own "+strings.Join(typeKeywords(s), ", ")))
			}
		}
		allOfRefs := make([]*openapi3.SchemaRef, 0, len(schema.AllOf)+1)
		allOfRefs = append(allOfRefs, schema.AllOf...)
		allOfRefs = append(allOfRefs, own)
		mergedSchema, err = merge(ctx, allOfRefs, path)
	} else {
		// The parent is a pure wrapper with no structural siblings.
		// mergeSchemasV3's single-element fast path returns the
		// referenced type unchanged, preserving named-type identity.
		mergedSchema, err = merge(ctx, schema.AllOf, path)
	}
	if err != nil {
		return Schema{}, fmt.Errorf("error merging schemas: %w", err)
	}
	mergedSchema.OAPISchema = schema
	// Description is metadata, not a structural constraint, so it
	// doesn't go through the merge. Copy it from the parent when set.
	// Issue #1960.
	if schema.Description != "" {
		mergedSchema.Description = schema.Description
	}
	// x-go-type on the parent is handled by the early return above
	// (combined extensions). For x-go-type-skip-optional-pointer, only
	// override the merged value when the parent sets it explicitly —
	// otherwise we would clobber the value MergeSchemas computed from
	// the decorator idiom (an inline allOf member that carries the
	// extension; see mergeSchemasV3 and issue #1957).
	if _, ok := extensions[extPropGoTypeSkipOptionalPointer]; ok {
		mergedSchema.SkipOptionalPointer = skipOptionalPointer
	}
	// Something underneath referred back to this composition, so it has
	// to resolve to a named type. When nothing did — the overwhelmingly
	// common case — fall through with the anonymous struct this has
	// always produced, byte for byte.
	if frame.consulted {
		switch {
		case mergedSchema.RefType == frame.typeName:
			// Already defined under the promised name: generating the
			// merged body hoisted it (generate-types-for-anonymous-schemas).
		case mergedSchema.RefType != "":
			// The name handed to the recursive members is not the one the
			// type ended up with, so those references would dangle. Fail
			// loudly rather than emit code that does not compile.
			return Schema{}, fmt.Errorf(
				"recursive allOf composition at %s was generated as %q but its self-references were resolved to %q",
				strings.Join(ctx.nameHint, "."), mergedSchema.RefType, frame.typeName)
		case ctx.rootPosition:
			// GenerateTypesForSchemas names this one, from renameSchema
			// rather than from the path, so the name handed to the
			// members above is not the one it will be defined under.
			// Believed unreachable (see genContext.rootPosition); say so
			// rather than emit code that will not compile.
			return Schema{}, fmt.Errorf(
				"recursive allOf composition at the root of %s is not supported: give the composition its own schema",
				strings.Join(ctx.nameHint, "."))
		default:
			typeDef := TypeDefinition{
				TypeName: frame.typeName,
				JsonName: strings.Join(ctx.nameHint, "."),
				Schema:   mergedSchema,
			}
			mergedSchema.AdditionalTypes = append(mergedSchema.AdditionalTypes, typeDef)
			mergedSchema.RefType = frame.typeName
		}
	}
	return mergedSchema, nil
}
