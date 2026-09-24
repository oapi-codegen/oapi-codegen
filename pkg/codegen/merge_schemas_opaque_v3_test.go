package codegen

// These tests cover how schema-merging-behavior v3 treats allOf members whose
// schema the merge can't read: schemas in another document, and schemas that
// x-go-type replaces. Such a member can be annotated, but not merged.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const opaqueCommonSpec = `openapi: 3.0.3
info: {title: common, version: "1.0.0"}
paths: {}
components:
  schemas:
    User:
      type: object
      properties:
        name: {type: string}
        friend: {$ref: '#/components/schemas/User'}
`

// opaqueUserDocument is a document that is a schema, for whole-document refs.
const opaqueUserDocument = `type: object
properties:
  name: {type: string}
`

// generateWithCommonV3 generates models under v3 from spec, which may refer
// to ./common.yaml (opaqueCommonSpec), imported from example.com/common, and
// to the whole document ./user.yaml (opaqueUserDocument).
func generateWithCommonV3(t *testing.T, spec string, opts ...func(*Configuration)) (string, error) {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "common.yaml"), []byte(opaqueCommonSpec), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "user.yaml"), []byte(opaqueUserDocument), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "api.yaml"), []byte(spec), 0o600))

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	swagger, err := loader.LoadFromFile(filepath.Join(dir, "api.yaml"))
	require.NoError(t, err)

	cfg := Configuration{
		PackageName:   "api",
		Generate:      GenerateOptions{Models: true},
		OutputOptions: OutputOptions{SkipPrune: true},
		ImportMapping: map[string]string{"./common.yaml": "example.com/common"},
		Compatibility: CompatibilityOptions{SchemaMergingBehavior: schemaMergingV3},
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return Generate(swagger, cfg)
}

// withV3 selects schema-merging-behavior v3 for generateSpec.
func withV3(c *Configuration) { c.Compatibility.SchemaMergingBehavior = schemaMergingV3 }

const opaqueSpecHeader = `openapi: 3.0.3
info: {title: api, version: "1.0.0"}
paths: {}
components:
  schemas:
`

const opaqueSpecHeader31 = `openapi: 3.1.0
info: {title: api, version: "1.0.0"}
paths: {}
components:
  schemas:
`

// TestExternalAllOfMemberAnnotated: an external schema plus members that only
// annotate it is that schema's type, never a copy of its body.
func TestExternalAllOfMemberAnnotated(t *testing.T) {
	code, err := generateWithCommonV3(t, opaqueSpecHeader+`
    Decorated:
      allOf:
        - $ref: './common.yaml#/components/schemas/User'
        - description: A user.
          nullable: true
    Holder:
      type: object
      required: [required_user]
      properties:
        user:
          allOf:
            - $ref: './common.yaml#/components/schemas/User'
            - nullable: true
              readOnly: true
              maxProperties: 3
        required_user:
          allOf:
            - $ref: './common.yaml#/components/schemas/User'
            - description: Always there.
        no_pointer:
          allOf:
            - $ref: './common.yaml#/components/schemas/User'
            - x-go-type-skip-optional-pointer: true
`)
	require.NoError(t, err)
	assert.Contains(t, code, "type Decorated = externalRef0.User")
	assertField(t, code, "User", "*externalRef0.User")
	assertField(t, code, "RequiredUser", "externalRef0.User")
	assertField(t, code, "NoPointer", "externalRef0.User")
	assert.NotContains(t, code, "Friend", "the external schema's body must not be copied")
}

func TestExternalAllOfMemberMergedIsAnError(t *testing.T) {
	for name, tc := range map[string]struct {
		spec string
		with string
	}{
		"inline properties": {
			spec: `
    Enriched:
      allOf:
        - $ref: './common.yaml#/components/schemas/User'
        - type: object
          properties:
            extra: {type: string}
`,
			with: "with an inline schema with type, properties",
		},
		"the parent's own properties": {
			spec: `
    Enriched:
      type: object
      properties:
        extra: {type: string}
      allOf:
        - $ref: './common.yaml#/components/schemas/User'
`,
			with: "with the schema's own type, properties",
		},
		"a local $ref (#2470)": {
			spec: `
    Local:
      type: object
      properties:
        viewed: {type: boolean}
    Enriched:
      allOf:
        - $ref: './common.yaml#/components/schemas/User'
        - $ref: '#/components/schemas/Local'
