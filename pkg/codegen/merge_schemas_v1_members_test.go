package codegen

// Tests of how schema-merging-behavior v1 lowers an allOf member that is a
// $ref to a schema whose Go type is not a struct (issue #2590).
//
// v1 embeds every $ref member in a struct. encoding/json flattens only an
// embedded struct; it encodes an embedded field of any other type under a key
// named after the type. So a composition that holds just a primitive, array,
// map or untyped value is that member's type, as under v2, and an untyped
// member next to members that add fields stays out of the JSON. A member
// whose Go type may carry JSON methods of its own (time.Time, an x-go-type) is
// embedded as before: the struct inherits those methods.
//
// These tests record how v1 behaves. A bug fix may change them, but they must
// not be changed to accept a regression, and a commit that changes them must
// say why.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMergeSchemasV1ValueMember pins, for each kind of member schema, whether
// a composition of it and an annotation is the member's type or a struct that
// embeds it, as a component and as a property.
func TestMergeSchemasV1ValueMember(t *testing.T) {
	tests := []struct {
		name   string
		member string
		// alias is whether the composition is the member's type.
		alias bool
	}{
		{name: "string", member: "{type: string}", alias: true},
		{name: "integer", member: "{type: integer}", alias: true},
		{name: "enum", member: "{type: string, enum: [a, b]}", alias: true},
		{name: "typeless enum", member: "{enum: [a, 1]}", alias: true},
		{name: "array", member: "{type: array, items: {type: string}}", alias: true},
		{name: "map", member: "{type: object, additionalProperties: {type: string}}", alias: true},
		{name: "untyped", member: "{}", alias: true},
		{name: "x-go-type builtin", member: "{x-go-type: int64}", alias: true},
		{name: "x-go-type-name", member: "{x-go-type-name: MemberValue}", alias: true},
		{name: "object", member: "{type: object, properties: {v: {type: string}}}", alias: false},
		{name: "date-time", member: "{type: string, format: date-time}", alias: false},
		{name: "uuid", member: "{type: string, format: uuid}", alias: false},
		{name: "x-go-type named", member: "{x-go-type: LocalType}", alias: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := `openapi: 3.0.3
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    Member: ` + tt.member + `
    Wrap:
      allOf:
        - $ref: '#/components/schemas/Member'
        - description: A decorated Member.
    Holder:
      type: object
      properties:
        one:
          allOf:
            - $ref: '#/components/schemas/Member'
`
			code := generateSpec(t, spec, withSchemaMergingV1)
			if tt.alias {
				assert.Contains(t, code, "type Wrap = Member\n")
				assertField(t, code, "One", "*Member")
				assert.NotContains(t, code, "Member `yaml:\",inline\"`")
			} else {
				assert.Contains(t, code, "type Wrap struct {")
				assert.Contains(t, code, "One *struct {")
				assert.Equal(t, 2, strings.Count(code, "\tMember `yaml:\",inline\"`"))
			}
		})
	}
}

// TestMergeSchemasV1ValueMemberChain follows a composition that is itself an
// allOf of a value member: it is that member's type, so a composition of it
// is too.
func TestMergeSchemasV1ValueMemberChain(t *testing.T) {
	code := generateSpec(t, `openapi: 3.0.3
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    Name: {type: string}
    Wrap:
      allOf:
        - $ref: '#/components/schemas/Name'
    WrapWrap:
      allOf:
        - $ref: '#/components/schemas/Wrap'
        - description: Wrapped twice.
`, withSchemaMergingV1)

	assert.Contains(t, code, "type Wrap = Name\n")
	assert.Contains(t, code, "type WrapWrap = Wrap\n")
}

// TestMergeSchemasV1ValueMemberLongChain generates a chain of compositions,
// each an allOf of the one before. Classifying a member generates the schema
// it refers to, and the result is remembered (see genContext.v1MemberKinds);
// without that, each link is generated twice, and this takes time
// exponential in the chain's length.
func TestMergeSchemasV1ValueMemberLongChain(t *testing.T) {
	var spec strings.Builder
	spec.WriteString(`openapi: 3.0.3
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    Link0: {type: string}
`)
	for i := 1; i <= 60; i++ {
		fmt.Fprintf(&spec, "    Link%d:\n      allOf:\n        - $ref: '#/components/schemas/Link%d'\n        - description: Link %d.\n", i, i-1, i)
	}
	code := generateSpec(t, spec.String(), withSchemaMergingV1)

	assert.Contains(t, code, "type Link1 = Link0\n")
	assert.Contains(t, code, "type Link60 = Link59\n")
}

