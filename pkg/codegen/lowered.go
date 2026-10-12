package codegen

// This file holds how the generator finds the schema a Go type was taken
// from, when a schema's type is not its own but one it refers to. Two
// classifications need it, each with a question of its own: whether a
// strict-server response body can be its own type (operations.go), and what
// schema-merging-behavior v1 makes of an allOf member (merge_schemas_v1.go).

import "github.com/getkin/kin-openapi/openapi3"

// subschemaLists returns a schema's allOf, anyOf and oneOf.
func subschemaLists(s *openapi3.Schema) []openapi3.SchemaRefs {
	return []openapi3.SchemaRefs{s.AllOf, s.AnyOf, s.OneOf}
}

// namedTypeDef returns the type definition x-go-type-name declared beside
// the schema, when the schema's Go type is that name. It is where the
// schema's type is spelled out.
func (s Schema) namedTypeDef() (TypeDefinition, bool) {
	decl := s.TypeDecl()
	for _, td := range s.AdditionalTypes {
		if td.TypeName == decl {
			return td, true
		}
	}
	return TypeDefinition{}, false
}

// loweredToMember returns the $ref member a composition was lowered to, with
// the schema generate makes of that member: the member whose generated type
// is decl, the composition's own type, such as the $ref of `allOf: [$ref X]`
// or the branch a nullable oneOf collapsed to. Whether a composition is
// lowered that way depends on the schema-merging behavior, so the generated
// types are compared rather than the shape of the schema: a composition that
// became a struct or union type of its own matches no member. A $ref
// generates as its name without being followed, so generate is cheap here.
func loweredToMember(schema *openapi3.Schema, decl string, generate func(*openapi3.SchemaRef) (Schema, error)) (*openapi3.SchemaRef, Schema, error) {
	for _, members := range subschemaLists(schema) {
		for _, member := range members {
			if member == nil || !IsGoTypeReference(member.Ref) {
				continue
			}
			memberSchema, err := generate(member)
			if err != nil {
				return nil, Schema{}, err
			}
			if memberSchema.TypeDecl() == decl {
				return member, memberSchema, nil
			}
		}
	}
	return nil, Schema{}, nil
}