`,
			with: "with #/components/schemas/Local",
		},
		"required only": {
			spec: `
    Enriched:
      allOf:
        - $ref: './common.yaml#/components/schemas/User'
        - required: [name]
`,
			with: "with an inline schema with required",
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := generateWithCommonV3(t, opaqueSpecHeader+tc.spec)
			require.Error(t, err)
			msg := err.Error()
			assert.Contains(t, msg, "allOf can't merge ./common.yaml#/components/schemas/User "+tc.with)
			assert.Contains(t, msg, "a reference to another document can't be merged with other allOf members")
			assert.Contains(t, msg, "Define the schema in this document")
			assert.NotContains(t, msg, "schema-merging-behavior")
		})
	}
}

// TestExternalAllOfMemberThroughLocalComposition: a local schema that is an
// annotated external schema is just as opaque.
func TestExternalAllOfMemberThroughLocalComposition(t *testing.T) {
	const decorated = `
    Decorated:
      allOf:
        - $ref: './common.yaml#/components/schemas/User'
        - description: A user.
`
	code, err := generateWithCommonV3(t, opaqueSpecHeader+decorated+`
    Nullable:
      allOf:
        - $ref: '#/components/schemas/Decorated'
        - nullable: true
`)
	require.NoError(t, err)
	assert.Contains(t, code, "type Decorated = externalRef0.User")
	assert.Contains(t, code, "type Nullable = Decorated")

	_, err = generateWithCommonV3(t, opaqueSpecHeader+decorated+`
    Enriched:
      allOf:
        - $ref: '#/components/schemas/Decorated'
        - properties:
            extra: {type: string}
`)
	require.Error(t, err)
	assert.Contains(t, err.Error(),
		"allOf can't merge #/components/schemas/Decorated (whose allOf includes ./common.yaml#/components/schemas/User) "+
			"with an inline schema with properties")
}

// TestExternalAllOfMemberThroughLocalAlias: a component that is only a $ref
// to another document is that document's schema, however many such
// components the $ref goes through.
func TestExternalAllOfMemberThroughLocalAlias(t *testing.T) {
	const aliases = `
    LocalUser:
      $ref: './common.yaml#/components/schemas/User'
    AliasOfAlias:
      $ref: '#/components/schemas/LocalUser'
`
	code, err := generateWithCommonV3(t, opaqueSpecHeader+aliases+`
    Nullable:
      allOf:
        - $ref: '#/components/schemas/AliasOfAlias'
        - nullable: true
    Holder:
      type: object
      properties:
        owner:
          allOf:
            - $ref: '#/components/schemas/LocalUser'
            - description: The owner.
`)
	require.NoError(t, err)
	assert.Contains(t, code, "type LocalUser = externalRef0.User")
	assert.Contains(t, code, "type Nullable = AliasOfAlias")
	assertField(t, code, "Owner", "*LocalUser")

	for _, alias := range []string{"LocalUser", "AliasOfAlias"} {
		t.Run(alias, func(t *testing.T) {
			_, err := generateWithCommonV3(t, opaqueSpecHeader+aliases+`
    Enriched:
      allOf:
        - $ref: '#/components/schemas/`+alias+`'
        - properties:
            extra: {type: string}
`)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "allOf can't merge #/components/schemas/"+alias+
				" (a $ref to another document) with an inline schema with properties: "+
				"a reference to another document can't be merged")
		})
	}
}

// TestExternalAllOfMemberFlattened: the merge doesn't read an external schema
// however deep it finds one, and the error names the member that has it.
func TestExternalAllOfMemberFlattened(t *testing.T) {
	_, err := generateWithCommonV3(t, opaqueSpecHeader+`
    Enriched:
      allOf:
        - allOf:
            - $ref: './common.yaml#/components/schemas/User'
            - properties:
                inner: {type: string}
        - properties:
            outer: {type: string}
`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "allOf can't merge an inline schema with allOf "+
		"(whose allOf includes ./common.yaml#/components/schemas/User) with an inline schema with properties")

	_, err = generateWithCommonV3(t, opaqueSpecHeader+`
    Enriched:
      allOf:
        - $ref: '#/components/schemas/Inner'
        - properties:
            outer: {type: string}
    Inner:
      type: object
      properties:
        inner: {type: string}
      allOf:
        - $ref: './common.yaml#/components/schemas/User'
