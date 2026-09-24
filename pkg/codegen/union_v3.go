package codegen

// This file holds the anyOf/oneOf half of schema-merging-behavior v3, the
// version under development. Its allOf half is in merge_schemas_v3.go. It
// started as a copy of v2's (union_v2.go).

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// generateUnionsV3 generates the anyOf and oneOf of an object schema into
// outSchema's union members.
func generateUnionsV3(ctx genContext, outSchema *Schema, schema *openapi3.Schema, path []string) error {
	// Inline union members are named <path><index>. A schema with both
	// anyOf and oneOf would name both lists' members the same way and
	// fail with a duplicate type name (issue #839), so each list gets
	// its own path segment then. A schema with only one keeps its
	// names.
	anyOfPath, oneOfPath := path, path
	if schema.AnyOf != nil && schema.OneOf != nil {
		anyOfPath = append(slices.Clone(path), "AnyOf")
		oneOfPath = append(slices.Clone(path), "OneOf")
	}
	if schema.AnyOf != nil {
		if err := generateUnionV3(ctx, outSchema, schema.AnyOf, schema.Discriminator, anyOfPath); err != nil {
			return fmt.Errorf("error generating type for anyOf: %w", err)
		}
	}
	if schema.OneOf != nil {
		if err := generateUnionV3(ctx, outSchema, schema.OneOf, schema.Discriminator, oneOfPath); err != nil {
			return fmt.Errorf("error generating type for oneOf: %w", err)
		}
	}
	return nil
}

func generateUnionV3(ctx genContext, outSchema *Schema, elements openapi3.SchemaRefs, discriminator *openapi3.Discriminator, path []string) error {
	if discriminator != nil {
		outSchema.Discriminator = &Discriminator{
			Property:  discriminator.PropertyName,
			Mapping:   make(map[string]string),
			ValueType: discriminatorValueType(outSchema, elements, discriminator.PropertyName),
		}
	}

	// First pass: count effective (non-null) branches. In OpenAPI 3.1, a
	// bare `{"type": "null"}` branch in anyOf/oneOf is a nullability
	// marker, not a real union variant -- there's no Go type that
	// corresponds to "only the JSON value null". The parent schema's
	// nullability is captured by schemaIsNullable, which inspects
	// anyOf/oneOf for the same idiom and wraps the result in a pointer
	// at the call site.
	effectiveCount := 0
	hadNullBranch := false
	var soleEffective *openapi3.SchemaRef
	for _, e := range elements {
		if e != nil && isNullTypeSchema(e.Value) {
			hadNullBranch = true
			continue
		}
		effectiveCount++
		if soleEffective == nil {
			soleEffective = e
		}
	}

	// Collapse: if filtering out null branches leaves exactly one
	// effective branch and there is no discriminator, the schema is
	// semantically equivalent to that single branch (made nullable by
	// the original null branch). Produce the same Go shape the
	// type-array idiom would: `anyOf: [{type: string}, {type: "null"}]`
	// must generate the same `*string` field as `type: ["string",
	// "null"]`. Without this, the single remaining branch would be
	// wrapped in a one-variant union type, exposing a needless
	// `FromX`/`AsX` accessor API.
	//
	// We do not collapse when there was no null branch (`anyOf: [{type:
	// X}]` alone) to avoid changing behavior for existing single-branch
	// union specs that may rely on the wrapper shape. The narrow
	// condition keeps this change scoped to the bug fix.
	if effectiveCount == 1 && hadNullBranch && discriminator == nil {
		elementSchema, err := generateGoSchema(ctx.at(path), soleEffective, path)
		if err != nil {
			return err
		}
		// Inherit the single branch's underlying representation. The
		// caller will apply nullability (schemaIsNullable returns true
		// because the original anyOf/oneOf contained a null branch).
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

	refToGoTypeMap := make(map[string]string)
	for i, element := range elements {
		// Skip null-only branches: nullability marker, not a real
		// union variant. See the collapse comment above for context.
		if element != nil && isNullTypeSchema(element.Value) {
			continue
		}
		elementPath := append(path, fmt.Sprint(i))
		elementSchema, err := generateGoSchema(ctx.at(elementPath), element, elementPath)
		if err != nil {
			return err
		}

		if element.Ref == "" {
			elementName := SchemaNameToTypeName(PathToTypeName(elementPath))
			if elementSchema.TypeDecl() == elementName {
				elementSchema.GoType = elementName
			} else {
				td := TypeDefinition{Schema: elementSchema, TypeName: elementName, JsonName: strings.Join(elementPath, ".")}
				outSchema.AdditionalTypes = append(outSchema.AdditionalTypes, td)
				elementSchema.GoType = td.TypeName
			}
			outSchema.AdditionalTypes = append(outSchema.AdditionalTypes, elementSchema.AdditionalTypes...)
		} else {
			refToGoTypeMap[element.Ref] = elementSchema.GoType
		}

		if discriminator != nil {
			if len(discriminator.Mapping) != 0 && element.Ref == "" {
				return errors.New("ambiguous discriminator.mapping: please replace inlined object with $ref")
			}

			// Explicit mapping.
			var mapped bool
			for k, v := range discriminator.Mapping {
				if v.Ref == element.Ref {
					outSchema.Discriminator.Mapping[k] = elementSchema.GoType
					mapped = true
				}
			}
			// Implicit mapping.
			if !mapped {
				outSchema.Discriminator.Mapping[RefPathToObjName(element.Ref)] = elementSchema.GoType
			}
		}
		// The same type can appear twice, e.g. as a member of both an anyOf
		// and a oneOf; its accessors are generated once.
		if !slices.Contains(outSchema.UnionElements, UnionElement(elementSchema.GoType)) {
			outSchema.UnionElements = append(outSchema.UnionElements, UnionElement(elementSchema.GoType))
		}
		for _, name := range propertyNames(element.Value, 0) {
			if !slices.Contains(outSchema.UnionVariantProperties, name) {
				outSchema.UnionVariantProperties = append(outSchema.UnionVariantProperties, name)
			}
		}
	}
	slices.Sort(outSchema.UnionVariantProperties)

	// Compare against effectiveCount (non-null branches actually
	// processed) rather than len(elements). For a nullable
	// discriminated union (`oneOf: [Cat, Dog, {type: "null"}]`), the
	// null-branch skip above leaves the discriminator with one fewer
	// mapping than the raw element count, and we must not flag that as
	// incomplete -- the null branch is a nullability marker, not a real
	// variant that needs a mapping.
	if (outSchema.Discriminator != nil) && len(outSchema.Discriminator.Mapping) < effectiveCount {
		return errors.New("discriminator: not all schemas were mapped")
	}

	return nil
}
