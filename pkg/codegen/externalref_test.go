package codegen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQualifyExternalTypeNames(t *testing.T) {
	for _, tt := range []struct {
		name     string
		typeExpr string
		expected string
	}{
		{name: "model", typeExpr: "Foo", expected: "ext.Foo"},
		{name: "any", typeExpr: "any", expected: "any"},
		{name: "empty interface", typeExpr: "interface{}", expected: "interface{}"},
		{name: "builtin", typeExpr: "string", expected: "string"},
		{name: "error", typeExpr: "error", expected: "error"},
		{name: "pointer to builtin", typeExpr: "*string", expected: "*string"},
		{name: "pointer to model", typeExpr: "*Foo", expected: "*ext.Foo"},
		{name: "slice of models", typeExpr: "[]Foo", expected: "[]ext.Foo"},
		{name: "array of models", typeExpr: "[4]Foo", expected: "[4]ext.Foo"},
		{name: "slice of builtins", typeExpr: "[]string", expected: "[]string"},
		{name: "map of models", typeExpr: "map[string]Foo", expected: "map[string]ext.Foo"},
		{name: "map keyed by model", typeExpr: "map[Key]*Foo", expected: "map[ext.Key]*ext.Foo"},
		{name: "qualified", typeExpr: "time.Time", expected: "time.Time"},
		{name: "slice of qualified", typeExpr: "[]openapi_types.UUID", expected: "[]openapi_types.UUID"},
		{name: "other external package", typeExpr: "externalRef1.Bar", expected: "externalRef1.Bar"},
		{name: "generic", typeExpr: "Page[Foo, int]", expected: "ext.Page[ext.Foo, int]"},
		{name: "interface methods", typeExpr: "interface{ M(Foo) Bar }", expected: "interface{ M(ext.Foo) ext.Bar }"},
		{
			name:     "struct literal",
			typeExpr: "struct {\n\t// F refers to a Foo\n\tF *Foo `json:\"f,omitempty\"`\n\tS string `json:\"s\"`\n}",
			expected: "struct {\n\t// F refers to a Foo\n\tF *ext.Foo `json:\"f,omitempty\"`\n\tS string `json:\"s\"`\n}",
		},
		{name: "not a type", typeExpr: "not a type!", expected: "not a type!"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, qualifyExternalTypeNames(tt.typeExpr, "ext"))
		})
	}
}

func TestEnsureExternalRefsInSchema(t *testing.T) {
	old := globalState.importMapping
	globalState.importMapping = constructImportMapping(map[string]string{"./ext.yaml": "example.com/ext"})
	defer func() { globalState.importMapping = old }()
	pkg := globalState.importMapping["./ext.yaml"].Name

	const ref = "./ext.yaml#/components/responses/R"
	for _, tt := range []struct {
		name   string
		schema Schema
		want   string
		isRef  bool
	}{
		{name: "model", schema: Schema{GoType: "Foo"}, want: pkg + ".Foo", isRef: true},
		{name: "builtin", schema: Schema{GoType: "any"}, want: "any"},
		{name: "pointer", schema: Schema{GoType: "*string"}, want: "*string"},
		{name: "slice of models", schema: Schema{GoType: "[]Foo"}, want: "[]" + pkg + ".Foo"},
		{name: "qualified", schema: Schema{GoType: "time.Time"}, want: "time.Time"},
		{name: "struct literal", schema: Schema{GoType: "struct {\n\tF Foo\n}"}, want: "struct {\n\tF " + pkg + ".Foo\n}"},
		// A named type the schema already refers to keeps that name when
		// its Go type isn't a plain model name.
		{name: "qualified ref", schema: Schema{GoType: "string", RefType: pkg + ".OpJSONBody"}, want: pkg + ".OpJSONBody", isRef: true},
		{name: "local ref", schema: Schema{GoType: "string", RefType: "OpParamsQ"}, want: pkg + ".OpParamsQ", isRef: true},
		// These are qualified as they always were.
		{name: "model with ref", schema: Schema{GoType: "Foo", RefType: pkg + ".OpJSONBody"}, want: pkg + ".Foo", isRef: true},
		{name: "struct literal with ref", schema: Schema{GoType: "struct {\n\tF Foo\n}", RefType: "OpBody"}, want: "OpBody", isRef: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.schema
			ensureExternalRefsInSchema(&schema, ref)
			assert.Equal(t, tt.want, schema.TypeDecl())
			assert.Equal(t, tt.isRef, schema.IsRef())
		})
	}

	t.Run("response headers", func(t *testing.T) {
		defs := []ResponseDefinition{{Headers: []ResponseHeaderDefinition{
			{Name: "X-Model", Schema: Schema{GoType: "Foo"}},
			{Name: "X-Builtin", Schema: Schema{GoType: "string"}},
			// A header $ref into a third document is already qualified.
			{Name: "X-Third", Schema: Schema{GoType: "Bar", RefType: "externalRef1.Bar"}},
		}}}
		ensureExternalRefsInResponseDefinitions(&defs, ref)
		assert.Equal(t, pkg+".Foo", defs[0].Headers[0].Schema.TypeDecl())
		assert.Equal(t, "string", defs[0].Headers[1].Schema.TypeDecl())
		assert.Equal(t, "externalRef1.Bar", defs[0].Headers[2].Schema.TypeDecl())
	})

	t.Run("local ref is left alone", func(t *testing.T) {
		schema := Schema{GoType: "any"}
		ensureExternalRefsInSchema(&schema, "#/components/responses/R")
		assert.Equal(t, Schema{GoType: "any"}, schema)
	})
}