// TestMergeSchemasV1ValueMemberRecursive keeps a composition embedding its
// member when being an alias of the member's type would make a cycle of
// aliases: `type Root = Arr` with `type Arr = []Root` doesn't compile. A map
// is a defined type, which ends such a cycle, so Loop is an alias of it.
func TestMergeSchemasV1ValueMemberRecursive(t *testing.T) {
	code := generateSpec(t, `openapi: 3.0.3
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    Root:
      allOf:
        - $ref: '#/components/schemas/Arr'
        - description: A decorated list.
    Arr:
      type: array
      items: {$ref: '#/components/schemas/Root'}
    Loop:
      allOf:
        - $ref: '#/components/schemas/LoopMap'
    LoopMap:
      type: object
      additionalProperties: {$ref: '#/components/schemas/Loop'}
`, withSchemaMergingV1)

	assert.Contains(t, code, "type Arr = []Root\n")
	assert.Contains(t, code, "type Root struct {")
	assert.Contains(t, code, "\tArr `yaml:\",inline\"`")
	assert.Contains(t, code, "type Loop = LoopMap\n")
}

// TestMergeSchemasV1ClassifyingMembersIsIsolated generates a member's schema
// only to see its type, and that must not tell the composition being
// generated that something referred back to it. Here B's items refer back to
// Root, whose merge is in progress when B is classified. Were Root's frame
// told, generation would fail: a recursive composition at the root of a
// component is not supported.
func TestMergeSchemasV1ClassifyingMembersIsIsolated(t *testing.T) {
	code := generateSpec(t, `openapi: 3.0.3
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    Root:
      allOf:
        - $ref: '#/components/schemas/B'
        - type: object
          properties:
            x: {type: string}
    B:
      type: array
      items:
        allOf:
          - $ref: '#/components/schemas/Root'
`, withSchemaMergingV1)

	assert.Contains(t, code, "type Root struct {")
	assert.Contains(t, code, "\tB `yaml:\",inline\"`")
	assertField(t, code, "X", "*string")
	assert.Contains(t, code, "type B = []struct {")
}

// TestMergeSchemasV1UntypedMember records what becomes of an untyped member,
// which places no constraint and carries no value of its own. A composition
// of untyped members is the first one's type, and one with a member that
// holds a value is that member's type. Next to a member that adds fields or
// brings a type of its own, the untyped member is embedded but stays out of
// the JSON.
func TestMergeSchemasV1UntypedMember(t *testing.T) {
	code := generateSpec(t, `openapi: 3.0.3
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    Untyped: {}
    Other: {}
    Obj:
      type: object
      properties:
        v: {type: string}
    Labels:
      type: object
      additionalProperties: {type: string}
    Stamp:
      type: string
      format: date-time
    OnlyUntyped:
      allOf:
        - $ref: '#/components/schemas/Untyped'
        - $ref: '#/components/schemas/Other'
        - description: Untyped twice over.
    WithMap:
      allOf:
        - $ref: '#/components/schemas/Untyped'
        - $ref: '#/components/schemas/Labels'
        - description: Labels, decorated.
    WithInline:
      allOf:
        - $ref: '#/components/schemas/Untyped'
        - type: object
          properties:
            name: {type: string}
    WithRef:
      allOf:
        - $ref: '#/components/schemas/Untyped'
        - $ref: '#/components/schemas/Obj'
    WithOpaque:
      allOf:
        - $ref: '#/components/schemas/Untyped'
        - $ref: '#/components/schemas/Stamp'
`, withSchemaMergingV1)

	assert.Contains(t, code, "type OnlyUntyped = Untyped\n")
	assert.Contains(t, code, "type WithMap = Labels\n")

	hidden := "\tUntyped `json:\"-\" yaml:\",inline\"`"
	assert.Equal(t, 3, strings.Count(code, hidden), "WithInline, WithRef and WithOpaque hide Untyped")
	assert.NotContains(t, code, "\tUntyped `yaml:\",inline\"`")
	assert.Contains(t, code, "\tObj `yaml:\",inline\"`")
	assert.Contains(t, code, "\tStamp `yaml:\",inline\"`")
}

func TestV1GoTypeKind(t *testing.T) {
	for decl, kind := range map[string]v1MemberKind{
		"struct {\n\tV string\n}": v1Fields,
		"any":                     v1Untyped,
		"interface{}":             v1Untyped,
		"interface{ V() string }": v1Opaque,
		"string":                  v1Value,
		"int64":                   v1Value,
		"[]string":                v1Value,
		"[]Item":                  v1Value,
		"map[string]string":       v1Value,
		"error":                   v1Opaque,
		"time.Time":               v1Opaque,
		"openapi_types.UUID":      v1Opaque,
		"LocalType":               v1Opaque,
		"*string":                 v1Opaque,
		"nullable.Nullable[int]":  v1Opaque,
		"not a type (":            v1Opaque,
	} {
		assert.Equal(t, kind, v1GoTypeKind(decl), decl)
	}
}
