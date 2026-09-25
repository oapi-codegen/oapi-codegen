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
//
// A schema that combines several unions, an anyOf and a oneOf or an allOf's
// merge of several, is a value that is one of each union's variants at once
// (see unionComponent), and generates as generateUnionComponentsV3 describes.
func generateUnionsV3(ctx genContext, outSchema *Schema, schema *openapi3.Schema, path []string) error {
	components := ctx.unionComponents[schema]
	if components == nil && schema.AnyOf != nil && schema.OneOf != nil {
		components = []unionComponent{
			{branches: schema.AnyOf, anyOf: true},
			{branches: schema.OneOf, discriminator: schema.Discriminator},
		}
		if sameBranches(schema.AnyOf, schema.OneOf) {
			components = components[1:]
		}
	}
	switch {
	case len(components) > 1:
		return generateUnionComponentsV3(ctx, outSchema, components, path)
	case len(components) == 1:
		if _, err := generateUnionV3(ctx, outSchema, components[0].branches, components[0].discriminator, path, true); err != nil {
			return fmt.Errorf("error generating type for oneOf: %w", err)
		}
	case schema.AnyOf != nil:
		if _, err := generateUnionV3(ctx, outSchema, schema.AnyOf, schema.Discriminator, path, true); err != nil {
			return fmt.Errorf("error generating type for anyOf: %w", err)
		}
	case schema.OneOf != nil:
		if _, err := generateUnionV3(ctx, outSchema, schema.OneOf, schema.Discriminator, path, true); err != nil {
			return fmt.Errorf("error generating type for oneOf: %w", err)
		}
	}
	return nil
}

// generateUnionComponentsV3 generates a union that combines several (see
// unionComponent): the variants of each, and for each variant the JSON keys
// its own union owns, those only its union's variants declare, which From*
// replaces while keeping the other unions' data (see
// Schema.UnionOwnedKeys). A key several unions declare, such as a shared id,
// belongs to all of them and is never removed. A union that is a $ref to a
// union type also gets As* and From* for that type.
//
// Inline variants are named <path><union><index>, where <union> is OneOf or
// AnyOf, numbered when there are several of a kind.
func generateUnionComponentsV3(ctx genContext, outSchema *Schema, components []unionComponent, path []string) error {
	// An opaque variant, from another document or an x-go-type, declares no
	// keys: its fields aren't ours to read (see isOpaqueSchema).
	// The discriminator's property is its union's too, even when the variants
	// leave it to the stamp.
	declared := make([][]string, len(components))
	for k, c := range components {
		for _, b := range c.branches {
			if b != nil && b.Value != nil && !isOpaqueSchema(b) {
				declared[k] = append(declared[k], variantKeys(b.Value)...)
			}
		}
		if c.discriminator != nil {
			declared[k] = append(declared[k], c.discriminator.PropertyName)
		}
	}
	owned := make([][]string, len(components))
	for k := range components {
		owned[k] = []string{}
		for _, key := range declared[k] {
			shared := false
			for j := range components {
				shared = shared || (j != k && slices.Contains(declared[j], key))
			}
			if !shared && !slices.Contains(owned[k], key) {
				owned[k] = append(owned[k], key)
			}
		}
		slices.Sort(owned[k])
	}

	outSchema.UnionOwnedKeys = make(map[UnionElement][]string)
	for k, c := range components {
		kind, count, index := "OneOf", 0, 0
		if c.anyOf {
			kind = "AnyOf"
		}
		for j, other := range components {
			if other.anyOf == c.anyOf {
				count++
				if j < k {
					index++
				}
			}
		}
		if count > 1 {
			kind += fmt.Sprint(index)
		}
		// The union has one discriminator, which the merge gives to one of
		// the unions (see allOfMerge.placeDiscriminator).
		discriminator := c.discriminator
		if outSchema.Discriminator != nil {
			discriminator = nil
		}
		members, err := generateUnionV3(ctx, outSchema, c.branches, discriminator, append(slices.Clone(path), kind), false)
		if err != nil {
			return fmt.Errorf("error generating type for %s: %w", strings.ToLower(kind[:1])+kind[1:], err)
		}
		// A variant of several of the unions, such as the same $ref in two,
		// is each one's: setting it replaces what each of them owns.
		for _, e := range members {
			keys := append(slices.Clone(outSchema.UnionOwnedKeys[e]), owned[k]...)
			slices.Sort(keys)
			outSchema.UnionOwnedKeys[e] = slices.Compact(keys)
		}
		if c.ref != nil {
			union, err := generateGoSchema(ctx.at(path), c.ref, path)
			if err != nil {
				return err
			}
			var others []string
			for j := range components {
				if j != k {
					others = append(others, owned[j]...)
				}
			}
			slices.Sort(others)
			outSchema.UnionRefComponents = append(outSchema.UnionRefComponents,
				UnionRefComponent{Type: UnionElement(union.GoType), OwnedKeys: owned[k], OtherKeys: slices.Compact(others)})
		}
	}
	// A union type that is also a variant of another union has that
	// variant's As* and From* already.
	outSchema.UnionRefComponents = slices.DeleteFunc(outSchema.UnionRefComponents, func(c UnionRefComponent) bool {
		return slices.Contains(outSchema.UnionElements, c.Type)
	})
	return nil
}