const externalTypesCommonSpec = `
openapi: "3.0.4"
info: {title: common, version: "1"}
paths:
  /things/{id}:
    post:
      operationId: updateThing
      parameters:
        - {name: id, in: path, required: true, schema: {type: string}}
        - {name: limit, in: query, schema: {type: integer}}
      requestBody:
        required: true
        content:
          application/json:
            schema: {type: array, items: {$ref: "#/components/schemas/Thing"}}
      responses:
        "200":
          description: ok
          headers:
            X-Thing-ID:
              required: true
              schema: {$ref: "#/components/schemas/ThingID"}
          content:
            application/json:
              schema: {type: array, items: {$ref: "#/components/schemas/Thing"}}
        default:
          description: untyped
          content:
            application/json:
              schema: {}
components:
  schemas:
    ThingID: {type: string}
    Thing:
      type: object
      properties:
        id: {$ref: "#/components/schemas/ThingID"}
  responses:
    Untyped:
      description: untyped
      content:
        application/json:
          schema: {}
    Pointer:
      description: pointer
      content:
        application/json:
          schema: {x-go-type: "*string"}
    Things:
      description: things
      content:
        application/json:
          schema: {type: array, items: {$ref: "#/components/schemas/Thing"}}
`

const externalTypesAPISpec = `
openapi: "3.0.4"
info: {title: api, version: "1"}
paths:
  /things/{id}:
    $ref: "./common.yaml#/paths/~1things~1{id}"
  /untyped:
    get:
      operationId: getUntyped
      responses:
        default: {$ref: "./common.yaml#/components/responses/Untyped"}
  /pointer:
    get:
      operationId: getPointer
      responses:
        default: {$ref: "./common.yaml#/components/responses/Pointer"}
  /things:
    get:
      operationId: listThings
      responses:
        default: {$ref: "./common.yaml#/components/responses/Things"}
`

// Types that come from an import-mapped document but are not named types of
// it are qualified only where they name the document's models.
func TestExternalTypeQualification(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "common.yaml"), []byte(externalTypesCommonSpec), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "api.yaml"), []byte(externalTypesAPISpec), 0o600))

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadFromFile(filepath.Join(dir, "api.yaml"))
	require.NoError(t, err)

	code, err := Generate(swagger, Configuration{
		PackageName: "api",
		Generate: GenerateOptions{
			StdHTTPServer: true,
			Strict:        true,
			Models:        true,
			Client:        true,
		},
		ImportMapping: map[string]string{"./common.yaml": "example.com/common"},
	})
	require.NoError(t, err)

	// A reusable response reached from a non-fixed status code.
	assertField(t, code, "Body", "any")
	assertField(t, code, "Body", "*string")
	assertField(t, code, "Body", "[]externalRef0.Thing")
	assert.NotContains(t, code, "externalRef0.any")
	assert.NotContains(t, code, "externalRef0.*string")

	// An operation of a path item reused by $ref.
	assertField(t, code, "Limit", "*int")
	assertField(t, code, "Id", "string")
	assert.Contains(t, code, "type UpdateThingJSONRequestBody = externalRef0.UpdateThingJSONBody")
	assertField(t, code, "XThingID", "externalRef0.ThingID")
	assertField(t, code, "JSON200", "*[]externalRef0.Thing")
	assertField(t, code, "JSONDefault", "*any")
	assert.NotContains(t, code, "externalRef0.string")
	assert.NotContains(t, code, "externalRef0.int")
}
