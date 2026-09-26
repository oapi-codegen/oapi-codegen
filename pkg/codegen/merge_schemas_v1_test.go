package codegen

// Tests of schema-merging-behavior v1 on recursive allOf compositions.
//
// v1 embeds a $ref member as a Go struct and inlines the fields of an inline
// member, without dereferencing anything, so the shapes of issue #2542 never
// recursed under it: the $ref back into the composition is embedded as the
// referenced type. What did recurse without end was a whole-document $ref
// (`$ref: ./tree.yaml`), whose value is generated inline because it names no
// Go type: v1 used to generate each allOf member with a fresh genContext, so
// the recursion guard of generateAllOfV2 never saw the enclosing frame.
//
// These tests record how v1 behaves. A bug fix may change them, but they must
// not be changed to accept a regression, and a commit that changes them must
// say why.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withSchemaMergingV1 selects schema-merging-behavior v1 for generateSpec.
func withSchemaMergingV1(c *Configuration) {
	c.Compatibility.SchemaMergingBehavior = SchemaMergingV1
}

// TestMergeSchemasV1RecursiveObjectAllOf pins what v1 generates for the
// object shape of issue #2542: the $ref member is embedded, the sibling's
// fields are inlined, and the recursion closes on Node through the embedded
// struct, inside the anonymous item type v1 has always generated.
func TestMergeSchemasV1RecursiveObjectAllOf(t *testing.T) {
	code := generateSpec(t, specRecursiveObject, withSchemaMergingV1)

	assert.Contains(t, code, "type Node struct {")
	assert.Contains(t, code, "Children *[]struct {")
	assert.Contains(t, code, "Node `yaml:\",inline\"`")
	// Nested one level deeper than assertField expects, in the anonymous
	// item struct.
	assert.Regexp(t, "\n\t\tExtra\\s+\\*string\\s", code)
	assert.NotContains(t, code, "Node_Children_Item",
		"v1 names no type for a composition nothing refers back to")
}

// TestMergeSchemasV1RecursiveShapesTerminate runs the other shapes of issue
// #2542 under v1. None of them ever overflowed under v1, and none of them
// may start to: the $ref back into the composition is embedded, and the
// sibling's field is inlined next to it.
func TestMergeSchemasV1RecursiveShapesTerminate(t *testing.T) {
	const header = `openapi: 3.0.0
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
`
	tests := []struct {
		name     string
		schemas  string
		embedded string
	}{
		{
			// The shape from the issue itself: Node is an anyOf.
			name: "anyOf",
			schemas: `    Node:
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
`,
			embedded: "Node `yaml:\",inline\"`",
		},
		{
			// The self-$ref sits inside an allOf nested inside another
			// allOf member.
			name: "transitive",
			schemas: `    Node:
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
`,
			// v1 inlines the fields of the inline member, and an allOf
			// of a single $ref has none of its own, so the reference is
			// not embedded. That is v1's behavior for a nested allOf, not
			// a regression of this change: the output is byte-identical.
			embedded: "// Embedded fields due to inline allOf schema",
		},
		{
			// A cycle that never reaches the component being generated.
			name: "mutual",
			schemas: `    MutualA:
      type: object
      properties:
        b:
          allOf:
            - $ref: '#/components/schemas/MutualB'
            - type: object
              properties:
                tag:
                  type: string
    MutualB:
      type: object
      properties:
        c:
          allOf:
            - $ref: '#/components/schemas/MutualC'
            - type: object
              properties:
                x:
                  type: string
    MutualC:
      type: object
      properties:
        back:
          allOf:
            - $ref: '#/components/schemas/MutualB'
            - type: object
              properties:
                y:
                  type: string
`,
			embedded: "MutualB `yaml:\",inline\"`",
		},
		{
			// The cycle closes through a map at the component root.
			name: "additionalProperties",
			schemas: `    NodeMap:
      type: object
      additionalProperties:
        allOf:
          - $ref: '#/components/schemas/NodeMap'
          - type: object
            properties:
              extra:
                type: string
`,
			embedded: "NodeMap `yaml:\",inline\"`",
		},
		{
			// A required property composes its own type: the generator
			// terminates, and the compiler then rejects the value cycle
			// (T contains a T), as it does under v2.
			name: "requiredValue",
			schemas: `    T:
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
`,
			embedded: "T `yaml:\",inline\"`",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := generateSpec(t, header+tt.schemas, withSchemaMergingV1)
			assert.Contains(t, code, tt.embedded)
			assert.NotContains(t, code, "_Item ", "v1 names no type for these compositions")
		})
	}
}

// specWholeDocumentTree is a schema document, not an OpenAPI document: a
// tree whose array items compose the document itself, by a $ref to the whole
// file, with an extra field. A $ref to a whole document names no Go type, so
// its value is generated inline wherever it appears, which is where v1 lost
// the recursion state and overflowed the stack. v2 already terminated.
const specWholeDocumentTree = `type: object
properties:
  name:
    type: string
  children:
    type: array
    items:
      allOf:
        - $ref: './tree.yaml'
        - type: object
          properties:
            extra:
              type: string
`

const specWholeDocumentUser = `openapi: 3.0.0
info: {title: repro, version: "1.0.0"}
paths: {}
components:
  schemas:
    Tree:
      $ref: './tree.yaml'
`

// TestMergeSchemasWholeDocumentRecursiveAllOf: the composition refers back to
// itself through a whole-document $ref, so it has to become a named type
// that refers to itself through the slice, under v1 as under v2. v1 used to
// overflow the stack on it.
func TestMergeSchemasWholeDocumentRecursiveAllOf(t *testing.T) {
	for _, version := range []string{SchemaMergingV1, SchemaMergingV2} {
		t.Run(version, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, "tree.yaml"), []byte(specWholeDocumentTree), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "user.yaml"), []byte(specWholeDocumentUser), 0o600))

			loader := openapi3.NewLoader()
			loader.IsExternalRefsAllowed = true
			swagger, err := loader.LoadFromFile(filepath.Join(dir, "user.yaml"))
			require.NoError(t, err)

			code, err := Generate(swagger, Configuration{
				PackageName:   "repro",
				Generate:      GenerateOptions{Models: true},
				OutputOptions: OutputOptions{SkipPrune: true},
				Compatibility: CompatibilityOptions{SchemaMergingBehavior: version},
			})
			require.NoError(t, err)

			assert.Contains(t, code, "type Tree struct {")
			assert.Contains(t, code, "type Tree_Children_Item struct {")
			assertField(t, code, "Children", "*[]Tree_Children_Item")
			assertField(t, code, "Name", "*string")
			assertField(t, code, "Extra", "*string")
		})
	}
}
