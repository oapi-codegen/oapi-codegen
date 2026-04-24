package codegen

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// MergeSchemas merges all the fields in the schemas supplied into one giant schema.
// The idea is that we merge all fields together into one schema.
func MergeSchemas(allOf []*openapi3.SchemaRef, path []string) (Schema, error) {
	// If someone asked for the old way, for backward compatibility, return the
	// old style result.
	if globalState.options.Compatibility.OldMergeSchemas {
		return mergeSchemasV1(allOf, path)
	}
	return mergeSchemas(allOf, path)
}

func mergeSchemas(allOf []*openapi3.SchemaRef, path []string) (Schema, error) {
	n := len(allOf)

	if n == 1 {
		return GenerateGoSchema(allOf[0], path)
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
		if m.Ref == "" && isExtensionOnlySchema(m.Value) {
			decoratorIdiom = true
			break
		}
	}

	schema, err := valueWithPropagatedRef(allOf[0])
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
		var err error
		oneOfSchema, err := valueWithPropagatedRef(allOf[i])
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
		schema, err = mergeOpenapiSchemas(schema, oneOfSchema, true, seenSchemaRef)
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
		// reallocates schema.Extensions in mergeOpenapiSchemas before
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

	return GenerateGoSchema(openapi3.NewSchemaRef("", &schema), path)
}