`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error converting Schema Enriched to Go type")
	assert.Contains(t, err.Error(), "allOf can't merge #/components/schemas/Inner "+
		"(whose allOf includes ./common.yaml#/components/schemas/User) with an inline schema with properties")
}

// TestExternalAllOfMember31: OpenAPI 3.1 says nullable with a "null" type,
// and the error says so.
func TestExternalAllOfMember31(t *testing.T) {
	code, err := generateWithCommonV3(t, opaqueSpecHeader31+`
    Holder:
      type: object
      properties:
        user:
          allOf:
            - $ref: './common.yaml#/components/schemas/User'
            - type: "null"
`)
	require.NoError(t, err)
	assertField(t, code, "User", "*externalRef0.User")

	_, err = generateWithCommonV3(t, opaqueSpecHeader31+`
    Enriched:
      allOf:
        - $ref: './common.yaml#/components/schemas/User'
        - type: [object, "null"]
`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "with an inline schema with type: ")
	assert.Contains(t, err.Error(), `only annotated (description, type: "null", ...)`)
}

// TestWholeDocumentAllOfMember: a $ref to a document that is a schema has no
// Go type of its own; oapi-codegen inlines it, so an allOf merges it like an
// inline schema.
func TestWholeDocumentAllOfMember(t *testing.T) {
	code, err := generateWithCommonV3(t, opaqueSpecHeader+`
    LocalUser:
      $ref: './user.yaml'
    Enriched:
      allOf:
        - $ref: './user.yaml'
        - properties:
            extra: {type: string}
    EnrichedLocal:
      allOf:
        - $ref: '#/components/schemas/LocalUser'
        - properties:
            extra: {type: string}
