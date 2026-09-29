package codegen

// This file holds schema-merging-behavior v1: allOf members that are $refs
// are embedded in the Go struct, and the fields of inline members are inlined.
// It was the only behavior before oapi-codegen v1.11.0, and OldMergeSchemas:
// true selects it too. For the rest of allOf (the parent's siblings, recursive
// compositions) and for anyOf/oneOf, v1 uses v2's code.
//
// v1 is kept for compatibility. Bug fixes are welcome, but a change here
// must not cause a regression: a spec that generated working code must keep
// generating it. New behavior goes into a new version, in files of its own.

import (
	"errors"
	"fmt"
	"go/ast"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// mergeSchemasV1 generates the members with the context of the allOf they
// belong to, as mergeSchemasV2 does, so that a member which leads back into a
// composition an enclosing frame is still generating resolves to that
// composition's type (see genContext.inProgress) instead of being generated
// again without end. A member that is a $ref to a Go type never gets that
// far, since generateGoSchema returns the referenced type before descending,
// which is why v1 only ever recursed through a whole-document $ref
// (`$ref: ./node.yaml`), whose value is generated inline.
func mergeSchemasV1(ctx genContext, allOf []*openapi3.SchemaRef, path []string) (Schema, error) {
	var outSchema Schema
	kinds := make([]v1MemberKind, len(allOf))
	for i, schemaOrRef := range allOf {
		ref := schemaOrRef.Ref

		var refType string
		var err error
		if IsGoTypeReference(ref) {
			refType, err = RefPathToGoType(ref)
			if err != nil {
				return Schema{}, fmt.Errorf("error converting reference path to a go type: %w", err)
			}
		}

		schema, err := generateGoSchema(ctx, schemaOrRef, path)
		if err != nil {
			return Schema{}, fmt.Errorf("error generating Go schema in allOf: %w", err)
		}
		schema.RefType = refType
		if IsGoTypeReference(ref) {
			kinds[i] = v1RefMemberKind(ctx, schemaOrRef, map[string]bool{})
		} else {
			kinds[i] = v1InlineMemberKind(schema)
		}

		for _, p := range schema.Properties {
			err = outSchema.AddProperty(p)
			if err != nil {
				return Schema{}, fmt.Errorf("error merging properties: %w", err)
			}
		}

		if schema.HasAdditionalProperties {
			if outSchema.HasAdditionalProperties {
				// Both this schema, and the aggregate schema have additional
				// properties, they must match.
				if schema.AdditionalPropertiesType.TypeDecl() != outSchema.AdditionalPropertiesType.TypeDecl() {
					return Schema{}, errors.New("additional properties in allOf have incompatible types")
				}
			} else {
				// We're switching from having no additional properties to having
				// them
				outSchema.HasAdditionalProperties = true
				outSchema.AdditionalPropertiesType = schema.AdditionalPropertiesType
			}
		}
	}

	// A composition of one $ref to a primitive, array or map schema, or of
	// untyped $refs, next to members that add no fields, holds just that
	// value. Embedded in a struct, encoding/json would put it under a key
	// named after its type (issue #2590), so it is the member's type
	// instead, as under v2, unless being an alias of that type would make an
	// invalid cycle of aliases.
	if i, ok := v1SoleValueMember(kinds); ok && !v1AliasRefersBack(ctx, allOf[i]) {
		return generateGoSchema(ctx, allOf[i], path)
	}

	// Now, we generate the struct which merges together all the fields.
	var err error
	outSchema.GoType, err = genStructFromAllOf(ctx, allOf, kinds, path)
	if err != nil {
		return Schema{}, fmt.Errorf("unable to generate aggregate type for AllOf: %w", err)
	}
	return outSchema, nil
}

// GenStructFromAllOf generates an object that is the union of the objects in the
// input array. In the case of Ref objects, we use an embedded struct, otherwise,
// we inline the fields.
//
// It starts a fresh generation context; within the package, mergeSchemasV1
// hands genStructFromAllOf the context of the allOf being merged.
func GenStructFromAllOf(allOf []*openapi3.SchemaRef, path []string) (string, error) {
	ctx := newGenContext(path)
	kinds := make([]v1MemberKind, len(allOf))
	for i, schemaOrRef := range allOf {
		if IsGoTypeReference(schemaOrRef.Ref) {
			kinds[i] = v1RefMemberKind(ctx, schemaOrRef, map[string]bool{})
			continue
		}
		schema, err := generateGoSchema(ctx, schemaOrRef, path)
		if err != nil {
			return "", err
		}
		kinds[i] = v1InlineMemberKind(schema)
	}
	return genStructFromAllOf(ctx, allOf, kinds, path)
}

// genStructFromAllOf generates the struct for allOf, whose members kinds
// classifies (see v1MemberKind).
func genStructFromAllOf(ctx genContext, allOf []*openapi3.SchemaRef, kinds []v1MemberKind, path []string) (string, error) {
	// An untyped member places no constraint of its own and carries no value
	// apart from the one the other members describe, so when one of them
	// adds fields or brings a type of its own, the untyped member stays out
	// of the JSON rather than adding a key named after its type (issue
	// #2590). Only when no member does is it left as it was.
	hideUntyped := slices.ContainsFunc(kinds, func(kind v1MemberKind) bool {
		return kind != v1NoFields && kind != v1Untyped
	})

	// Start out with struct {
	objectParts := []string{"struct {"}
	for i, schemaOrRef := range allOf {
		ref := schemaOrRef.Ref
		if IsGoTypeReference(ref) {
			// We have a referenced type, we will generate an inlined struct
			// member.
			// struct {
			//   InlinedMember
			//   ...
			// }
			goType, err := RefPathToGoType(ref)
			if err != nil {
				return "", err
			}
			objectParts = append(objectParts,
				fmt.Sprintf("   // Embedded struct due to allOf(%s)", ref))
			tag := "`yaml:\",inline\"`"
			if hideUntyped && kinds[i] == v1Untyped {
				tag = "`json:\"-\" yaml:\",inline\"`"
			}
			objectParts = append(objectParts,
				fmt.Sprintf("   %s %s", goType, tag))
		} else {
			// Inline all the fields from the schema into the output struct,
			// just like in the simple case of generating an object.
			goSchema, err := generateGoSchema(ctx, schemaOrRef, path)
			if err != nil {
				return "", err
			}
			objectParts = append(objectParts, "   // Embedded fields due to inline allOf schema")
			objectParts = append(objectParts, GenFieldsFromProperties(goSchema.Properties)...)

			if goSchema.HasAdditionalProperties {
				addPropsType := goSchema.AdditionalPropertiesType.GoType
				if goSchema.AdditionalPropertiesType.RefType != "" {
					addPropsType = goSchema.AdditionalPropertiesType.RefType
				}

				additionalPropertiesPart := fmt.Sprintf("AdditionalProperties map[string]%s `json:\"-\"`", addPropsType)
				if !slices.Contains(objectParts, additionalPropertiesPart) {
					objectParts = append(objectParts, additionalPropertiesPart)
				}
			}
		}
	}
	objectParts = append(objectParts, "}")
	return strings.Join(objectParts, "\n"), nil
}

// v1MemberKind says what an allOf member contributes to the struct v1
// generates for the allOf, and what encoding/json makes of it there.
type v1MemberKind int

const (
	// v1Fields: the member adds fields. It is an inline member with
	// properties or additionalProperties, or a $ref to a struct, whose
	// fields encoding/json flattens into the struct that embeds it.
	v1Fields v1MemberKind = iota
	// v1NoFields: an inline member that adds no fields, such as one that
	// only carries a description.
	v1NoFields
	// v1Untyped: a $ref to a schema whose Go type is any.
	v1Untyped
	// v1Value: a $ref to a primitive, array or map. Its Go type has no
	// methods of its own, and encoding/json encodes it embedded as a field
	// named after the type rather than flattening it.
	v1Value
	// v1Opaque: a $ref whose Go type may have JSON methods of its own that
	// the struct embedding it inherits, such as time.Time or an x-go-type,
	// or whose type this can't tell. It is embedded, as it always has been.
	v1Opaque
)

// v1InlineMemberKind classifies an inline allOf member from its generated
// schema: the struct gets its fields, and nothing else of it.
func v1InlineMemberKind(schema Schema) v1MemberKind {
	if len(schema.Properties) > 0 || schema.HasAdditionalProperties {
		return v1Fields
	}
	return v1NoFields
}

// v1RefMemberKind classifies an allOf member that is a $ref to a Go type by
// the type the generator declares for the referenced schema. seen holds the
// $refs already followed, against a cycle of them.
func v1RefMemberKind(ctx genContext, member *openapi3.SchemaRef, seen map[string]bool) v1MemberKind {
	// An x-go-type next to the $ref names a type this can't see.
	if _, ok := member.Extensions[extPropGoType]; ok || member.Value == nil || seen[member.Ref] {
		return v1Opaque
	}
	if kind, ok := ctx.v1MemberKinds[member.Value]; ok {
		return kind
	}
	seen[member.Ref] = true
	kind := v1ClassifyRefMember(ctx, member, seen)
	ctx.v1MemberKinds[member.Value] = kind
	return kind
}

// v1ClassifyRefMember is v1RefMemberKind without its cache.
func v1ClassifyRefMember(ctx genContext, member *openapi3.SchemaRef, seen map[string]bool) v1MemberKind {
	typeName, err := RefPathToGoType(member.Ref)
	if err != nil {
		return v1Opaque
	}
	// Generate the referenced schema as its component is generated, under
	// its own name, in a context of its own (see genContext.isolated).
	path := []string{typeName}
	schema, err := generateGoSchema(ctx.isolated(path), &openapi3.SchemaRef{Value: member.Value}, path)
	if err != nil {
		return v1Opaque
	}
	decl := schema.TypeDecl()
	if kind := v1GoTypeKind(decl); kind != v1Opaque {
		return kind
	}
	// A composition lowered to the type of one of its $ref members, such as
	// an allOf of a single $ref, is that member's type.
	for _, members := range []openapi3.SchemaRefs{member.Value.AllOf, member.Value.AnyOf, member.Value.OneOf} {
		for _, m := range members {
			if m == nil || !IsGoTypeReference(m.Ref) {
				continue
			}
			if name, err := RefPathToGoType(m.Ref); err == nil && name == decl {
				return v1RefMemberKind(ctx, m, seen)
			}
		}
	}
	// x-go-type-name declares the real type alongside the schema.
	for _, td := range schema.AdditionalTypes {
		if td.TypeName == decl {
			return v1GoTypeKind(td.Schema.TypeDecl())
		}
	}
	return v1Opaque
}

// v1GoTypeKind classifies a Go type declaration by what embedding it does.
// Only types that can't carry methods of their own are v1Untyped or v1Value:
// a named type, even a local one, may have JSON methods the embedding struct
// inherits.
func v1GoTypeKind(decl string) v1MemberKind {
	expr, _, ok := parseGoType(decl)
	if !ok {
		return v1Opaque
	}
	switch expr := unparen(expr).(type) {
	case *ast.StructType:
		return v1Fields
	case *ast.InterfaceType:
		// interface{} is any. An interface with methods is a type of its
		// own, as a named type is.
		if expr.Methods == nil || len(expr.Methods.List) == 0 {
			return v1Untyped
		}
		return v1Opaque
	case *ast.ArrayType, *ast.MapType:
		return v1Value
	case *ast.Ident:
		switch {
		case expr.Name == "any":
			return v1Untyped
		case expr.Name == "error" || expr.Name == "comparable":
			return v1Opaque
		case isPredeclaredType(expr.Name):
			return v1Value
		}
	}
	return v1Opaque
}

// v1SoleValueMember returns the index of the member of an allOf that holds
// its whole value: a $ref to a primitive, array or map schema, when every
// other member adds no fields or is untyped. An untyped member places no
// constraint and carries no value of its own, so a composition of untyped
// members and members that add no fields is the first untyped member's type.
func v1SoleValueMember(kinds []v1MemberKind) (int, bool) {
	value, untyped := -1, -1
	for i, kind := range kinds {
		switch kind {
		case v1NoFields:
		case v1Untyped:
			if untyped < 0 {
				untyped = i
			}
		case v1Value:
			if value >= 0 {
				return 0, false
			}
			value = i
		default:
			return 0, false
		}
	}
	if value >= 0 {
		return value, true
	}
	return untyped, untyped >= 0
}

// v1AliasRefersBack reports whether member's type, followed through the
// schemas that may be generated as aliases (arrays, and compositions, which
// may lower to one of their members), leads back to a composition being
// generated. The composition would then be an alias of a type that is itself
// an alias mentioning it, such as `type Root = Arr` with `type Arr = []Root`,
// which Go rejects as an invalid recursive type. A struct, map or enum is a
// defined type, which ends such a cycle, so the walk stops there.
func v1AliasRefersBack(ctx genContext, member *openapi3.SchemaRef) bool {
	seen := map[*openapi3.Schema]bool{}
	var refersBack func(*openapi3.SchemaRef) bool
	refersBack = func(sref *openapi3.SchemaRef) bool {
		if sref == nil || sref.Value == nil {
			return false
		}
		schema := sref.Value
		if _, ok := ctx.inProgress[schema]; ok {
			return true
		}
		if seen[schema] {
			return false
		}
		seen[schema] = true
		if schemaPrimaryType(schema.Type).Is("array") && refersBack(schema.Items) {
			return true
		}
		for _, members := range []openapi3.SchemaRefs{schema.AllOf, schema.AnyOf, schema.OneOf} {
			for _, m := range members {
				if refersBack(m) {
					return true
				}
			}
		}
		return false
	}
	return refersBack(member)
}
