package codegen

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// schemaWithExtension builds a minimal *openapi3.T whose single schema
// component carries the supplied extension map.
func schemaWithExtension(ext map[string]any) *openapi3.T {
	s := openapi3.NewSchema()
	s.Type = &openapi3.Types{"string"}
	s.Extensions = ext
	ref := &openapi3.SchemaRef{Value: s}
	return &openapi3.T{
		OpenAPI: "3.0.3",
		Info:    &openapi3.Info{Title: "t", Version: "0"},
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{"S": ref},
		},
	}
}

func TestValidateSpec_XGoTypeImport_Valid(t *testing.T) {
	cases := []struct {
		name string
		ext  map[string]any
	}{
		{
			name: "identifier alias",
			ext: map[string]any{
				"x-go-type": "mypkg.MyType",
				"x-go-type-import": map[string]any{
					"name": "mypkg",
					"path": "example.com/mymodule/mypkg",
				},
			},
		},
		{
			name: "blank alias",
			ext: map[string]any{
				"x-go-type": "something.T",
				"x-go-type-import": map[string]any{
					"name": "_",
					"path": "example.com/mod",
				},
			},
		},
		{
			name: "dot alias",
			ext: map[string]any{
				"x-go-type": "T",
				"x-go-type-import": map[string]any{
					"name": ".",
					"path": "example.com/mod",
				},
			},
		},
		{
			name: "no alias",
			ext: map[string]any{
				"x-go-type": "pkg.T",
				"x-go-type-import": map[string]any{
					"path": "example.com/mod",
				},
			},
		},
		{
			name: "empty alias",
			ext: map[string]any{
				"x-go-type": "pkg.T",
				"x-go-type-import": map[string]any{
					"name": "",
					"path": "example.com/mod",
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSpec(schemaWithExtension(tc.ext))
			assert.NoError(t, err)
		})
	}
}

func TestValidateSpec_XGoTypeImport_NameInjection(t *testing.T) {
	// GHSA-9c2f-gr95-7wqw: x-go-type-import.name is emitted unquoted into the
	// import block; a newline closes the block and allows arbitrary code injection.
	injected := "_ \"unsafe\"\n)\nfunc init() { panic(\"injected\") }\nvar (\n_ ="
	ext := map[string]any{
		"x-go-type": "os.FileInfo",
		"x-go-type-import": map[string]any{
			"name": injected,
			"path": "ignored",
		},
	}
	err := ValidateSpec(schemaWithExtension(ext))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "x-go-type-import name")
}

func TestValidateSpec_XGoTypeImport_NameWithNewline(t *testing.T) {
	ext := map[string]any{
		"x-go-type": "pkg.T",
		"x-go-type-import": map[string]any{
			"name": "alias\ninjected",
			"path": "example.com/mod",
		},
	}
	err := ValidateSpec(schemaWithExtension(ext))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "x-go-type-import name")
}

func TestValidateSpec_XGoTypeImport_NameWithBacktick(t *testing.T) {
	ext := map[string]any{
		"x-go-type": "pkg.T",
		"x-go-type-import": map[string]any{
			"name": "alias`bad",
			"path": "example.com/mod",
		},
	}
	err := ValidateSpec(schemaWithExtension(ext))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "x-go-type-import name")
}

func TestValidateSpec_XGoTypeImport_PathWithControlChar(t *testing.T) {
	ext := map[string]any{
		"x-go-type": "pkg.T",
		"x-go-type-import": map[string]any{
			"name": "mypkg",
			"path": "example.com/mod\x00injected",
		},
	}
	err := ValidateSpec(schemaWithExtension(ext))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "x-go-type-import path")
}

func TestValidateSpec_XGoTypeImport_CaseInsensitiveKeys(t *testing.T) {
	// Matches the case-insensitive key handling in ParseGoImportExtension.
	ext := map[string]any{
		"x-go-type": "pkg.T",
		"x-go-type-import": map[string]any{
			"Name": "mypkg",
			"Path": "example.com/mod",
		},
	}
	err := ValidateSpec(schemaWithExtension(ext))
	assert.NoError(t, err)
}