`)
	require.NoError(t, err)
	for _, name := range []string{"Enriched", "EnrichedLocal"} {
		assert.Regexp(t, `type `+name+` struct \{\n\tExtra \*string [^\n]*\n\tName  \*string`, code)
	}
}

func TestExternalArrayItems(t *testing.T) {
	code, err := generateWithCommonV3(t, opaqueSpecHeader+`
    Users:
      allOf:
        - type: array
          items: {$ref: './common.yaml#/components/schemas/User'}
        - type: array
          items: {description: A user.}
          minItems: 1
    MaybeUsers:
      allOf:
        - type: array
          items: {$ref: './common.yaml#/components/schemas/User'}
        - type: array
          items: {nullable: true}
`)
	require.NoError(t, err)
	assert.Contains(t, code, "type Users = []externalRef0.User")
	assert.Contains(t, code, "type MaybeUsers = []*externalRef0.User", "the annotating item's nullable is kept")

	_, err = generateWithCommonV3(t, opaqueSpecHeader+`
    Users:
      allOf:
        - type: array
          items: {$ref: './common.yaml#/components/schemas/User'}
        - type: array
          items:
            properties:
              extra: {type: string}
`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error generating type for array: error merging schemas: "+
		"allOf can't merge ./common.yaml#/components/schemas/User with an inline schema with properties")
}

// TestExternalRefsOneLevelDown: an external schema used as a property, items
// or additionalProperties of a member is just a type name, and merges fine.
func TestExternalRefsOneLevelDown(t *testing.T) {
	code, err := generateWithCommonV3(t, opaqueSpecHeader+`
    Base:
      type: object
      properties:
        owner: {$ref: './common.yaml#/components/schemas/User'}
    Enriched:
      allOf:
        - $ref: '#/components/schemas/Base'
        - type: object
          properties:
            members:
              type: array
              items: {$ref: './common.yaml#/components/schemas/User'}
            byName:
              type: object
              additionalProperties: {$ref: './common.yaml#/components/schemas/User'}
`)
	require.NoError(t, err)
	assertField(t, code, "Owner", "*externalRef0.User")
	assertField(t, code, "Members", "*[]externalRef0.User")
	assertField(t, code, "ByName", "*map[string]externalRef0.User")
}

// TestExternalRefsInUnions: oneOf and anyOf don't read their variants, so
// external variants keep working.
func TestExternalRefsInUnions(t *testing.T) {
	code, err := generateWithCommonV3(t, opaqueSpecHeader+`
    Either:
      oneOf:
        - $ref: './common.yaml#/components/schemas/User'
        - type: string
`)
	require.NoError(t, err)
	assert.Contains(t, code, "func (t Either) AsExternalRef0User() (externalRef0.User, error)")
}

// TestAnnotatedOpaqueTypeName: x-go-type-name on an annotating member names
// the composition's type, as it does anywhere else.
func TestAnnotatedOpaqueTypeName(t *testing.T) {
	code, err := generateWithCommonV3(t, opaqueSpecHeader+`
    Decorated:
      allOf:
        - $ref: './common.yaml#/components/schemas/User'
        - x-go-type-name: MyUser
    Holder:
      type: object
      properties:
        owner:
          allOf:
            - $ref: './common.yaml#/components/schemas/User'
            - x-go-type-name: Owner
              nullable: true
`)
	require.NoError(t, err)
	assert.Contains(t, code, "type Decorated = MyUser")
	assert.Contains(t, code, "type MyUser = externalRef0.User")
	assert.Contains(t, code, "type Owner = externalRef0.User")
	assertField(t, code, "Owner", "*Owner")
}

// TestExternalAllOfMemberSamePackage: a document imported into this package
// (import-mapping "-") is still generated by another run.
func TestExternalAllOfMemberSamePackage(t *testing.T) {
	_, err := generateWithCommonV3(t, opaqueSpecHeader+`
    Enriched:
      allOf:
        - $ref: './common.yaml#/components/schemas/User'
        - properties:
            extra: {type: string}
`, func(c *Configuration) { c.ImportMapping = map[string]string{"./common.yaml": "-"} })
	require.Error(t, err)
	assert.Contains(t, err.Error(), "a reference to another document can't be merged")
}

const xGoTypeClient = `
    Client:
      type: object
      properties:
        name: {type: string}
      x-go-type: OverlayClient
`

func TestXGoTypeAllOfMemberAnnotated(t *testing.T) {
	code := generateSpec(t, opaqueSpecHeader+xGoTypeClient+`
    Decorated:
      allOf:
        - $ref: '#/components/schemas/Client'
        - description: A client.
          nullable: true
    Holder:
      type: object
      properties:
        count:
          allOf:
            - x-go-type: int64
            - nullable: true
`, withV3)
	assert.Contains(t, code, "type Decorated = Client")
	assertField(t, code, "Count", "*int64")
}

func TestXGoTypeAllOfMemberAnnotatedMore(t *testing.T) {
	code := generateSpec(t, opaqueSpecHeader+xGoTypeClient+`
    Plain:
      type: object
      properties:
        name: {type: string}
    Clients:
      allOf:
        - type: array
          items: {$ref: '#/components/schemas/Client'}
        - type: array
          items: {nullable: true}
    Holder:
      type: object
      properties:
        client:
          allOf:
            - $ref: '#/components/schemas/Client'
            - x-go-type-skip-optional-pointer: true
        wrapped:
          allOf:
            - $ref: '#/components/schemas/Plain'
              x-go-type: Wrapper
            - nullable: true
`, withV3)
	assert.Contains(t, code, "type Clients = []*Client")
	assertField(t, code, "Client", "Client")
	assertField(t, code, "Wrapped", "*Wrapper")
}

func TestXGoTypeAllOfMemberMergedIsAnError(t *testing.T) {
	for name, tc := range map[string]struct {
		spec    string
		message string
	}{
		"a $ref member (#2335)": {
			spec: xGoTypeClient + `
    ClientWithId:
      allOf:
        - $ref: '#/components/schemas/Client'
        - properties:
            id: {type: string}
`,
			message: "allOf can't merge #/components/schemas/Client with an inline schema with properties: " +
				"x-go-type replaces #/components/schemas/Client with OverlayClient",
		},
		"an inline member": {
			spec: `
    Timestamped:
      allOf:
        - x-go-type: time.Time
        - properties:
            zone: {type: string}
`,
			message: "allOf can't merge an inline schema with x-go-type time.Time with an inline schema with properties: " +
				"x-go-type replaces an inline schema with time.Time",
		},
		"x-go-type next to a $ref": {
			spec: `
    Plain:
      type: object
      properties:
        name: {type: string}
    Wrapped:
      allOf:
        - $ref: '#/components/schemas/Plain'
          x-go-type: Wrapper
        - properties:
            id: {type: string}
`,
			message: "x-go-type replaces #/components/schemas/Plain with Wrapper",
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := generateSpecErr(opaqueSpecHeader+tc.spec, withV3)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.message)
			assert.Contains(t, err.Error(), "Give the composition an x-go-type of its own")
			assert.NotContains(t, err.Error(), "schema-merging-behavior")
		})
	}
}

// TestXGoTypeAllOfMemberThroughLocalAlias: a component that is only a $ref
// to an x-go-type schema is that x-go-type too.
func TestXGoTypeAllOfMemberThroughLocalAlias(t *testing.T) {
	_, err := generateSpecErr(opaqueSpecHeader+xGoTypeClient+`
    ClientAlias:
      $ref: '#/components/schemas/Client'
    ClientWithId:
      allOf:
        - $ref: '#/components/schemas/ClientAlias'
        - properties:
            id: {type: string}
`, withV3)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "x-go-type replaces #/components/schemas/ClientAlias with OverlayClient")
}

// TestXGoTypeOnComposition: the composition's own x-go-type replaces the
// whole allOf, so its members are never merged.
func TestXGoTypeOnComposition(t *testing.T) {
	code := generateSpec(t, opaqueSpecHeader+xGoTypeClient+`
    ClientWithId:
      x-go-type: MyClientWithId
      allOf:
        - $ref: '#/components/schemas/Client'
        - properties:
            id: {type: string}
`, withV3)
	assert.Contains(t, code, "type ClientWithId = MyClientWithId")
}

// TestXGoTypeAllOfMemberV2: v2 drops the member's x-go-type and merges the
// schema it replaces, as it always has.
func TestXGoTypeAllOfMemberV2(t *testing.T) {
	code := generateSpec(t, opaqueSpecHeader+xGoTypeClient+`
    ClientWithId:
      allOf:
        - $ref: '#/components/schemas/Client'
        - properties:
            id: {type: string}
`)
	assert.Contains(t, code, "type ClientWithId struct {")
}

func TestTypeKeywords(t *testing.T) {
	maxLen := uint64(3)
	for name, tc := range map[string]struct {
		schema openapi3.Schema
		want   []string
	}{
		"documentation, nullability, validation and extensions": {
			schema: openapi3.Schema{
				Title: "t", Description: "d", Default: 1, Example: 2, Deprecated: true,
				ReadOnly: true, WriteOnly: true, Nullable: true,
				MinLength: 1, MaxLength: &maxLen, Pattern: "^a", UniqueItems: true,
				Extensions: map[string]any{"x-go-type-skip-optional-pointer": true},
			},
		},
		"3.1 null type":  {schema: openapi3.Schema{Type: &openapi3.Types{"null"}}},
		"empty required": {schema: openapi3.Schema{Required: []string{}}},
		"type":           {schema: openapi3.Schema{Type: &openapi3.Types{"object"}}, want: []string{"type"}},
		"nullable type":  {schema: openapi3.Schema{Type: &openapi3.Types{"string", "null"}}, want: []string{"type"}},
		"format":         {schema: openapi3.Schema{Format: "date-time"}, want: []string{"format"}},
		"enum":           {schema: openapi3.Schema{Enum: []any{"a"}}, want: []string{"enum"}},
		"object": {
			schema: openapi3.Schema{Required: []string{"a"}, Properties: openapi3.Schemas{"a": nil}},
			want:   []string{"required", "properties"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, typeKeywords(tc.schema))
		})
	}
}

func TestAnnotatesOnly(t *testing.T) {
	assert.True(t, annotatesOnly(&openapi3.SchemaRef{Value: &openapi3.Schema{Nullable: true}}))
	assert.False(t, annotatesOnly(&openapi3.SchemaRef{Ref: "#/components/schemas/A", Value: &openapi3.Schema{}}),
		"a $ref contributes a schema")
	assert.False(t, annotatesOnly(&openapi3.SchemaRef{Value: &openapi3.Schema{
		Extensions: map[string]any{extPropGoType: "T"},
	}}), "an x-go-type member is a type")
	assert.False(t, annotatesOnly(&openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"object"}}}))
}
