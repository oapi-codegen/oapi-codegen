package codegen

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateSpec loads an inline OpenAPI spec and generates models from it.
func generateSpec(t *testing.T, spec string, opts ...func(*Configuration)) string {
	t.Helper()
	code, err := generateSpecErr(spec, opts...)
	require.NoError(t, err)
	return code
}

// generateSpecErr is generateSpec for the cases that are meant to fail.
func generateSpecErr(spec string, opts ...func(*Configuration)) (string, error) {
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData([]byte(spec))
	if err != nil {
		return "", err
	}
	cfg := Configuration{
		PackageName: "repro",
		Generate: GenerateOptions{
			Models: true,
		},
		OutputOptions: OutputOptions{
			SkipPrune: true,
		},
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return Generate(swagger, cfg)
}

// assertField asserts that code declares a struct field of the given name and
// Go type, ignoring the column alignment gofmt applies.
func assertField(t *testing.T, code, name, goType string) {
	t.Helper()
	assert.Regexp(t, `\n\t`+regexp.QuoteMeta(name)+`\s+`+regexp.QuoteMeta(goType)+`\s`, code,
		"expected a field %s %s", name, goType)
}

// specRecursiveObject is the plain-object shape of issue #2542: a tree whose
// array items compose the tree type itself with an extra field.
const specRecursiveObject = `openapi: 3.0.0
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    Node:
      type: object
      properties:
        children:
          type: array
          items:
            allOf:
              - $ref: '#/components/schemas/Node'
              - type: object
                properties:
                  extra:
                    type: string
`

// TestMergeSchemasRecursiveObjectAllOf covers the object-recursion variant of
// issue #2542. Generation used to overflow the stack; the composition now
// becomes a named Go type that refers back to itself through the slice, which
// is how the generator represents every other allOf composition — flattened,
// concrete, no json.RawMessage in sight.
func TestMergeSchemasRecursiveObjectAllOf(t *testing.T) {
	code := generateSpec(t, specRecursiveObject)

	assert.Contains(t, code, "type Node_Children_Item struct {")
	// Node ∧ {extra}: the referenced schema's own fields are merged in, and
	// the recursion closes on the composed type, not on Node — grandchildren
	// carry `extra` too, which is what the spec says.
	assertField(t, code, "Children", "*[]Node_Children_Item")
	assertField(t, code, "Extra", "*string")
	assert.NotContains(t, code, "union json.RawMessage",
		"an allOf composition must not be represented as a union")
	assert.NotContains(t, code, "AsNode(")
}

// TestMergeSchemasRecursiveAnyOfAllOf reproduces the exact schema from issue
// #2542. Node is genuinely an anyOf, so the composed type is still a union —
// but over Node's own branches, regenerated, exactly as a non-recursive
// allOf over a union already generates. The recursion closes on the composed
// item type.
func TestMergeSchemasRecursiveAnyOfAllOf(t *testing.T) {
	const spec = `openapi: 3.0.0
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    Node:
      anyOf:
        - type: object
          properties:
            leaf:
              type: string
        - type: object
          properties:
            children:
              type: array
              items:
                allOf:
                  - $ref: '#/components/schemas/Node'
                  - type: object
                    properties:
                      extra:
                        type: string
`

	code := generateSpec(t, spec)

	assert.Contains(t, code, "type Node_1_Children_Item struct {")
	assertField(t, code, "Extra", "*string")
	// The branches of the composed union are Node's branches regenerated at
	// this position, and the recursive one points back at the composed type.
	assert.Contains(t, code, "type Node1Children1 struct {")
	assertField(t, code, "Children", "*[]Node_1_Children_Item")
	assert.Contains(t, code, "func (t Node_1_Children_Item) AsNode1Children1() (Node1Children1, error)")
}

// TestMergeSchemasNestedAllOfSelfRef covers the self-$ref hiding behind a
// nested allOf member, which escapes a guard that only inspects the members
// of the allOf it was handed.
func TestMergeSchemasNestedAllOfSelfRef(t *testing.T) {
	const spec = `openapi: 3.0.0
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    Node:
      type: object
      properties:
        children:
          type: array
          items:
            allOf:
              - allOf:
                  - $ref: '#/components/schemas/Node'
              - type: object
                properties:
                  extra:
                    type: string
`

	code := generateSpec(t, spec)
	assert.Contains(t, code, "type Node_Children_Item struct {")
	assertField(t, code, "Children", "*[]Node_Children_Item")
	assertField(t, code, "Extra", "*string")
}

// TestMergeSchemasRecursionThroughAdditionalProperties covers a cycle that
// closes through additionalProperties. items and additionalProperties reuse
// their parent's path, so a guard keyed on path depth misses this one even
// though it sits at the component root.
func TestMergeSchemasRecursionThroughAdditionalProperties(t *testing.T) {
	const spec = `openapi: 3.0.0
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    Node:
      type: object
      additionalProperties:
        allOf:
          - $ref: '#/components/schemas/Node'
          - type: object
            properties:
              extra:
                type: string
`

	code := generateSpec(t, spec)
	assert.Contains(t, code, "Node_AdditionalProperties")
	assert.Contains(t, code, "map[string]Node_AdditionalProperties")
}

// TestMergeSchemasMutualRecursion covers a cycle that never passes through
// the component being generated: generating A walks into B and C, which
// compose each other. A guard keyed on "refers to the schema at path[0]"
// never fires here.
func TestMergeSchemasMutualRecursion(t *testing.T) {
	const spec = `openapi: 3.0.0
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    A:
      type: object
      properties:
        b:
          allOf:
            - $ref: '#/components/schemas/B'
            - type: object
              properties:
                tag:
                  type: string
    B:
      type: object
      properties:
        c:
          allOf:
            - $ref: '#/components/schemas/C'
            - type: object
              properties:
                x:
                  type: string
    C:
      type: object
      properties:
        back:
          allOf:
            - $ref: '#/components/schemas/B'
            - type: object
              properties:
                y:
                  type: string
`

	code := generateSpec(t, spec)
	for _, want := range []string{"type A struct {", "type B struct {", "type C struct {"} {
		assert.Contains(t, code, want)
	}
	// The cycle has to close on a named type rather than unrolling: the type
	// generated for B.c contains a field whose type is itself.
	assert.Regexp(t, `type B_C struct \{(.|\n)*\*B_C`, code)
	assert.NotContains(t, code, "union json.RawMessage")
}

// TestMergeSchemasRecursionViaOtherComponent covers a cycle reached from a
// component that is not part of it: Wrapper inlines Node's body, and it is
// that inlined body that re-enters itself.
func TestMergeSchemasRecursionViaOtherComponent(t *testing.T) {
	const spec = `openapi: 3.0.0
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    Node:
      type: object
      properties:
        children:
          type: array
          items:
            allOf:
              - $ref: '#/components/schemas/Node'
              - type: object
                properties:
                  extra:
                    type: string
    Wrapper:
      type: object
      properties:
        n:
          allOf:
            - $ref: '#/components/schemas/Node'
            - type: object
              properties:
                extra:
                  type: string
`

	code := generateSpec(t, spec)
	assert.Contains(t, code, "type Wrapper struct {")
	assert.Contains(t, code, "type Node_Children_Item struct {")
}

// TestMergeSchemasValueRecursionGeneratesUncompilableGo pins a deliberate
// choice. A composition that refers to itself with no pointer, slice or map
// in between describes a Go value that contains itself. The generator emits
// what the spec declares and leaves the objection to the compiler, which says
// `invalid recursive type: T_Child refers to itself` and points at the line.
// Detecting it here would mean predicting, before generating the body, what
// the field rendering will do with it — which cannot be done for a
// SkipOptionalPointer that the merge itself produces (the #1957 decorator
// idiom). If this test starts failing, that trade-off is being revisited.
func TestMergeSchemasValueRecursionGeneratesUncompilableGo(t *testing.T) {
	const spec = `openapi: 3.0.0
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    T:
      type: object
      required: [child]
      properties:
        child:
          allOf:
            - $ref: '#/components/schemas/T'
            - type: object
              properties:
                w:
                  type: integer
`

	code := generateSpec(t, spec)
	// Generation terminates rather than overflowing the stack (issue #2542).
	assertField(t, code, "Child", "T_Child")
	assert.Contains(t, code, "type T_Child struct {")
}

// TestMergeSchemasValueRecursionOptionalIsFine is the same shape with the
// property left optional, which renders as a pointer and so terminates.
func TestMergeSchemasValueRecursionOptionalIsFine(t *testing.T) {
	const spec = `openapi: 3.0.0
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    T:
      type: object
      properties:
        child:
          allOf:
            - $ref: '#/components/schemas/T'
            - type: object
              properties:
                w:
                  type: integer
`

	code := generateSpec(t, spec)
	assertField(t, code, "Child", "*T_Child")
	assertField(t, code, "W", "*int")
}

// TestMergeSchemasNonRecursiveAllOfUnaffected pins the property this change
// depends on for backwards compatibility: a composition nothing refers back
// to is still emitted as the inline anonymous struct it always was, with no
// named type invented for it.
func TestMergeSchemasNonRecursiveAllOfUnaffected(t *testing.T) {
	const spec = `openapi: 3.0.0
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    Base:
      type: object
      properties:
        name:
          type: string
    Holder:
      type: object
      properties:
        list:
          type: array
          items:
            allOf:
              - $ref: '#/components/schemas/Base'
              - type: object
                properties:
                  weight:
                    type: integer
`

	code := generateSpec(t, spec)
	assert.Contains(t, code, "List *[]struct {")
	assert.NotContains(t, code, "Holder_List_Item",
		"a non-recursive composition must not gain a named type")
	assert.Equal(t, 1, strings.Count(code, "type Holder struct {"))
}

// TestMergeSchemasRecursiveWithAnonymousSchemaTypes covers the interaction
// with generate-types-for-anonymous-schemas, which hoists the merged body
// under its own path. The name a recursive member was handed has to be the
// one that ends up defined, or the generated code refers to a type that does
// not exist.
func TestMergeSchemasRecursiveWithAnonymousSchemaTypes(t *testing.T) {
	anon := func(c *Configuration) { c.OutputOptions.GenerateTypesForAnonymousSchemas = true }

	t.Run("array items", func(t *testing.T) {
		code := generateSpec(t, specRecursiveObject, anon)
		assertDefinesWhatItReferences(t, code)
	})

	t.Run("additionalProperties", func(t *testing.T) {
		const spec = `openapi: 3.0.0
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    Node:
      type: object
      properties:
        kids:
          type: object
          additionalProperties:
            allOf:
              - $ref: '#/components/schemas/Node'
              - type: object
                properties:
                  extra:
                    type: string
`
		code := generateSpec(t, spec, anon)
		assertDefinesWhatItReferences(t, code)
	})
}

// assertDefinesWhatItReferences checks that every generated type referenced
// from a struct field is also declared, which is the cheapest stand-in for
// "the output compiles".
func assertDefinesWhatItReferences(t *testing.T, code string) {
	t.Helper()
	declared := map[string]bool{}
	for _, m := range regexp.MustCompile(`(?m)^type (\w+) `).FindAllStringSubmatch(code, -1) {
		declared[m[1]] = true
	}
	referenced := regexp.MustCompile(`\*?\[?\]?\*?(Node\w*)\s+`+"`json:").FindAllStringSubmatch(code, -1)
	require.NotEmpty(t, referenced, "expected the output to reference a generated type")
	for _, m := range referenced {
		assert.True(t, declared[m[1]], "field refers to %s, which is never declared:\n%s", m[1], code)
	}
}

// TestMergeSchemasRecursionUnderOldMergeSchemas pins that the legacy merge
// path is unaffected. It embeds $ref members instead of inlining them, so it
// never recursed on this shape. (It once bypassed the state this fix threads;
// it now shares it, see merge_schemas_v1_test.go, which changes nothing
// here.)
//
// This test records how schema-merging-behavior v1 behaves. A bug fix may
// change it, but it must not be changed to accept a regression, and a commit
// that changes it must say why.
func TestMergeSchemasRecursionUnderOldMergeSchemas(t *testing.T) {
	code := generateSpec(t, specRecursiveObject, func(c *Configuration) {
		c.Compatibility.OldMergeSchemas = true
	})
	assert.Contains(t, code, "type Node struct {")
}

const remoteRecursiveCommonSpec = `openapi: 3.0.3
info: {title: common, version: "1.0.0"}
paths: {}
components:
  schemas:
    Tree:
      type: object
      properties:
        name:
          type: string
        kids:
          type: array
          items:
            $ref: '#/components/schemas/Tree'
`

const remoteRecursiveUserSpec = `openapi: 3.0.3
info: {title: user, version: "1.0.0"}
paths: {}
components:
  schemas:
    A:
      allOf:
        - $ref: './common.yml#/components/schemas/Tree'
        - type: object
          properties:
            extra:
              type: string
    B:
      allOf:
        - $ref: './common.yml#/components/schemas/Tree'
        - type: object
          properties:
            other:
              type: string
`

// TestMergeSchemasRemoteRecursiveSchemaFlattenedTwice is issue #2557 through
// the path users hit it on: a self-recursive schema in another document,
// composed via allOf from two components. Flattening it once always worked;
// the second flatten overflowed the stack.
func TestMergeSchemasRemoteRecursiveSchemaFlattenedTwice(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "common.yml"), []byte(remoteRecursiveCommonSpec), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "user.yml"), []byte(remoteRecursiveUserSpec), 0o600))

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadFromFile(filepath.Join(dir, "user.yml"))
	require.NoError(t, err)

	code, err := Generate(swagger, Configuration{
		PackageName:   "user",
		Generate:      GenerateOptions{Models: true},
		OutputOptions: OutputOptions{SkipPrune: true},
		ImportMapping: map[string]string{"./common.yml": "example.com/common"},
	})
	require.NoError(t, err)

	// Both composed types keep the recursive member pointing at the named
	// type in the document it came from, rather than inlining it.
	assert.Equal(t, 2, strings.Count(code, "*[]externalRef0.Tree"),
		"both A and B should reference the external Tree:\n%s", code)
	assert.Contains(t, code, "type A struct {")
	assert.Contains(t, code, "type B struct {")
}