// Regression: pre-existing extensions must still be validated.
func TestValidateSpec_XGoName_StillValidated(t *testing.T) {
	ext := map[string]any{
		"x-go-name": "bad\nname",
	}
	err := ValidateSpec(schemaWithExtension(ext))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "x-go-name")
}

func TestValidateSpec_XGoType_StillValidated(t *testing.T) {
	ext := map[string]any{
		"x-go-type": "pkg.T\nfunc init() {}",
	}
	err := ValidateSpec(schemaWithExtension(ext))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "x-go-type")
}

func TestValidateSpec_XGoTypeName_StillValidated(t *testing.T) {
	ext := map[string]any{
		"x-go-type-name": "Bad\nName",
	}
	err := ValidateSpec(schemaWithExtension(ext))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "x-go-type-name")
}

func TestValidateSpec_XOapiCodegenExtraTags_StillValidated(t *testing.T) {
	ext := map[string]any{
		"x-oapi-codegen-extra-tags": map[string]any{
			"bad\ntag": "value",
		},
	}
	err := ValidateSpec(schemaWithExtension(ext))
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "x-oapi-codegen-extra-tags"))
}

// minimalOperationSpec builds a document with one GET /x operation whose
// operationId and responses are set by the caller.
func minimalOperationSpec(operationID string, responseKeys ...string) *openapi3.T {
	responses := openapi3.NewResponses()
	for _, key := range responseKeys {
		desc := "d"
		responses.Set(key, &openapi3.ResponseRef{Value: &openapi3.Response{Description: &desc}})
	}
	op := &openapi3.Operation{OperationID: operationID, Responses: responses}
	paths := openapi3.NewPaths()
	paths.Set("/x", &openapi3.PathItem{Get: op})
	return &openapi3.T{
		OpenAPI: "3.0.3",
		Info:    &openapi3.Info{Title: "t", Version: "0"},
		Paths:   paths,
	}
}

func TestValidateSpec_OperationId(t *testing.T) {
	// The raw operationId is copied, quoted, into the generated code as the
	// Echo per-operation middleware key, so quotes are fine but a control
	// character is not.
	require.NoError(t, ValidateSpec(minimalOperationSpec(`x"]...); panic("code-injection"); //`, "200")))

	err := ValidateSpec(minimalOperationSpec("x\nfunc init() {}", "200"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "operationId")
}

func TestValidateSpec_ResponseKeys(t *testing.T) {
	for _, key := range []string{"100", "200", "404", "599", "999", "1XX", "2XX", "5XX", "default"} {
		assert.NoError(t, ValidateSpec(minimalOperationSpec("op", key)), key)
	}
	for _, key := range []string{"200 || os.Exit(1) == 0", "2xx", "20", "2000", "abc", "", "Default", "0XX", "6XX"} {
		err := ValidateSpec(minimalOperationSpec("op", key))
		require.Error(t, err, "%q", key)
		assert.Contains(t, err.Error(), "response key", "%q", key)
	}
}

func TestValidateSpec_ComponentNamesWithControlCharacters(t *testing.T) {
	// Component names reach single-line doc comments ("X defines model for
	// <name>.") verbatim, so a newline would end the comment and start code.
	bad := "Name\nfunc init() { panic(\"injected\") }\n//"
	desc := "d"
	schema := &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}
	cases := map[string]*openapi3.Components{
		"schema name":                 {Schemas: openapi3.Schemas{bad: schema}},
		"component parameter name":    {Parameters: openapi3.ParametersMap{bad: {Value: &openapi3.Parameter{Name: "p", In: "query", Schema: schema}}}},
		"component request body name": {RequestBodies: openapi3.RequestBodies{bad: {Value: &openapi3.RequestBody{}}}},
		"component response name":     {Responses: openapi3.ResponseBodies{bad: {Value: &openapi3.Response{Description: &desc}}}},
		"component header name":       {Headers: openapi3.Headers{bad: {Value: &openapi3.Header{}}}},
		"component callback name":     {Callbacks: openapi3.Callbacks{bad: {Value: &openapi3.Callback{}}}},
		"security scheme name":        {SecuritySchemes: openapi3.SecuritySchemes{bad: {Value: openapi3.NewSecurityScheme()}}},
	}
	for what, components := range cases {
		t.Run(what, func(t *testing.T) {
			spec := &openapi3.T{OpenAPI: "3.0.3", Info: &openapi3.Info{Title: "t", Version: "0"}, Components: components}
			err := ValidateSpec(spec)
			require.Error(t, err)
			assert.Contains(t, err.Error(), what)
			assert.Contains(t, err.Error(), "newline")
		})
	}

	// Quotes in a component name are fine: the name only ever reaches a
	// comment verbatim, and a Go identifier after sanitization.
	spec := &openapi3.T{OpenAPI: "3.0.3", Info: &openapi3.Info{Title: "t", Version: "0"},
		Components: &openapi3.Components{Schemas: openapi3.Schemas{`Quoted"Name`: schema}}}
	require.NoError(t, ValidateSpec(spec))
}

