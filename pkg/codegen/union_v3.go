package codegen

// This file holds the anyOf/oneOf half of schema-merging-behavior v3, the
// version under development. Its allOf half is in merge_schemas_v3.go. It
// started as a copy of v2's (union_v2.go).

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode"

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
		valueType := discriminatorValueType(outSchema, elements, discriminator.PropertyName)
		if !discriminatorPropertyTyped(outSchema, elements, discriminator.PropertyName) {
			var err error
			if valueType, err = pinnedValueType(elements, discriminator.PropertyName); err != nil {
				return nil, err
			}
		}
		outSchema.Discriminator = &Discriminator{
			Property:  discriminator.PropertyName,
			Mapping:   make(map[string]string),
			ValueType: valueType,
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
		outSchema.UnionElements = elementSchema.UnionElements
		outSchema.Discriminator = elementSchema.Discriminator
		outSchema.AdditionalTypes = append(outSchema.AdditionalTypes, elementSchema.AdditionalTypes...)
		return nil, nil
	}

	refToGoTypeMap := make(map[string]string)
	var members []UnionElement
	// inlineKeys are the discriminator values inline variants declare (see
	// inlineDiscriminatorValue), which no other variant may also map to.
	inlineKeys := make(map[string]bool)
	// mappedCount counts the variants the discriminator leads to.
	mappedCount := 0
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

		// An inline variant has no name for the mapping to use. It can still
		// say which value it takes, with a single-value enum or a const on the
		// discriminator property. One that doesn't has no value: the
		// discriminator doesn't lead to it, and From* stamps none, but its As*
		// and From* work as for any variant.
		var key string
		var keyed bool
		if discriminator != nil && element.Ref == "" {
			var err error
			if key, keyed, err = inlineDiscriminatorValue(element.Value, discriminator.PropertyName); err != nil {
				return nil, fmt.Errorf("discriminator: the %s value of the inline schema %s %w", discriminator.PropertyName, strings.Join(elementPath, "."), err)
			}
		}
		discriminated := keyed || (discriminator != nil && element.Ref != "")
		switch {
		case keyed:
			// The value is written into Go and JSON string literals.
			for _, r := range key {
				if r == '"' || r == '`' || r == '\\' || unicode.IsControl(r) {
					return nil, fmt.Errorf("discriminator: the %s value %q of the inline schema %s may not contain %s", discriminator.PropertyName, key, strings.Join(elementPath, "."), describeRune(r))
				}
			}
			_, mapped := discriminator.Mapping[key]
			_, taken := outSchema.Discriminator.Mapping[key]
			if mapped || taken {
				return nil, fmt.Errorf("discriminator: the inline schema %s takes the %s value %q, which another schema is mapped to", strings.Join(elementPath, "."), discriminator.PropertyName, key)
			}
			inlineKeys[key] = true
			outSchema.Discriminator.Mapping[key] = elementSchema.GoType
		case discriminated:
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
				key := RefPathToObjName(element.Ref)
				if inlineKeys[key] {
					return nil, fmt.Errorf("discriminator: %s takes the %s value %q, which an inline schema also takes", element.Ref, discriminator.PropertyName, key)
				}
				outSchema.Discriminator.Mapping[key] = elementSchema.GoType
			}
		}
		members = append(members, UnionElement(elementSchema.GoType))
		// The same type can appear twice, e.g. as a member of both an anyOf
		// and a oneOf; its accessors are generated once.
		if !slices.Contains(outSchema.UnionElements, UnionElement(elementSchema.GoType)) {
			outSchema.UnionElements = append(outSchema.UnionElements, UnionElement(elementSchema.GoType))
		}
		if discriminated {
			mappedCount++
			if !slices.Contains(outSchema.Discriminator.variants, UnionElement(elementSchema.GoType)) {
				outSchema.Discriminator.variants = append(outSchema.Discriminator.variants, UnionElement(elementSchema.GoType))
			}
		}
		for _, name := range propertyNames(element.Value, 0) {
			if !slices.Contains(outSchema.UnionVariantProperties, name) {
				outSchema.UnionVariantProperties = append(outSchema.UnionVariantProperties, name)
			}
		}
	}
	slices.Sort(outSchema.UnionVariantProperties)

	// Compare against the variants the discriminator leads to rather than
	// len(elements): a null branch is a nullability marker, not a variant
	// that needs a mapping, and an inline variant that pins no value has
	// none.
	if discriminator != nil && len(outSchema.Discriminator.Mapping) < mappedCount {
		return nil, errors.New("discriminator: not all schemas were mapped")
	}
	// An integer or number discriminator is compared as a number (see
	// Discriminator.SwitchType): each value is written as a literal of that
	// type, and one value can't lead to two variants.
	if d := outSchema.Discriminator; discriminator != nil && (d.ValueType == openapi3.TypeInteger || d.ValueType == openapi3.TypeNumber) {
		type lead struct{ key, goType string }
		leads := make(map[string]lead)
		d.literals = make(map[string]string, len(d.Mapping))
		for _, key := range SortedMapKeys(d.Mapping) {
			literal, err := numberLiteral(key, d.ValueType)
			if err != nil {
				return nil, fmt.Errorf("discriminator: the %s value %s %w", discriminator.PropertyName, key, err)
			}
			goType := d.Mapping[key]
			if other, ok := leads[literal]; ok && other.goType != goType {
				return nil, fmt.Errorf("discriminator: the %s values %s and %s are the same number, but lead to %s and %s",
					discriminator.PropertyName, other.key, key, other.goType, goType)
			}
			leads[literal] = lead{key, goType}
			d.literals[key] = literal
		}
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

// inlineDiscriminatorValue returns the value an inline union variant pins for
// the discriminator property (see pinnedDiscriminatorValue), as the
// discriminator mapping spells it: a string as it is, a number or a boolean as
// encoding/json writes it. An integer of magnitude 2^53 or more is an error:
// kin-openapi reads numbers as float64, so it may already have been rounded,
// and the value written would differ from the spec's.
func inlineDiscriminatorValue(s *openapi3.Schema, property string) (string, bool, error) {
	switch v := pinnedDiscriminatorValue(s, property).(type) {
	case string:
		return v, true, nil
	case bool:
		return strconv.FormatBool(v), true, nil
	case float64:
		if v == math.Trunc(v) && math.Abs(v) >= 1<<53 {
			return "", false, errors.New("can't be read exactly: numbers are read as float64, which holds integers only up to 2^53; map a $ref variant to it with an explicit mapping key instead")
		}
		b, err := json.Marshal(v)
		return string(b), err == nil, nil
	}
	return "", false, nil
}

// pinnedDiscriminatorValue returns the value a union variant pins for the
// discriminator property, with a const or an enum of one value besides null.
// That is on any declaration of the property, in the variant or its allOf
// members, or in an allOf member of the property's own schema. It returns nil
// when there is none.
func pinnedDiscriminatorValue(s *openapi3.Schema, property string) any {
	seen := make(map[*openapi3.Schema]bool)
	var pinned, declared func(s *openapi3.Schema) any
	pinned = func(p *openapi3.Schema) any {
		if p == nil || seen[p] {
			return nil
		}
		seen[p] = true
		if p.Const != nil {
			return p.Const
		}
		if values := slices.DeleteFunc(slices.Clone(p.Enum), func(v any) bool { return v == nil }); len(values) == 1 {
			return values[0]
		}
		for _, m := range p.AllOf {
			if m != nil {
				if v := pinned(m.Value); v != nil {
					return v
				}
			}
		}
		return nil
	}
	declared = func(s *openapi3.Schema) any {
		if s == nil || seen[s] {
			return nil
		}
		seen[s] = true
		if p := s.Properties[property]; p != nil {
			if v := pinned(p.Value); v != nil {
				return v
			}
		}
		for _, m := range s.AllOf {
			if m != nil {
				if v := declared(m.Value); v != nil {
					return v
				}
			}
		}
		return nil
	}
	return declared(s)
}

// discriminatorPropertyTyped reports whether the union itself, or any
// declaration of the property in its variants or their allOf members, gives
// the discriminator property a type.
func discriminatorPropertyTyped(outSchema *Schema, elements openapi3.SchemaRefs, property string) bool {
	for _, p := range outSchema.Properties {
		if p.JsonFieldName == property && p.Schema.OAPISchema != nil && len(nonNullTypes(p.Schema.OAPISchema.Type)) > 0 {
			return true
		}
	}
	seen := make(map[*openapi3.Schema]bool)
	var typed func(s *openapi3.Schema) bool
	typed = func(s *openapi3.Schema) bool {
		if s == nil || seen[s] {
			return false
		}
		seen[s] = true
		if p := s.Properties[property]; p != nil && p.Value != nil && len(nonNullTypes(p.Value.Type)) > 0 {
			return true
		}
		return slices.ContainsFunc(s.AllOf, func(m *openapi3.SchemaRef) bool { return m != nil && typed(m.Value) })
	}
	return slices.ContainsFunc(elements, func(e *openapi3.SchemaRef) bool { return e != nil && typed(e.Value) })
}

// pinnedValueType returns the JSON type the union's variants' pinned
// discriminator values have (see pinnedDiscriminatorValue), for a
// discriminator property that declares no type, like 3.1's `const: true`:
// "boolean" or "number", or "" for strings or when nothing is pinned. Values
// of different types are an error: no one discriminator reads them all.
func pinnedValueType(elements openapi3.SchemaRefs, property string) (string, error) {
	found := ""
	for _, e := range elements {
		if e == nil || e.Value == nil || isNullTypeSchema(e.Value) {
			continue
		}
		var typ string
		switch pinnedDiscriminatorValue(e.Value, property).(type) {
		case bool:
			typ = openapi3.TypeBoolean
		case float64:
			typ = openapi3.TypeNumber
		case string:
			typ = openapi3.TypeString
		default:
			continue
		}
		if found != "" && typ != found {
			return "", fmt.Errorf("discriminator: the variants pin %s values of different JSON types, %s and %s", property, found, typ)
		}
		found = typ
	}
	if found == openapi3.TypeString {
		return "", nil
	}
	return found, nil
}

// numberLiteral writes a value of an integer or number discriminator as a Go
// literal of the type ValueByDiscriminator decodes it into, int64 or float64
// (see Discriminator.SwitchType): the value itself, whatever its spelling, so
// 1.0 of an integer discriminator is 1.
func numberLiteral(text, valueType string) (string, error) {
	if text == "" || strings.TrimSpace(text) != text || !json.Valid([]byte(text)) ||
		(text[0] != '-' && (text[0] < '0' || text[0] > '9')) {
		return "", errors.New("isn't a JSON number")
	}
	if valueType == openapi3.TypeInteger {
		i, err := strconv.ParseInt(text, 10, 64)
		if err == nil {
			return strconv.FormatInt(i, 10), nil
		}
		if errors.Is(err, strconv.ErrRange) {
			return "", errors.New("doesn't fit in an int64")
		}
		// Not written as an integer, like 1.0 or 1e3: read as a float64,
		// which holds integers exactly only up to 2^53.
		f, err := strconv.ParseFloat(text, 64)
		switch {
		case err != nil || f != math.Trunc(f):
			return "", errors.New("isn't an integer")
		case math.Abs(f) >= 1<<53:
			return "", errors.New("can't be read exactly written this way; write it as an integer")
		}
		return strconv.FormatInt(int64(f), 10), nil
	}
	f, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return "", errors.New("doesn't fit in a float64")
	}
	if f == 0 {
		// -0 is 0, as a Go constant and when compared.
		return "0", nil
	}
	b, err := json.Marshal(f)
	return string(b), err
}
