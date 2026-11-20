package codegen

import (
	"regexp"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// A oneOf/anyOf branch that is an array of an inline enum must not give the
// items' enum the same type name as the branch (issue #2640).
func TestUnionArrayBranchOfInlineEnum(t *testing.T) {
	const spec = `
openapi: 3.0.3
info: {title: t, version: "1"}
paths:
  /items:
    post:
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                roles:
                  oneOf:
                    - {type: string}
                    - {type: array, items: {type: string, enum: [admin, viewer]}}
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Item'
components:
  schemas:
    Item:
      type: object
      properties:
        roles:
          anyOf:
            - {type: string}
            - {type: array, items: {type: string, enum: [admin, viewer]}}
`
	swagger, err := openapi3.NewLoader().LoadFromData([]byte(spec))
	require.NoError(t, err)

	code, err := Generate(swagger, Configuration{
		PackageName: "api",
		Generate:    GenerateOptions{Models: true},
	})
	require.NoError(t, err)

	for _, name := range []string{"PostItemsJSONBodyRoles1", "ItemRoles1"} {
		require.Regexp(t, regexp.MustCompile(`type `+name+` = ?\[\]`+name+`Item\n`), code)
		require.Regexp(t, regexp.MustCompile(`type `+name+`Item string\n`), code)
		require.Len(t, regexp.MustCompile(`(?m)^type `+name+` `).FindAllString(code, -1), 1)
	}
}

func TestUnionObjectBranchArrayPropertyInlineEnumName(t *testing.T) {
	const spec = `
openapi: 3.0.3
info: {title: t, version: "1"}
paths: {}
components:
  schemas:
    Choice:
      oneOf:
        - type: object
          properties:
            roles: {type: array, items: {type: string, enum: [admin, viewer]}}
        - {type: string}
`
	swagger, err := openapi3.NewLoader().LoadFromData([]byte(spec))
	require.NoError(t, err)
	code, err := Generate(swagger, Configuration{
		PackageName:   "api",
		Generate:      GenerateOptions{Models: true},
		OutputOptions: OutputOptions{SkipPrune: true},
	})
	require.NoError(t, err)
	require.Regexp(t, `(?m)^type Choice0Roles string$`, code)
	require.Regexp(t, `Roles \*\[\]Choice0Roles `, code)
	require.NotContains(t, code, "Choice0RolesItem")
}

func TestUnionArrayBranchExplicitEnumNameLegacyConstants(t *testing.T) {
	const spec = `
openapi: 3.0.3
info: {title: t, version: "1"}
paths: {}
components:
  schemas:
    Choice:
      oneOf:
        - {type: string}
        - {type: array, items: {type: string, enum: [admin, viewer], x-go-type-name: Role}}
`
	swagger, err := openapi3.NewLoader().LoadFromData([]byte(spec))
	require.NoError(t, err)
	code, err := Generate(swagger, Configuration{
		PackageName:   "api",
		Generate:      GenerateOptions{Models: true},
		OutputOptions: OutputOptions{SkipPrune: true},
		Compatibility: CompatibilityOptions{OldEnumConflicts: true},
	})
	require.NoError(t, err)
	require.Regexp(t, `(?m)^\s*Choice1Admin\s+Role\s*=`, code)
	require.Regexp(t, `(?m)^\s*Choice1Viewer\s+Role\s*=`, code)
	require.NotContains(t, code, "Choice1ItemAdmin")
}

func TestUnionArrayBranchOfOpenAPI31InlineEnums(t *testing.T) {
	for _, tc := range []struct{ name, items string }{
		{"const", "type: string, const: admin"},
		{"oneOf", "oneOf: [{title: Admin, const: admin}, {title: Viewer, const: viewer}]"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spec := `
openapi: 3.1.0
info: {title: t, version: "1"}
paths: {}
components:
  schemas:
    Choice:
      oneOf:
        - {type: string}
        - {type: array, items: {` + tc.items + `}}
`
			swagger, err := openapi3.NewLoader().LoadFromData([]byte(spec))
			require.NoError(t, err)
			code, err := Generate(swagger, Configuration{
				PackageName:   "api",
				Generate:      GenerateOptions{Models: true},
				OutputOptions: OutputOptions{SkipPrune: true},
			})
			require.NoError(t, err)
			require.Regexp(t, `(?m)^type Choice1 = ?\[\]Choice1Item$`, code)
			require.Regexp(t, `(?m)^type Choice1Item string$`, code)
		})
	}
}

// An inline enum reached from a branch through schemas that share the branch's
// path (items, additionalProperties, and an allOf around the enum) is named
// after its own position, whatever the nesting.
func TestUnionBranchNestedInlineEnumName(t *testing.T) {
	const enum = `{type: string, enum: [admin, viewer]}`
	for _, tc := range []struct{ name, branch, branchType, enumType string }{
		{"array of arrays", `{type: array, items: {type: array, items: ` + enum + `}}`,
			`type Choice1 = [][]Choice1Item`, `Choice1Item`},
		{"allOf items", `{type: array, items: {allOf: [` + enum + `]}}`,
			`type Choice1 = []Choice1Item`, `Choice1Item`},
		{"map", `{type: object, additionalProperties: ` + enum + `}`,
			`type Choice1 map[string]Choice1AdditionalProperties`, `Choice1AdditionalProperties`},
		{"object with additionalProperties", `{type: object, properties: {a: {type: string}}, additionalProperties: ` + enum + `}`,
			`type Choice1 struct {`, `Choice1AdditionalProperties`},
		{"array of maps", `{type: array, items: {type: object, additionalProperties: ` + enum + `}}`,
			`type Choice1 = []map[string]Choice1AdditionalProperties`, `Choice1AdditionalProperties`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spec := `
openapi: 3.0.3
info: {title: t, version: "1"}
paths: {}
components:
  schemas:
    Choice:
      oneOf:
        - {type: string}
        - ` + tc.branch + `
`
			swagger, err := openapi3.NewLoader().LoadFromData([]byte(spec))
			require.NoError(t, err)
			code, err := Generate(swagger, Configuration{
				PackageName:   "api",
				Generate:      GenerateOptions{Models: true},
				OutputOptions: OutputOptions{SkipPrune: true},
			})
			require.NoError(t, err)
			require.Contains(t, code, tc.branchType+"\n")
			require.Contains(t, code, "type "+tc.enumType+" string\n")
		})
	}
}

// An enum that is itself the branch has the branch's name, and defines it.
func TestUnionEnumBranchName(t *testing.T) {
	const spec = `
openapi: 3.0.3
info: {title: t, version: "1"}
paths: {}
components:
  schemas:
    Choice:
      oneOf:
        - {type: string, enum: [admin, viewer]}
        - {type: integer}
`
	swagger, err := openapi3.NewLoader().LoadFromData([]byte(spec))
	require.NoError(t, err)
	code, err := Generate(swagger, Configuration{
		PackageName:   "api",
		Generate:      GenerateOptions{Models: true},
		OutputOptions: OutputOptions{SkipPrune: true},
	})
	require.NoError(t, err)
	require.Regexp(t, `(?m)^type Choice0 string$`, code)
}