func TestValidateSpec_ResponseHeaderStyle(t *testing.T) {
	// A response header's style is copied into a string literal by the strict
	// server templates, like a parameter's style is, so it is checked the same way.
	desc := "d"
	responses := openapi3.NewResponses()
	responses.Set("200", &openapi3.ResponseRef{Value: &openapi3.Response{
		Description: &desc,
		Headers: openapi3.Headers{"X-Trace": {Value: &openapi3.Header{Parameter: openapi3.Parameter{
			Style:  `simple"); panic("injected`,
			Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
		}}}},
	}})
	paths := openapi3.NewPaths()
	paths.Set("/x", &openapi3.PathItem{Get: &openapi3.Operation{OperationID: "op", Responses: responses}})
	spec := &openapi3.T{OpenAPI: "3.0.3", Info: &openapi3.Info{Title: "t", Version: "0"}, Paths: paths}

	err := ValidateSpec(spec)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "style of header")
}

func TestValidateServerURLs(t *testing.T) {
	serverSpec := func(server *openapi3.Server) *openapi3.T {
		return &openapi3.T{OpenAPI: "3.0.3", Info: &openapi3.Info{Title: "t", Version: "0"}, Servers: openapi3.Servers{server}}
	}

	t.Run("a well-formed server passes", func(t *testing.T) {
		spec := serverSpec(&openapi3.Server{
			URL:        "https://{region}.example.com/{basePath}/{undeclared}",
			Extensions: map[string]any{"x-go-name": "Production"},
			Variables: map[string]*openapi3.ServerVariable{
				"region":   {Enum: []string{"eu", "usEast", "Üst", "8443", "_x"}, Default: "eu"},
				"basePath": {Default: "v1"},
			},
		})
		require.NoError(t, ValidateServerURLs(spec))
	})

	t.Run("x-go-name must be an identifier", func(t *testing.T) {
		err := ValidateServerURLs(serverSpec(&openapi3.Server{
			URL:        "https://example.com",
			Extensions: map[string]any{"x-go-name": "Prod; func init() { panic(1) }; var _"},
		}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "x-go-name")
	})

	t.Run("variable names become parameters and must be identifiers", func(t *testing.T) {
		err := ValidateServerURLs(serverSpec(&openapi3.Server{
			URL:       "https://example.com/{api-version}",
			Variables: map[string]*openapi3.ServerVariable{"api-version": {Default: "1"}},
		}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "server variable name")
	})

	t.Run("undeclared placeholders become parameters and must be identifiers", func(t *testing.T) {
		err := ValidateServerURLs(serverSpec(&openapi3.Server{
			URL: `https://example.com/{a string); panic(1); var _ = "}`,
		}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "undeclared URL placeholder")
	})

	t.Run("enum values are appended to the type name and must be identifier characters", func(t *testing.T) {
		for _, bad := range []string{`x = 1; func init() { panic(1) }; var _`, "us-east", "a b", ""} {
			err := ValidateServerURLs(serverSpec(&openapi3.Server{
				URL:       "https://{region}.example.com",
				Variables: map[string]*openapi3.ServerVariable{"region": {Enum: []string{"eu", bad}, Default: "eu"}},
			}))
			require.Error(t, err, "%q", bad)
			assert.Contains(t, err.Error(), "enum value of server variable")
		}
	})

	t.Run("variables the URL does not use produce no code and are not checked", func(t *testing.T) {
		spec := serverSpec(&openapi3.Server{
			URL: "https://{region}.example.com",
			Variables: map[string]*openapi3.ServerVariable{
				"region":      {Enum: []string{"eu"}, Default: "eu"},
				"api-version": {Default: "1"},
				"zone":        {Enum: []string{"us-east"}, Default: "us-east"},
			},
		})
		require.NoError(t, ValidateServerURLs(spec))
	})

	t.Run("ValidateSpec itself does not impose the identifier rules", func(t *testing.T) {
		// Hyphenated variable names are fine when server URLs are not generated.
		spec := serverSpec(&openapi3.Server{
			URL:       "https://example.com/{api-version}",
			Variables: map[string]*openapi3.ServerVariable{"api-version": {Default: "1"}},
		})
		require.NoError(t, ValidateSpec(spec))
	})
}

func TestValidateSpec_ImplicitDiscriminatorMappingKey(t *testing.T) {
	// A $ref variant that the mapping does not name is mapped under the last
	// segment of its ref, and that key is written into the generated union's
	// switch like an explicit mapping key, so it is checked like one. Quotes
	// are otherwise allowed in a schema name (see
	// TestValidateSpec_ComponentNamesWithControlCharacters), and a ref into
	// another document names a schema this document never defines.
	str := &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}
	variant := func() *openapi3.SchemaRef {
		return &openapi3.SchemaRef{Value: &openapi3.Schema{
			Type:       &openapi3.Types{"object"},
			Properties: openapi3.Schemas{"petType": str},
		}}
	}
	union := func(discriminator *openapi3.Discriminator, refs ...string) *openapi3.Schema {
		s := &openapi3.Schema{Discriminator: discriminator}
		for _, ref := range refs {
			s.OneOf = append(s.OneOf, &openapi3.SchemaRef{Ref: ref, Value: variant().Value})
		}
		return s
	}
	spec := func(schemas openapi3.Schemas) *openapi3.T {
		return &openapi3.T{OpenAPI: "3.0.3", Info: &openapi3.Info{Title: "t", Version: "0"},
			Components: &openapi3.Components{Schemas: schemas}}
	}

	t.Run("local ref with a quote in the schema name", func(t *testing.T) {
		name := `Dog"Name`
		err := ValidateSpec(spec(openapi3.Schemas{
			"Cat": variant(),
			name:  variant(),
			"Pet": {Value: union(&openapi3.Discriminator{PropertyName: "petType"},
				"#/components/schemas/Cat", "#/components/schemas/"+name)},
		}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "implicit discriminator mapping key")
		assert.Contains(t, err.Error(), "double quote")
	})

	t.Run("local ref with a backtick in the schema name, anyOf", func(t *testing.T) {
		name := "Dog`Name"
		s := union(&openapi3.Discriminator{PropertyName: "petType"})
		s.AnyOf = []*openapi3.SchemaRef{{Ref: "#/components/schemas/" + name, Value: variant().Value}}
		err := ValidateSpec(spec(openapi3.Schemas{name: variant(), "Pet": {Value: s}}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "implicit discriminator mapping key")
		assert.Contains(t, err.Error(), "backtick")
	})

	t.Run("external ref names a schema this document does not define", func(t *testing.T) {
		err := ValidateSpec(spec(openapi3.Schemas{
			"Pet": {Value: union(&openapi3.Discriminator{PropertyName: "petType"},
				`other.yaml#/components/schemas/Dog"Name`)},
		}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "implicit discriminator mapping key")
	})

	t.Run("local ref with a backslash in the schema name", func(t *testing.T) {
		// A backslash starts an escape sequence in the `case "..."` label.
		name := `Dog\Name`
		err := ValidateSpec(spec(openapi3.Schemas{
			name:  variant(),
			"Pet": {Value: union(&openapi3.Discriminator{PropertyName: "petType"}, "#/components/schemas/"+name)},
		}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "implicit discriminator mapping key")
		assert.Contains(t, err.Error(), "backslash")
	})

	t.Run("explicit mapping key with a backslash", func(t *testing.T) {
		d := &openapi3.Discriminator{PropertyName: "petType", Mapping: map[string]openapi3.MappingRef{
			`dog\x`: {Ref: "#/components/schemas/Cat"},
		}}
		err := ValidateSpec(spec(openapi3.Schemas{"Cat": variant(), "Pet": {Value: union(d, "#/components/schemas/Cat")}}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "discriminator mapping key")
		assert.Contains(t, err.Error(), "backslash")
	})

	t.Run("null branch is a nullability marker, not a key", func(t *testing.T) {
		// The union generators skip a null-typed branch before mapping, so
		// its name never reaches the generated code (OpenAPI 3.1).
		name := `Null"Value`
		null := &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"null"}}}
		s := union(&openapi3.Discriminator{PropertyName: "petType"}, "#/components/schemas/Cat")
		s.OneOf = append(s.OneOf, &openapi3.SchemaRef{Ref: "#/components/schemas/" + name, Value: null.Value})
		require.NoError(t, ValidateSpec(spec(openapi3.Schemas{"Cat": variant(), name: null, "Pet": {Value: s}})))
	})

	// The allOf merge concatenates the members' oneOf lists and lets the one
	// member with a discriminator govern the result, so the implicit keys of
	// a union split across members are checked on the merged view.
	name := `Dog"Name`
	variantsMember := func() *openapi3.SchemaRef {
		return &openapi3.SchemaRef{Value: union(nil, "#/components/schemas/Cat", "#/components/schemas/"+name)}
	}
	discriminatorMember := func(mapping map[string]openapi3.MappingRef) *openapi3.SchemaRef {
		return &openapi3.SchemaRef{Value: &openapi3.Schema{Discriminator: &openapi3.Discriminator{PropertyName: "petType", Mapping: mapping}}}
	}
	components := func(pet *openapi3.Schema) openapi3.Schemas {
		return openapi3.Schemas{"Cat": variant(), name: variant(), "Pet": {Value: pet}}
	}
	requireImplicitKeyError := func(t *testing.T, err error) {
		t.Helper()
		require.Error(t, err)
		assert.Contains(t, err.Error(), `implicit discriminator mapping key for #/components/schemas/Dog"Name in schema "Pet"`)
		assert.Contains(t, err.Error(), "double quote")
	}

	t.Run("allOf: variants in one member, discriminator in another", func(t *testing.T) {
		pet := &openapi3.Schema{AllOf: []*openapi3.SchemaRef{variantsMember(), discriminatorMember(nil)}}
		requireImplicitKeyError(t, ValidateSpec(spec(components(pet))))
	})

	t.Run("allOf: discriminator on the schema, variants in a member", func(t *testing.T) {
		pet := &openapi3.Schema{
			Discriminator: &openapi3.Discriminator{PropertyName: "petType"},
			AllOf:         []*openapi3.SchemaRef{variantsMember()},
		}
		requireImplicitKeyError(t, ValidateSpec(spec(components(pet))))
	})

	t.Run("allOf: variants behind a $ref member, nested allOf", func(t *testing.T) {
		// The merge resolves $ref members and flattens nested allOf, so the
		// variants are found through both.
		schemas := components(nil)
		schemas["Variants"] = variantsMember()
		schemas["Pet"] = &openapi3.SchemaRef{Value: &openapi3.Schema{AllOf: []*openapi3.SchemaRef{
			{Value: &openapi3.Schema{AllOf: []*openapi3.SchemaRef{{Ref: "#/components/schemas/Variants", Value: schemas["Variants"].Value}}}},
			discriminatorMember(nil),
		}}}
		requireImplicitKeyError(t, ValidateSpec(spec(schemas)))
	})

	t.Run("allOf: the discriminator member maps the ref explicitly", func(t *testing.T) {
		mapping := map[string]openapi3.MappingRef{"dog": {Ref: "#/components/schemas/" + name}}
		pet := &openapi3.Schema{AllOf: []*openapi3.SchemaRef{variantsMember(), discriminatorMember(mapping)}}
		require.NoError(t, ValidateSpec(spec(components(pet))))
	})

	t.Run("allOf: no discriminator anywhere", func(t *testing.T) {
		pet := &openapi3.Schema{AllOf: []*openapi3.SchemaRef{variantsMember(), {Value: &openapi3.Schema{Type: &openapi3.Types{"object"}}}}}
		require.NoError(t, ValidateSpec(spec(components(pet))))
	})

	t.Run("allOf: the v1 merge generates each member on its own", func(t *testing.T) {
		// Under old-merge-schemas the members are not combined, so the member
		// with the variants has no discriminator and no key is derived. A
		// member whose own oneOf and discriminator sit together is still a
		// discriminated union.
		globalState.options.Compatibility.OldMergeSchemas = true
		t.Cleanup(func() { globalState.options.Compatibility.OldMergeSchemas = false })
		pet := &openapi3.Schema{AllOf: []*openapi3.SchemaRef{variantsMember(), discriminatorMember(nil)}}
		require.NoError(t, ValidateSpec(spec(components(pet))))

		pet = &openapi3.Schema{AllOf: []*openapi3.SchemaRef{
			{Value: union(&openapi3.Discriminator{PropertyName: "petType"}, "#/components/schemas/Cat", "#/components/schemas/"+name)},
		}}
		err := ValidateSpec(spec(components(pet)))
		require.Error(t, err)
		assert.Contains(t, err.Error(), `implicit discriminator mapping key for #/components/schemas/Dog"Name in allOf[0] in schema "Pet"`)
	})

	t.Run("allOf: a composition that refers back to itself ends", func(t *testing.T) {
		pet := &openapi3.Schema{AllOf: []*openapi3.SchemaRef{discriminatorMember(nil)}}
		pet.AllOf = append(pet.AllOf, &openapi3.SchemaRef{Ref: "#/components/schemas/Pet", Value: pet})
		require.NoError(t, ValidateSpec(spec(components(pet))))
	})

	t.Run("explicitly mapped ref uses the mapping key instead", func(t *testing.T) {
		name := `Dog"Name`
		d := &openapi3.Discriminator{PropertyName: "petType", Mapping: map[string]openapi3.MappingRef{
			"dog": {Ref: "#/components/schemas/" + name},
		}}
		require.NoError(t, ValidateSpec(spec(openapi3.Schemas{
			name:  variant(),
			"Pet": {Value: union(d, "#/components/schemas/"+name)},
		})))
	})

	t.Run("no discriminator, nothing to map", func(t *testing.T) {
		name := `Dog"Name`
		require.NoError(t, ValidateSpec(spec(openapi3.Schemas{
			name:  variant(),
			"Pet": {Value: union(nil, "#/components/schemas/"+name)},
		})))
	})
}

func TestValidateSpec_WholeDocumentRefIsWalked(t *testing.T) {
	// A $ref without a fragment ("./other.yaml") is inlined by code generation,
	// so the document it points at reaches the generated source and is walked.
	// A ref with a fragment is not: it is a type reference.
	str := &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}}
	badName := "a`b"
	external := &openapi3.Schema{
		Type:       &openapi3.Types{"object"},
		Properties: openapi3.Schemas{badName: str},
	}
	container := func(ref string) openapi3.Schemas {
		return openapi3.Schemas{"Container": {Value: &openapi3.Schema{
			Type:       &openapi3.Types{"object"},
			Properties: openapi3.Schemas{"ext": {Ref: ref, Value: external}},
		}}}
	}
	spec := func(components *openapi3.Components) *openapi3.T {
		return &openapi3.T{OpenAPI: "3.0.3", Info: &openapi3.Info{Title: "t", Version: "0"}, Components: components}
	}

	t.Run("schema", func(t *testing.T) {
		err := ValidateSpec(spec(&openapi3.Components{Schemas: container("./evil.json")}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "property name")
		assert.Contains(t, err.Error(), "backtick")
		assert.Contains(t, err.Error(), `property "ext"`)
	})

	t.Run("x-go-type replaces the document, so it is not walked", func(t *testing.T) {
		// generateGoSchema returns the override before reading the document,
		// whether the extension sits next to the $ref or at the document's root.
		schemas := container("./evil.json")
		schemas["Container"].Value.Properties["ext"].Extensions = map[string]any{"x-go-type": "string"}
		require.NoError(t, ValidateSpec(spec(&openapi3.Components{Schemas: schemas})))

		rooted := &openapi3.Schema{
			Type:       &openapi3.Types{"object"},
			Properties: openapi3.Schemas{badName: str},
			Extensions: map[string]any{"x-go-type": "string"},
		}
		schemas = openapi3.Schemas{"Container": {Value: &openapi3.Schema{
			Type:       &openapi3.Types{"object"},
			Properties: openapi3.Schemas{"ext": {Ref: "./evil.json", Value: rooted}},
		}}}
		require.NoError(t, ValidateSpec(spec(&openapi3.Components{Schemas: schemas})))

		// The override itself is still checked.
		rooted.Extensions["x-go-type"] = "string; func init() {}"
		err := ValidateSpec(spec(&openapi3.Components{Schemas: schemas}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "x-go-type")
	})

	t.Run("a document replaced by x-go-type in one place is inlined in another", func(t *testing.T) {
		// The loader hands every ref to the same document the same value. The
		// ref with the override, walked first, must not mark the document as
		// visited, or the ref without it is never checked.
		schemas := openapi3.Schemas{"Container": {Value: &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			Properties: openapi3.Schemas{
				"a": {Ref: "./evil.json", Value: external, Extensions: map[string]any{"x-go-type": "string"}},
				"b": {Ref: "./evil.json", Value: external},
			},
		}}}
		err := ValidateSpec(spec(&openapi3.Components{Schemas: schemas}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), `property "b"`)
		assert.Contains(t, err.Error(), "backtick")
	})

	t.Run("fragment ref is a type reference and is not walked", func(t *testing.T) {
		require.NoError(t, ValidateSpec(spec(&openapi3.Components{Schemas: container("other.yaml#/components/schemas/X")})))
		require.NoError(t, ValidateSpec(spec(&openapi3.Components{Schemas: container("#/components/schemas/X")})))
	})

	t.Run("parameter", func(t *testing.T) {
		err := ValidateSpec(spec(&openapi3.Components{Parameters: openapi3.ParametersMap{
			"P": {Ref: "./param.yaml", Value: &openapi3.Parameter{Name: badName, In: "query", Schema: str}},
		}}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parameter name")
	})

	t.Run("response", func(t *testing.T) {
		desc := "d"
		err := ValidateSpec(spec(&openapi3.Components{Responses: openapi3.ResponseBodies{
			"R": {Ref: "./resp.yaml", Value: &openapi3.Response{Description: &desc, Headers: openapi3.Headers{badName: {Value: &openapi3.Header{}}}}},
		}}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "response header name")
	})

	t.Run("documents referring to each other end", func(t *testing.T) {
		a := &openapi3.Schema{Type: &openapi3.Types{"object"}, Properties: openapi3.Schemas{}}
		b := &openapi3.Schema{Type: &openapi3.Types{"object"}, Properties: openapi3.Schemas{}}
		a.Properties["b"] = &openapi3.SchemaRef{Ref: "./b.yaml", Value: b}
		b.Properties["a"] = &openapi3.SchemaRef{Ref: "./a.yaml", Value: a}
		require.NoError(t, ValidateSpec(spec(&openapi3.Components{Schemas: openapi3.Schemas{"A": {Value: a}}})))
	})
}
