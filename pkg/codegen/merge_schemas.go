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
		schema, err = mergeOpenapiSchemas(schema, oneOfSchema)
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
	if len(pathParts) != 2 {
		return openapi3.Schema{}, fmt.Errorf("unsupported reference: %s", ref.Ref)
	}
	remoteComponent, _ := pathParts[0], pathParts[1]

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
		schema, err = mergeOpenapiSchemas(schema, *schemaRef.Value)
		if err != nil {
			return openapi3.Schema{}, fmt.Errorf("error merging schemas for AllOf: %w", err)
		}
	}
	return schema, nil
}

// mergeOpenapiSchemas merges two openAPI schemas and returns the schema
// all of whose fields are composed.
func mergeOpenapiSchemas(s1, s2 openapi3.Schema) (openapi3.Schema, error) {
	var result openapi3.Schema
	if s1.Extensions != nil || s2.Extensions != nil {
		result.Extensions = make(map[string]interface{})
		if s1.Extensions != nil {
			for k, v := range s1.Extensions {
				result.Extensions[k] = v
			}
		}
		if s2.Extensions != nil {
			for k, v := range s2.Extensions {
				// TODO: Check for collisions
				result.Extensions[k] = v
			}
		}
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

	if s1.Type != "" && s2.Type != "" && s1.Type != s2.Type {
		return openapi3.Schema{}, errors.New("can not merge incompatible types")
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

	if SchemaHasAdditionalProperties(&s1) && SchemaHasAdditionalProperties(&s2) {
		return openapi3.Schema{}, errors.New("merging two schemas with additional properties, this is unhandled")
	}
	if s1.AdditionalProperties != nil {
		result.AdditionalProperties = s1.AdditionalProperties
	}
	if s2.AdditionalProperties != nil {
		result.AdditionalProperties = s2.AdditionalProperties
	}

	// Unhandled for now
	if s1.Discriminator != nil || s2.Discriminator != nil {
		return openapi3.Schema{}, errors.New("merging two schemas with discriminators is not supported")
	}

	return result, nil
}