// generateUnionV3 generates one oneOf or anyOf into outSchema's union members,
// and returns this union's members, some of which another union may have
// added already. alone is false when the union is one of several (see
// generateUnionComponentsV3), which then isn't collapsed into its only
// non-null branch.
func generateUnionV3(ctx genContext, outSchema *Schema, elements openapi3.SchemaRefs, discriminator *openapi3.Discriminator, path []string, alone bool) ([]UnionElement, error) {
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
	if alone && effectiveCount == 1 && hadNullBranch && discriminator == nil {
		elementSchema, err := generateGoSchema(ctx.at(path), soleEffective, path)
		if err != nil {
			return nil, err
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
		return nil, nil
	}

	refToGoTypeMap := make(map[string]string)
	var members []UnionElement
	for i, element := range elements {
		// Skip null-only branches: nullability marker, not a real
		// union variant. See the collapse comment above for context.
		if element != nil && isNullTypeSchema(element.Value) {
			continue
		}
		elementPath := append(path, fmt.Sprint(i))
		elementSchema, err := generateGoSchema(ctx.at(elementPath), element, elementPath)
		if err != nil {
			return nil, err
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
				return nil, errors.New("ambiguous discriminator.mapping: please replace inlined object with $ref")
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
		members = append(members, UnionElement(elementSchema.GoType))
		// The same type can appear twice, e.g. as a member of both an anyOf
		// and a oneOf; its accessors are generated once.
		if !slices.Contains(outSchema.UnionElements, UnionElement(elementSchema.GoType)) {
			outSchema.UnionElements = append(outSchema.UnionElements, UnionElement(elementSchema.GoType))
		}
		if discriminator != nil && !alone && !slices.Contains(outSchema.Discriminator.variants, UnionElement(elementSchema.GoType)) {
			outSchema.Discriminator.variants = append(outSchema.Discriminator.variants, UnionElement(elementSchema.GoType))
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
	if discriminator != nil && len(outSchema.Discriminator.Mapping) < effectiveCount {
		return nil, errors.New("discriminator: not all schemas were mapped")
	}

	return members, nil
}

// variantKeys returns the JSON keys a union variant's values can have: the
// properties it declares, with those of its allOf members however deep, under
// the names their Go fields marshal as. x-go-json-ignore leaves a property off
// the wire, and a json tag in x-oapi-codegen-extra-tags renames it.
func variantKeys(s *openapi3.Schema) []string {
	var keys []string
	seen := make(map[*openapi3.Schema]bool)
	var walk func(s *openapi3.Schema)
	walk = func(s *openapi3.Schema) {
		if s == nil || seen[s] {
			return
		}
		seen[s] = true
		for name, p := range s.Properties {
			if key, ok := propertyKey(name, combinedSchemaExtensions(p)); ok {
				keys = append(keys, key)
			}
		}
		for _, m := range s.AllOf {
			if m != nil {
				walk(m.Value)
			}
		}
	}
	walk(s)
	return keys
}

// propertyKey returns the JSON key a property's Go field marshals as, the way
// the struct tags are generated (see GenFieldsFromProperties), and false when
// the field isn't marshaled.
func propertyKey(name string, extensions map[string]any) (string, bool) {
	key := name
	if raw, ok := extensions[extPropGoJsonIgnore]; ok {
		if ignore, err := extParseGoJsonIgnore(raw); err == nil && ignore {
			key = "-"
		}
	}
	if raw, ok := extensions[extPropExtraTags]; ok {
		if tags, err := extExtraTags(raw); err == nil {
			if tag, ok := tags["json"]; ok {
				key, _, _ = strings.Cut(tag, ",")
				if key == "" {
					// encoding/json uses the Go field's name; keep the
					// property's, the nearest we can tell.
					key = name
				}
			}
		}
	}
	return key, key != "-"
}