// TestMergeSchemasAllOfOverArrayKeepsItems is the end-to-end case: an allOf over
// an array schema generates a typed slice, not []any.
func TestMergeSchemasAllOfOverArrayKeepsItems(t *testing.T) {
	const spec = `openapi: 3.0.0
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    Item:
      type: object
      properties:
        n:
          type: integer
    Items:
      type: array
      items:
        $ref: '#/components/schemas/Item'
    Wrapped:
      allOf:
        - $ref: '#/components/schemas/Items'
        - minItems: 1
`
	code := generateSpec(t, spec)
	assert.Contains(t, code, "type Wrapped = []Item")
}

// TestAllOfWithOneOfSibling covers a oneOf next to allOf on the same schema,
// which used to be dropped.
func TestAllOfWithOneOfSibling(t *testing.T) {
	const spec = `openapi: 3.0.0
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    Base:
      type: object
      properties:
        id:
          type: string
    Cat:
      type: object
      properties:
        meow:
          type: string
    Dog:
      type: object
      properties:
        bark:
          type: string
    Contact:
      type: object
      properties:
        email:
          type: string
        phone:
          type: string
    Pet:
      allOf:
        - $ref: '#/components/schemas/Base'
      oneOf:
        - $ref: '#/components/schemas/Cat'
        - $ref: '#/components/schemas/Dog'
    OneWayToReach:
      allOf:
        - $ref: '#/components/schemas/Contact'
      oneOf:
        - required: [email]
        - required: [phone]
    NeedEmail:
      required: [email]
    NeedPhone:
      required: [phone]
    OneWayByRef:
      allOf:
        - $ref: '#/components/schemas/Contact'
      oneOf:
        - $ref: '#/components/schemas/NeedEmail'
        - $ref: '#/components/schemas/NeedPhone'
`
	code := generateSpec(t, spec)

	assert.Contains(t, code, "type Pet struct {")
	assert.Contains(t, code, "func (t Pet) AsCat() (Cat, error)")
	assert.Contains(t, code, "func (t Pet) AsDog() (Dog, error)")

	// Branches that only add constraints are not types to choose between;
	// the schema stays what it was, an alias of its allOf member. That holds
	// when the branches are $refs to such schemas, too.
	assert.Contains(t, code, "type OneWayToReach = Contact")
	assert.Contains(t, code, "type OneWayByRef = Contact")
}

// TestAnyOfAndOneOfOnOneSchema is issue #839: inline members of an anyOf and
// a oneOf on the same schema used to get the same names.
func TestAnyOfAndOneOfOnOneSchema(t *testing.T) {
	const spec = `openapi: 3.0.0
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    Orange:
      type: object
      anyOf:
        - required: [a]
          properties:
            a:
              type: string
        - required: [b]
          properties:
            b:
              type: string
      oneOf:
        - required: [c]
          properties:
            c:
              type: string
    Lemon:
      type: object
      oneOf:
        - required: [c]
          properties:
            c:
              type: string
`
	code := generateSpec(t, spec)
	for _, name := range []string{"OrangeAnyOf0", "OrangeAnyOf1", "OrangeOneOf0"} {
		assert.Contains(t, code, "type "+name+" struct {")
		assert.Contains(t, code, "func (t Orange) As"+name+"() ("+name+", error)")
	}
	// A schema with only one of the keywords keeps its names.
	assert.Contains(t, code, "type Lemon0 struct {")
}
