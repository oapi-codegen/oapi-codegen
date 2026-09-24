package codegen

// This file holds the allOf half of schema-merging-behavior v2: allOf members
// merged keyword by keyword into one Go type. v2 has been the default since
// oapi-codegen v1.11.0. Its anyOf/oneOf half is in union_v2.go. v1 uses
// generateAllOfV2 too, with its own merge of the members.
//
// v2 is kept for compatibility. Bug fixes are welcome, but a change here
// must not cause a regression: a spec that generated working code must keep
// generating it. New behavior goes into a new version, in files of its own.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

func mergeSchemasV2(ctx genContext, allOf []*openapi3.SchemaRef, path []string) (Schema, error) {
	n := len(allOf)

	if n == 1 {
		return generateGoSchema(ctx, allOf[0], path)
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
	//      result is a NEW distinct type, and extensions like x-go-type on
	//      a source schema do NOT transfer (see issue #2335: Client has
	//      x-go-type=OverlayClient, but allOf[Client, {properties:{id}}]
	//      is ClientWithId — a different shape, not OverlayClient).
	//
	// A $ref member is excluded from the decorator check because it is by
	// construction delivering the referenced schema, not "decorating"
	// siblings — even if the referenced schema happens to carry only
	// extensions, that's a property of the target, not an intent on this
	// composition.
	decoratorIdiom := false
	for _, m := range allOf {
		if m.Ref == "" && isExtensionOnlySchemaV2(m.Value) {
			decoratorIdiom = true
			break
		}
	}

	schema, err := valueWithPropagatedRefV2(allOf[0])
	if err != nil {
		return Schema{}, err
	}

	// Seed allOf[0]'s ref so that if s1's own AllOf contains a back-reference
	// to itself, the cycle is detected during recursive merging.
	seenTopLevel := make(map[string]bool)
	if allOf[0].Ref != "" {
		seenTopLevel[allOf[0].Ref] = true
	}

	for i := 1; i < n; i++ {
		oneOfSchema, err := valueWithPropagatedRefV2(allOf[i])
		if err != nil {
			return Schema{}, err
		}

		seenSchemaRef := make(map[string]bool)
		for k := range seenTopLevel {
			seenSchemaRef[k] = true
		}
		if allOf[i].Ref != "" {
			seenSchemaRef[allOf[i].Ref] = true
			seenTopLevel[allOf[i].Ref] = true
		}
		schema, err = mergeOpenapiSchemasV2(schema, oneOfSchema, true, seenSchemaRef)
		if err != nil {
			return Schema{}, fmt.Errorf("error merging schemas for AllOf: %w", err)
		}
	}

	if !decoratorIdiom {
		// Drop only the type-identity directives. Other extensions
		// (user-defined x-* metadata, etc.) are preserved — we only
		// have concrete evidence that the identity-bound ones cause
		// incorrect aliasing across composition.
		//
		// Clone before mutating: the current merge path always
		// reallocates schema.Extensions in mergeOpenapiSchemasV2 before
		// we reach here, so the delete is safe today — but the
		// defensive copy keeps this correct if that invariant changes
		// (e.g. an allocation-skipping optimization). Cost is a small
		// map copy on a single code path.
		ext := maps.Clone(schema.Extensions)
		delete(ext, extPropGoType)
		delete(ext, extGoTypeName)
		delete(ext, extPropGoImport)
		schema.Extensions = ext
	}

	return generateGoSchema(ctx, openapi3.NewSchemaRef("", &schema), path)
}

// isExtensionOnlySchemaV2 reports whether a schema carries only extensions,
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
func isExtensionOnlySchemaV2(s *openapi3.Schema) bool {
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

// valueWithPropagatedRefV2 returns a copy of ref's schema with its Properties
// refs rewritten when ref itself is external, and with extensions placed
// next to the $ref folded in (ref-side wins over value-side). This is what
// allows allOf members to carry per-use sibling directives without
// mutating the referenced schema.
func valueWithPropagatedRefV2(ref *openapi3.SchemaRef) (openapi3.Schema, error) {
	schema := *ref.Value
	schema.Extensions = combinedSchemaExtensions(ref)

	if len(ref.Ref) == 0 || ref.Ref[0] == '#' {
		return schema, nil
	}

	pathParts := strings.Split(ref.Ref, "#")
	if len(pathParts) < 1 || len(pathParts) > 2 {
		return openapi3.Schema{}, fmt.Errorf("unsupported reference: %s", ref.Ref)
	}
	remoteComponent := pathParts[0]

	propagateRemoteRefsV2(remoteComponent, &schema)

	return schema, nil
}

// propagateRemoteRefsV2 rewrites local "#/..." refs within a schema to be
// qualified with the remote component path. This is needed so that when an
// external schema is flattened via allOf, nested type references (array items,
// additionalProperties, sub-object properties) retain their external
// qualification. See https://github.com/oapi-codegen/oapi-codegen/issues/2288
func propagateRemoteRefsV2(remoteComponent string, schema *openapi3.Schema) {
	for _, value := range schema.Properties {
		qualifyRemoteRefV2(remoteComponent, value)
	}
	qualifyRemoteRefV2(remoteComponent, schema.Items)
	qualifyRemoteRefV2(remoteComponent, schema.AdditionalProperties.Schema)
	for _, list := range [][]*openapi3.SchemaRef{schema.AllOf, schema.AnyOf, schema.OneOf} {
		for _, ref := range list {
			qualifyRemoteRefV2(remoteComponent, ref)
		}
	}
	qualifyRemoteRefV2(remoteComponent, schema.Not)
}

// qualifyRemoteRefV2 qualifies one position inside a schema being flattened out
// of a remote document: a local "#/..." ref is rewritten to point at that
// document, and an inline schema is walked for positions of its own.
//
// A $ref is never followed. The schema it names becomes a Go type generated
// from the document it lives in, so its body is not ours to rewrite — and a
// ref already qualified for some document must not be re-qualified for this
// one. Not following refs is also what makes this terminate: a schema can
// only refer back to itself through a $ref, so walking inline schemas alone
// cannot cycle. Following them recursed forever the second time a
// self-recursive remote schema was flattened, because the first pass had
// rewritten the very refs whose "#" prefix stopped the walk
// (https://github.com/oapi-codegen/oapi-codegen/issues/2557).
func qualifyRemoteRefV2(remoteComponent string, ref *openapi3.SchemaRef) {
	if ref == nil {
		return
	}
	if len(ref.Ref) > 0 {
		if ref.Ref[0] == '#' {
			ref.Ref = remoteComponent + ref.Ref
		}
		return
	}
	if ref.Value != nil {
		propagateRemoteRefsV2(remoteComponent, ref.Value)
	}
}

func mergeAllOfV2(allOf []*openapi3.SchemaRef, seenSchemaRef map[string]bool) (openapi3.Schema, error) {
	var schema openapi3.Schema
	for _, schemaRef := range allOf {
		if schemaRef.Ref != "" && seenSchemaRef[schemaRef.Ref] {
			continue
		}
		if schemaRef.Ref != "" {
			seenSchemaRef[schemaRef.Ref] = true
		}
		// Use valueWithPropagatedRefV2 so sibling extensions on a $ref
		// member of a transitively-flattened allOf reach the merged
		// schema, matching mergeSchemasV2' top-level handling.
		member, err := valueWithPropagatedRefV2(schemaRef)
		if err != nil {
			return openapi3.Schema{}, err
		}
		schema, err = mergeOpenapiSchemasV2(schema, member, true, seenSchemaRef)
		if err != nil {
			return openapi3.Schema{}, fmt.Errorf("error merging schemas for AllOf: %w", err)
		}
	}
	return schema, nil
}

// mergeOpenapiSchemasV2 merges two openAPI schemas and returns the schema
// all of whose fields are composed.
func mergeOpenapiSchemasV2(s1, s2 openapi3.Schema, allOf bool, seenSchemaRef map[string]bool) (openapi3.Schema, error) {
	var result openapi3.Schema

	result.Extensions = make(map[string]any, len(s1.Extensions)+len(s2.Extensions))
	maps.Copy(result.Extensions, s1.Extensions)
	// TODO: Check for collisions
	maps.Copy(result.Extensions, s2.Extensions)

	// Capture top-level OneOf/AnyOf before overwriting s1/s2 with transitive
	// AllOf merges. The merges may surface additional OneOf/AnyOf members from
	// nested allOf members (issue #1905), so we accumulate from both sources.
	oneOf := append(s1.OneOf, s2.OneOf...)
	anyOf := append(s1.AnyOf, s2.AnyOf...)

	// We are going to make AllOf transitive, so that merging an AllOf that
	// contains AllOf's will result in a flat object.
	var err error
	if s1.AllOf != nil {
		var merged openapi3.Schema
		merged, err = mergeAllOfV2(s1.AllOf, seenSchemaRef)
		if err != nil {
			return openapi3.Schema{}, fmt.Errorf("error transitive merging AllOf on schema 1")
		}
		oneOf = append(oneOf, merged.OneOf...)
		anyOf = append(anyOf, merged.AnyOf...)
		s1 = merged
	}
	if s2.AllOf != nil {
		var merged openapi3.Schema
		merged, err = mergeAllOfV2(s2.AllOf, seenSchemaRef)
		if err != nil {
			return openapi3.Schema{}, fmt.Errorf("error transitive merging AllOf on schema 2")
		}
		oneOf = append(oneOf, merged.OneOf...)
		anyOf = append(anyOf, merged.AnyOf...)
		s2 = merged
	}

	result.OneOf = oneOf
	result.AnyOf = anyOf
	result.AllOf = append(s1.AllOf, s2.AllOf...)

	// Type conflicts are an error only when both members declare a type.
	// When exactly one declares one, it propagates: a typeless member
	// contributes its other constraints (properties, required, ...) without
	// erasing the sibling's type. Taking s1.Type unconditionally here used
	// to silently drop s2's type, making the generated shape depend on
	// allOf member order (issue #2524).
	//
	// "null" is left out of the comparison. In 3.1 it is how a type array
	// spells nullability, which is unioned below rather than required to
	// match, the same as 3.0's `nullable`; so `[object]` and
	// `[object, "null"]` merge into a nullable object instead of failing.
	// The remaining types compare as sets, so their order doesn't matter.
	//
	// The union is deliberate, not an oversight of allOf's intersection
	// semantics. Read as an intersection, "null" in one member would mean
	// nothing unless every member declared it, yet a member only says it to
	// make the composed type nullable, as in the 3.0 idiom
	// `allOf: [$ref X, {nullable: true}]` (issue #1898). Both spec versions
	// give it that meaning.
	t1, t2 := nonNullTypes(s1.Type), nonNullTypes(s2.Type)
	if len(t1) > 0 && len(t2) > 0 && !sameTypeSetV2(t1, t2) {
		return openapi3.Schema{}, fmt.Errorf("can not merge incompatible types: %v, %v", s1.Type.Slice(), s2.Type.Slice())
	}
	result.Type = mergeTypesV2(s1.Type, s2.Type)

	// Format follows the same rule: error only when both members declare
	// a format and they differ. Erroring on the set-vs-unset case made the
	// allOf decorator idiom (e.g. $ref + nullable, issue #1898) fail for
	// refs to format-carrying scalars.
	if s1.Format != "" && s2.Format != "" && s1.Format != s2.Format {
		return openapi3.Schema{}, errors.New("can not merge incompatible formats")
	}
	result.Format = s1.Format
	if result.Format == "" {
		result.Format = s2.Format
	}

	// For Enums, do we union, or intersect? This is a bit vague. I choose
	// to be more permissive and union.
	result.Enum = append(s1.Enum, s2.Enum...)

	// Defaults are annotations: they don't affect the generated Go type, so
	// they can't conflict. Keep the later member's. This used to fail when
	// *either* member had a default, which rejected the everyday
	// `allOf: [$ref EnumWithDefault, {description: ...}]` decorator (issue
	// #1379).
	result.Default = s1.Default
	if s2.Default != nil {
		result.Default = s2.Default
	}

	// We skip Example
	// We skip ExternalDocs

	// uniqueItems and the exclusive bounds are validation constraints that
	// don't shape the Go type. Disagreeing on them used to be an error,
	// which rejected decorators such as `allOf: [$ref UniqueTags,
	// {description: ...}]`: kin-openapi can't tell an unset flag from an
	// explicit false. allOf requires every member's constraints, so a flag
	// set by either member is set on the result; a bound set by only one
	// member carries over.
	result.UniqueItems = s1.UniqueItems || s2.UniqueItems

	result.ExclusiveMin = s1.ExclusiveMin
	if !result.ExclusiveMin.IsSet() {
		result.ExclusiveMin = s2.ExclusiveMin
	}

	result.ExclusiveMax = s1.ExclusiveMax
	if !result.ExclusiveMax.IsSet() {
		result.ExclusiveMax = s2.ExclusiveMax
	}

	// Compare nullability via schemaIsNullable so this works the same way
	// regardless of spec version: in 3.0 it reads s.Nullable, in 3.1 it
	// reads "null" from the type array. Type merging itself is NOT version
	// branched -- the type check at result.Type assignment above ignores
	// "null" and mergeTypesV2 adds it back when either member has it:
	//
	//   3.0: ["string"] vs ["string"]                 -> ["string"]
	//   3.1: ["string","null"] vs ["string","null"]   -> ["string","null"]
	//   3.1: ["string","null"] vs ["string"]          -> ["string","null"]
	//
	// Because result.Type already carries a "null" entry when either
	// member had one, the merged result is correctly nullable in 3.1
	// without needing to touch result.Nullable. The result.Nullable copy
	// below is a no-op in 3.1 (s1.Nullable is always false there) but kept
	// for 3.0 correctness, where Nullable is the only nullability carrier.
	//
	// Nullability is UNIONed rather than required to match: if any member
	// is nullable, the merged schema is nullable. This supports the common
	// OpenAPI 3.0 idiom of decorating a $ref with nullability, which is only
	// expressible through allOf because 3.0 forbids siblings next to $ref
	// (issue #1898):
	//
	//   allOf:
	//     - $ref: "#/components/schemas/user"
	//     - nullable: true
	//
	// kin-openapi represents Nullable as a plain bool, so an unset value is
	// indistinguishable from an explicit `nullable: false`; erroring on a
	// mismatch made this widely-used idiom unusable. Union is also
	// consistent with how Required is merged below.
	if schemaIsNullable(&s1) || schemaIsNullable(&s2) {
		result.Nullable = true
	}

	// readOnly, writeOnly and allowEmptyValue are ORed for the same reason:
	// requiring them to match rejected `allOf: [$ref X, {readOnly: true}]`,
	// the only way OpenAPI 3.0 can mark a $ref read-only.
	result.ReadOnly = s1.ReadOnly || s2.ReadOnly
	result.WriteOnly = s1.WriteOnly || s2.WriteOnly
	result.AllowEmptyValue = s1.AllowEmptyValue || s2.AllowEmptyValue

	// Required. We merge these.
	result.Required = append(s1.Required, s2.Required...)

	// Items used to be dropped, so an allOf over an array schema, such as
	// `allOf: [$ref ArrayOfX, {minItems: 1}]`, generated []any.
	result.Items, err = mergeItemsV2(s1.Items, s2.Items, seenSchemaRef)
	if err != nil {
		return openapi3.Schema{}, err
	}

	// We merge all properties
	result.Properties = make(map[string]*openapi3.SchemaRef, len(s1.Properties)+len(s2.Properties))
	maps.Copy(result.Properties, s1.Properties)
	// TODO: detect conflicts
	maps.Copy(result.Properties, s2.Properties)

	if isAdditionalPropertiesExplicitFalse(&s1) || isAdditionalPropertiesExplicitFalse(&s2) {
		result.WithoutAdditionalProperties()
	} else if s1.AdditionalProperties.Schema != nil {
		// Two additionalProperties schemas merge only when they are the same
		// schema, e.g. two members that each allow extra string values.
		if s2.AdditionalProperties.Schema != nil && !sameSchemaV2(s1.AdditionalProperties.Schema, s2.AdditionalProperties.Schema) {
			return openapi3.Schema{}, errors.New("merging two schemas with different additional properties, this is unhandled")
		}
		result.AdditionalProperties.Schema = s1.AdditionalProperties.Schema
	} else {
		if s2.AdditionalProperties.Schema != nil {
			result.AdditionalProperties.Schema = s2.AdditionalProperties.Schema
		} else {
			if s1.AdditionalProperties.Has != nil || s2.AdditionalProperties.Has != nil {
				result.WithAnyAdditionalProperties()
			}
		}
	}

	// Allow discriminators for allOf merges, but disallow for one/anyOfs.
	if !allOf && (s1.Discriminator != nil || s2.Discriminator != nil) {
		return openapi3.Schema{}, errors.New("merging two schemas with discriminators is not supported")
	}

	// For allOf merges, propagate a discriminator if only one schema has it.
	// Merging two different discriminators is not supported.
	if s1.Discriminator != nil && s2.Discriminator != nil {
		return openapi3.Schema{}, errors.New("merging two schemas with discriminators is not supported")
	}
	if s1.Discriminator != nil {
		result.Discriminator = s1.Discriminator
	} else if s2.Discriminator != nil {
		result.Discriminator = s2.Discriminator
	}

	return result, nil
}

// sameTypeSetV2 reports whether two type lists name the same types, in any
// order.
func sameTypeSetV2(t1, t2 []string) bool {
	if len(t1) != len(t2) {
		return false
	}
	for _, typ := range t1 {
		if !slices.Contains(t2, typ) {
			return false
		}
	}
	return true
}

// mergeTypesV2 returns the type of an allOf merge whose members' non-null types
// are already known to agree: whichever member declares types, plus "null"
// when either member's type array has it. The member's own Types value is
// returned whenever it already says that, so merges that worked before
// produce identical output.
func mergeTypesV2(t1, t2 *openapi3.Types) *openapi3.Types {
	base := t1
	if len(nonNullTypes(t1)) == 0 && t2.Slice() != nil {
		base = t2
	}
	if base.Slice() == nil {
		return base
	}
	needsNull := slices.Contains(t1.Slice(), openapi3.TypeNull) || slices.Contains(t2.Slice(), openapi3.TypeNull)
	if !needsNull || slices.Contains(base.Slice(), openapi3.TypeNull) {
		return base
	}
	merged := openapi3.Types(append(slices.Clone(base.Slice()), openapi3.TypeNull))
	return &merged
}

// sameSchemaV2 reports whether two schema positions describe the same schema:
// the same $ref, or inline schemas with the same content. kin-openapi's
// source-location metadata is not part of the JSON encoding, so two identical
// schemas declared in different places compare equal.
func sameSchemaV2(r1, r2 *openapi3.SchemaRef) bool {
	if r1.Ref != "" || r2.Ref != "" {
		return r1.Ref == r2.Ref
	}
	b1, err1 := json.Marshal(r1.Value)
	b2, err2 := json.Marshal(r2.Value)
	return err1 == nil && err2 == nil && bytes.Equal(b1, b2)
}

// mergeItemsV2 merges the array items of two allOf members. A one-sided items
// carries over, and two different item schemas are merged with the same rules
// as their parents.
func mergeItemsV2(i1, i2 *openapi3.SchemaRef, seenSchemaRef map[string]bool) (*openapi3.SchemaRef, error) {
	switch {
	case i1 == nil:
		return i2, nil
	case i2 == nil:
		return i1, nil
	case sameSchemaV2(i1, i2):
		return i1, nil
	case (i1.Ref != "" && seenSchemaRef[i1.Ref]) || (i2.Ref != "" && seenSchemaRef[i2.Ref]):
		// Merging these items would re-enter a schema this merge is already
		// inside. Keep the behavior from before items were merged at all
		// (dropping them) rather than recursing forever.
		return nil, nil
	}
	seen := maps.Clone(seenSchemaRef)
	for _, r := range []*openapi3.SchemaRef{i1, i2} {
		if r.Ref != "" {
			seen[r.Ref] = true
		}
	}
	v1, err := valueWithPropagatedRefV2(i1)
	if err != nil {
		return nil, err
	}
	v2, err := valueWithPropagatedRefV2(i2)
	if err != nil {
		return nil, err
	}
	merged, err := mergeOpenapiSchemasV2(v1, v2, true, seen)
	if err != nil {
		return nil, fmt.Errorf("error merging array items: %w", err)
	}
	return openapi3.NewSchemaRef("", &merged), nil
}

// hasStructuralSiblingsV2 reports whether a schema with allOf also carries
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
// branches only add constraints (see isConstraintOnlyUnionV2).
func hasStructuralSiblingsV2(s *openapi3.Schema) bool {
	if s == nil {
		return false
	}
	return len(s.Properties) > 0 ||
		len(s.Required) > 0 ||
		s.AdditionalProperties.Has != nil ||
		s.AdditionalProperties.Schema != nil ||
		(len(s.OneOf) > 0 && !isConstraintOnlyUnionV2(s.OneOf)) ||
		(len(s.AnyOf) > 0 && !isConstraintOnlyUnionV2(s.AnyOf))
}

// isConstraintOnlyUnionV2 reports whether every branch of a oneOf/anyOf only
// adds constraints, such as `oneOf: [{required: [email]}, {required: [phone]}]`,
// rather than naming alternative types. Such a list next to allOf used to be
// dropped; it keeps being dropped, since generating a union of `any` branches
// would turn the merged type into something harder to use, not more correct.
// A `$ref` branch is judged by the schema it refers to.
func isConstraintOnlyUnionV2(branches openapi3.SchemaRefs) bool {
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

// generateAllOfV2 lowers a schema with allOf for generateGoSchema: the
// composition merged into one Go type by merge, with the parent's structural
// siblings merged in, its description kept, and a recursive composition named.
func generateAllOfV2(ctx genContext, schema *openapi3.Schema, path []string, extensions map[string]any, skipOptionalPointer bool, merge schemaMerger) (Schema, error) {
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
	frame := &mergeFrame{typeName: ctx.typeName(path)}
	ctx.inProgress[schema] = frame
	defer delete(ctx.inProgress, schema)

	var mergedSchema Schema
	// Behavior is gated on Compatibility.OldAllOfSiblingMerging:
	// when set, the parent's structural siblings and Description are
	// silently discarded (the historical behavior). When unset
	// (default), they are merged into the result.
	mergeSiblings := !globalState.options.Compatibility.OldAllOfSiblingMerging
	if mergeSiblings && hasStructuralSiblingsV2(schema) {
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
		allOfRefs := make([]*openapi3.SchemaRef, 0, len(schema.AllOf)+1)
		allOfRefs = append(allOfRefs, schema.AllOf...)
		allOfRefs = append(allOfRefs, &openapi3.SchemaRef{Value: &s})
		mergedSchema, err = merge(ctx, allOfRefs, path)
	} else {
		// Either the user opted into legacy behavior, or the parent is
		// a pure wrapper with no structural siblings. In the wrapper
		// case under v2, mergeSchemasV2's single-element fast path
		// returns the referenced type unchanged, preserving named-type
		// identity; v1's merge has no such path.
		mergedSchema, err = merge(ctx, schema.AllOf, path)
	}
	if err != nil {
		return Schema{}, fmt.Errorf("error merging schemas: %w", err)
	}
	mergedSchema.OAPISchema = schema
	// Description is metadata, not a structural constraint, so it
	// doesn't go through the merge. Copy it from the parent when set.
	// Issue #1960. Gated on the same compatibility flag as the
	// sibling-merge above.
	if mergeSiblings && schema.Description != "" {
		mergedSchema.Description = schema.Description
	}
	// x-go-type on the parent is handled by the early return above
	// (combined extensions). For x-go-type-skip-optional-pointer, only
	// override the merged value when the parent sets it explicitly —
	// otherwise we would clobber the value MergeSchemas computed from
	// the decorator idiom (an inline allOf member that carries the
	// extension; see mergeSchemasV2 and issue #1957).
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
