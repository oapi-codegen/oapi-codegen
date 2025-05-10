package codegen

import (
	"errors"
	"fmt"
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

	schema, err := valueWithPropagatedRef(allOf[0])
	if err != nil {
		return Schema{}, err
	}

	for i := 1; i < n; i++ {
		var err error
		oneOfSchema, err := valueWithPropagatedRef(allOf[i])
		if err != nil {
			return Schema{}, err
		}
		schema, err = mergeOpenapiSchemas(schema, oneOfSchema, true)
		if err != nil {
			return Schema{}, fmt.Errorf("error merging schemas for AllOf: %w", err)
		}
	}
	return GenerateGoSchema(openapi3.NewSchemaRef("", &schema), path)
}

// valueWithPropagatedRef returns a copy of ref schema with its Properties refs
// updated if ref itself is external. Otherwise, return ref.Value as-is.
func valueWithPropagatedRef(ref *openapi3.SchemaRef) (openapi3.Schema, error) {
	if len(ref.Ref) == 0 || ref.Ref[0] == '#' {
		return *ref.Value, nil
	}

	pathParts := strings.Split(ref.Ref, "#")
	if len(pathParts) < 1 || len(pathParts) > 2 {
		return openapi3.Schema{}, fmt.Errorf("unsupported reference: %s", ref.Ref)
	}
	remoteComponent := pathParts[0]

	// remote ref
	schema := *ref.Value
	for _, value := range schema.Properties {
		if len(value.Ref) > 0 && value.Ref[0] == '#' {
			// local reference, should propagate remote
			value.Ref = remoteComponent + value.Ref
		}
	}

	return schema, nil
}

func mergeAllOf(allOf []*openapi3.SchemaRef) (openapi3.Schema, error) {
	var schema openapi3.Schema
	for _, schemaRef := range allOf {
		var err error
		schema, err = mergeOpenapiSchemas(schema, *schemaRef.Value, true)
		if err != nil {
			return openapi3.Schema{}, fmt.Errorf("error merging schemas for AllOf: %w", err)
		}
	}
	return schema, nil
}

// mergeOpenapiSchemas merges two openAPI schemas and returns the schema
// all of whose fields are composed.
func mergeOpenapiSchemas(s1, s2 openapi3.Schema, allOf bool) (openapi3.Schema, error) {
	var result openapi3.Schema

	result.Extensions = make(map[string]interface{})
	for k, v := range s1.Extensions {
		result.Extensions[k] = v
	}
	for k, v := range s2.Extensions {
		// TODO: Check for collisions
		result.Extensions[k] = v
	}

	result.OneOf = append(s1.OneOf, s2.OneOf...)

	// We are going to make AllOf transitive, so that merging an AllOf that
	// contains AllOf's will result in a flat object.
	var err error
	if s1.AllOf != nil {
		var merged openapi3.Schema
		merged, err = mergeAllOf(s1.AllOf)
		if err != nil {
			return openapi3.Schema{}, fmt.Errorf("error transitive merging AllOf on schema 1")
		}
		s1 = merged
	}
	if s2.AllOf != nil {
		var merged openapi3.Schema
		merged, err = mergeAllOf(s2.AllOf)
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

	if s1.ExclusiveMin != s2.ExclusiveMin {
		return openapi3.Schema{}, errors.New("merging two schemas with different ExclusiveMin")

	}
	result.ExclusiveMin = s1.ExclusiveMin

	if s1.ExclusiveMax != s2.ExclusiveMax {
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
	result.Properties = make(map[string]*openapi3.SchemaRef)
	for k, v := range s1.Properties {
		result.Properties[k] = v
	}
	for k, v := range s2.Properties {
		// TODO: detect conflicts
		result.Properties[k] = v
	}

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
