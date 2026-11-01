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