// isExtensionOnlySchema reports whether a schema carries only extensions,
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
func isExtensionOnlySchema(s *openapi3.Schema) bool {
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

// valueWithPropagatedRef returns a copy of ref's schema with its Properties
// refs rewritten when ref itself is external, and with extensions placed
// next to the $ref folded in (ref-side wins over value-side). This is what
// allows allOf members to carry per-use sibling directives without
// mutating the referenced schema.
func valueWithPropagatedRef(ref *openapi3.SchemaRef) (openapi3.Schema, error) {
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

	propagateRemoteRefs(remoteComponent, &schema)

	return schema, nil
}

// propagateRemoteRefs rewrites local "#/..." refs within a schema to be
// qualified with the remote component path. This is needed so that when an
// external schema is flattened via allOf, nested type references (array items,
// additionalProperties, sub-object properties) retain their external
// qualification. See https://github.com/oapi-codegen/oapi-codegen/issues/2288
func propagateRemoteRefs(remoteComponent string, schema *openapi3.Schema) {
	for _, value := range schema.Properties {
		if len(value.Ref) > 0 && value.Ref[0] == '#' {
			value.Ref = remoteComponent + value.Ref
		} else if value.Value != nil {
			propagateRemoteRefs(remoteComponent, value.Value)
		}
	}

	if schema.Items != nil {
		if len(schema.Items.Ref) > 0 && schema.Items.Ref[0] == '#' {
			schema.Items.Ref = remoteComponent + schema.Items.Ref
		} else if schema.Items.Value != nil {
			propagateRemoteRefs(remoteComponent, schema.Items.Value)
		}
	}

	if schema.AdditionalProperties.Schema != nil {
		ap := schema.AdditionalProperties.Schema
		if len(ap.Ref) > 0 && ap.Ref[0] == '#' {
			ap.Ref = remoteComponent + ap.Ref
		} else if ap.Value != nil {
			propagateRemoteRefs(remoteComponent, ap.Value)
		}
	}

	for _, list := range [][]*openapi3.SchemaRef{schema.AllOf, schema.AnyOf, schema.OneOf} {
		for _, ref := range list {
			if len(ref.Ref) > 0 && ref.Ref[0] == '#' {
				ref.Ref = remoteComponent + ref.Ref
			} else if ref.Value != nil {
				propagateRemoteRefs(remoteComponent, ref.Value)
			}
		}
	}

	if schema.Not != nil {
		if len(schema.Not.Ref) > 0 && schema.Not.Ref[0] == '#' {
			schema.Not.Ref = remoteComponent + schema.Not.Ref
		} else if schema.Not.Value != nil {
			propagateRemoteRefs(remoteComponent, schema.Not.Value)
		}
	}
}

func mergeAllOf(allOf []*openapi3.SchemaRef, seenSchemaRef map[string]bool) (openapi3.Schema, error) {
	var schema openapi3.Schema
	for _, schemaRef := range allOf {
		var err error
		if schemaRef.Ref != "" && seenSchemaRef[schemaRef.Ref] {
			continue
		}
		if schemaRef.Ref != "" {
			seenSchemaRef[schemaRef.Ref] = true
		}
		// Use valueWithPropagatedRef so sibling extensions on a $ref
		// member of a transitively-flattened allOf reach the merged
		// schema, matching mergeSchemas' top-level handling.
		member, err := valueWithPropagatedRef(schemaRef)
		if err != nil {
			return openapi3.Schema{}, err
		}
		schema, err = mergeOpenapiSchemas(schema, member, true, seenSchemaRef)
		if err != nil {
			return openapi3.Schema{}, fmt.Errorf("error merging schemas for AllOf: %w", err)
		}
	}
	return schema, nil
}

// mergeOpenapiSchemas merges two openAPI schemas and returns the schema
// all of whose fields are composed.
func mergeOpenapiSchemas(s1, s2 openapi3.Schema, allOf bool, seenSchemaRef map[string]bool) (openapi3.Schema, error) {
	var result openapi3.Schema

	result.Extensions = make(map[string]any, len(s1.Extensions)+len(s2.Extensions))
	maps.Copy(result.Extensions, s1.Extensions)
	// TODO: Check for collisions
	maps.Copy(result.Extensions, s2.Extensions)

	result.OneOf = append(s1.OneOf, s2.OneOf...)

	// We are going to make AllOf transitive, so that merging an AllOf that
	// contains AllOf's will result in a flat object.
	var err error
	if s1.AllOf != nil {
		var merged openapi3.Schema
		merged, err = mergeAllOf(s1.AllOf, seenSchemaRef)
		if err != nil {
			return openapi3.Schema{}, fmt.Errorf("error transitive merging AllOf on schema 1")
		}
		s1 = merged
	}
	if s2.AllOf != nil {
		var merged openapi3.Schema
		merged, err = mergeAllOf(s2.AllOf, seenSchemaRef)
		if err != nil {
			return openapi3.Schema{}, fmt.Errorf("error transitive merging AllOf on schema 2")
		}
		s2 = merged
	}

	result.AllOf = append(s1.AllOf, s2.AllOf...)

	if s1.Type.Slice() != nil && s2.Type.Slice() != nil && !equalTypes(s1.Type, s2.Type) {
		return openapi3.Schema{}, fmt.Errorf("can not merge incompatible types: %v, %v", s1.Type.Slice(), s2.Type.Slice())
	}
	result.Type = s1.Type

	if s1.Format != s2.Format {
		return openapi3.Schema{}, errors.New("can not merge incompatible formats")
	}
	result.Format = s1.Format

	// For Enums, do we union, or intersect? This is a bit vague. I choose
	// to be more permissive and union.
	result.Enum = append(s1.Enum, s2.Enum...)

	// I don't know how to handle two different defaults.
	if s1.Default != nil || s2.Default != nil {
		return openapi3.Schema{}, errors.New("merging two sets of defaults is undefined")
	}
	if s1.Default != nil {
		result.Default = s1.Default
	}
	if s2.Default != nil {
		result.Default = s2.Default
	}

	// We skip Example
	// We skip ExternalDocs

	// If two schemas disagree on any of these flags, we error out.
	if s1.UniqueItems != s2.UniqueItems {
		return openapi3.Schema{}, errors.New("merging two schemas with different UniqueItems")

	}
	result.UniqueItems = s1.UniqueItems

	if !reflect.DeepEqual(s1.ExclusiveMin, s2.ExclusiveMin) {
		return openapi3.Schema{}, errors.New("merging two schemas with different ExclusiveMin")

	}
	result.ExclusiveMin = s1.ExclusiveMin

	if !reflect.DeepEqual(s1.ExclusiveMax, s2.ExclusiveMax) {
		return openapi3.Schema{}, errors.New("merging two schemas with different ExclusiveMax")

	}
	result.ExclusiveMax = s1.ExclusiveMax

	if s1.Nullable != s2.Nullable {
		return openapi3.Schema{}, errors.New("merging two schemas with different Nullable")

	}
	result.Nullable = s1.Nullable

	if s1.ReadOnly != s2.ReadOnly {
		return openapi3.Schema{}, errors.New("merging two schemas with different ReadOnly")

	}
	result.ReadOnly = s1.ReadOnly

	if s1.WriteOnly != s2.WriteOnly {
		return openapi3.Schema{}, errors.New("merging two schemas with different WriteOnly")

	}
	result.WriteOnly = s1.WriteOnly

	if s1.AllowEmptyValue != s2.AllowEmptyValue {
		return openapi3.Schema{}, errors.New("merging two schemas with different AllowEmptyValue")

	}
	result.AllowEmptyValue = s1.AllowEmptyValue

	// Required. We merge these.
	result.Required = append(s1.Required, s2.Required...)

	// We merge all properties
	result.Properties = make(map[string]*openapi3.SchemaRef, len(s1.Properties)+len(s2.Properties))
	maps.Copy(result.Properties, s1.Properties)
	// TODO: detect conflicts
	maps.Copy(result.Properties, s2.Properties)

	if isAdditionalPropertiesExplicitFalse(&s1) || isAdditionalPropertiesExplicitFalse(&s2) {
		result.WithoutAdditionalProperties()
	} else if s1.AdditionalProperties.Schema != nil {
		if s2.AdditionalProperties.Schema != nil {
			return openapi3.Schema{}, errors.New("merging two schemas with additional properties, this is unhandled")
		} else {
			result.AdditionalProperties.Schema = s1.AdditionalProperties.Schema
		}
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

func equalTypes(t1 *openapi3.Types, t2 *openapi3.Types) bool {
	s1 := t1.Slice()
	s2 := t2.Slice()

	if len(s1) != len(s2) {
		return false
	}

	// NOTE that ideally we'd use `slices.Equal` but as we're currently supporting Go 1.20+, we can't use it (yet https://github.com/oapi-codegen/oapi-codegen/issues/1634)
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}

	return true
}
