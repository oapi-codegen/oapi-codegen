package codegen

import (
	"go/format"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const contentTypesVendoredSpec = `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Content types test
paths:
  /pets:
    post:
      operationId: addPet
      requestBody:
        required: true
        content:
          application/vnd.mycompany.v1+json:
            schema:
              $ref: '#/components/schemas/Pet'
      responses:
        '200':
          description: ok
          content:
            application/vnd.mycompany.v1+json:
              schema:
                $ref: '#/components/schemas/Pet'
components:
  schemas:
    Pet:
      type: object
      properties:
        name:
          type: string
`

func contentTypesTestConfiguration(contentTypes map[string][]string) Configuration {
	return Configuration{
		PackageName: "api",
		Generate: GenerateOptions{
			StdHTTPServer: true,
			Strict:        true,
			Client:        true,
			Models:        true,
		},
		OutputOptions: OutputOptions{
			ContentTypes: contentTypes,
		},
	}
}

func TestContentTypesShortName(t *testing.T) {
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData([]byte(contentTypesVendoredSpec))
	require.NoError(t, err)

	code, err := Generate(swagger, contentTypesTestConfiguration(map[string][]string{
		"V1": {`^application/vnd\.mycompany\.v1\+json$`},
	}))
	require.NoError(t, err)

	_, err = format.Source([]byte(code))
	require.NoError(t, err)

	// Request body model and client method use the short name.
	assert.Contains(t, code, "type AddPetV1RequestBody = Pet")
	assert.Contains(t, code, "AddPetWithV1Body(")
	// Client response wrapper field uses the short name and is deserialized
	// as JSON (derived from the media type, not the name).
	assert.Contains(t, code, "V1200 *Pet")
	assert.Contains(t, code, "response.V1200 = &dest")
	assert.Contains(t, code, "json.Unmarshal(bodyBytes, &dest)")
	// Strict server envelope uses the short name.
	assert.Contains(t, code, "type AddPet200V1Response Pet")
}

func TestContentTypesUnsetKeepsDefaultNames(t *testing.T) {
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData([]byte(contentTypesVendoredSpec))
	require.NoError(t, err)

	code, err := Generate(swagger, contentTypesTestConfiguration(nil))
	require.NoError(t, err)

	assert.NotContains(t, code, "V1200")
	assert.NotContains(t, code, "AddPetV1RequestBody")
}

func TestContentTypesRenamesBuiltinWithoutChangingBehavior(t *testing.T) {
	spec := `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Content types form test
paths:
  /login:
    post:
      operationId: login
      requestBody:
        required: true
        content:
          application/x-www-form-urlencoded:
            schema:
              type: object
              properties:
                user:
                  type: string
      responses:
        '204':
          description: no content
`
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	// Renaming a built-in means adding the new key and disabling the default
	// one; leaving both matching is an ambiguity error (checked below).
	code, err := Generate(swagger, contentTypesTestConfiguration(map[string][]string{
		"Form":     {`^application/x-www-form-urlencoded$`},
		"Formdata": {},
	}))
	require.NoError(t, err)

	_, err = format.Source([]byte(code))
	require.NoError(t, err)

	// The type names use the custom short name instead of "Formdata"...
	assert.Contains(t, code, "LoginFormRequestBody")
	assert.NotContains(t, code, "LoginFormdataRequestBody")
	// ...but form handling is still derived from the media type.
	assert.Contains(t, code, "runtime.BindForm")
	assert.Contains(t, code, "runtime.MarshalForm")

	// Without disabling the default key, the media type matches under both
	// short names.
	_, err = Generate(swagger, contentTypesTestConfiguration(map[string][]string{
		"Form": {`^application/x-www-form-urlencoded$`},
	}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple short names")
	assert.Contains(t, err.Error(), "Formdata")
}

func TestContentTypesGeneratesModelForUnsupportedMediaType(t *testing.T) {
	spec := `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Content types csv test
paths:
  /report:
    post:
      operationId: uploadReport
      requestBody:
        required: true
        content:
          text/csv:
            schema:
              type: string
      responses:
        '200':
          description: the processed report
          content:
            text/csv:
              schema:
                type: string
`
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	// Without a mapping, text/csv gets no model at all.
	code, err := Generate(swagger, contentTypesTestConfiguration(nil))
	require.NoError(t, err)
	assert.NotContains(t, code, "UploadReportCSVRequestBody")

	code, err = Generate(swagger, contentTypesTestConfiguration(map[string][]string{
		"CSV": {`^text/csv$`},
	}))
	require.NoError(t, err)

	_, err = format.Source([]byte(code))
	require.NoError(t, err)

	// The model is generated with the short name, but the wire-level handling
	// stays the generic io.Reader passthrough.
	assert.Contains(t, code, "type UploadReportCSVBody = string")
	assert.Contains(t, code, "type UploadReportCSVRequestBody = UploadReportCSVBody")
	assert.Contains(t, code, "Body io.Reader")
	assert.NotContains(t, code, "NewUploadReportRequestWithCSVBody")

	// On the response side the mapping affects naming only: typed client
	// response fields are generated solely for media types we know how to
	// deserialize (JSON, YAML, XML), so the mapped CSV response exposes just
	// the raw body — no CSV200 field and no unmarshal case.
	assert.Contains(t, code, "ParseUploadReportResponse")
	assert.NotContains(t, code, "CSV200")
	assert.NotContains(t, code, "response.CSV200")
}

func TestContentTypesAmbiguousMatchErrors(t *testing.T) {
	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData([]byte(contentTypesVendoredSpec))
	require.NoError(t, err)

	_, err = Generate(swagger, contentTypesTestConfiguration(map[string][]string{
		"Mine":  {`^application/vnd\.mycompany\.`},
		"AnyV1": {`\.v1\+json$`},
	}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple short names")
	assert.Contains(t, err.Error(), "AnyV1")
	assert.Contains(t, err.Error(), "Mine")
}

func TestContentTypesValidation(t *testing.T) {
	badRegex := OutputOptions{ContentTypes: map[string][]string{
		"JSON": {`[`},
	}}
	problems := badRegex.Validate()
	require.Contains(t, problems, "content-types")
	assert.Contains(t, problems["content-types"], "invalid content-types pattern")

	badName := OutputOptions{ContentTypes: map[string][]string{
		"not-a-go-name": {`^application/json$`},
	}}
	problems = badName.Validate()
	require.Contains(t, problems, "content-types")
	assert.Contains(t, problems["content-types"], "not usable in Go type names")

	ok := OutputOptions{ContentTypes: map[string][]string{
		"V1": {`^application/vnd\.mycompany\.v1\+json$`},
	}}
	assert.Empty(t, ok.Validate())
}

func TestContentTypeNameTagsTagFor(t *testing.T) {
	// With no user entries, the defaults apply.
	m, err := compileContentTypeNameTags(nil)
	require.NoError(t, err)

	for contentType, want := range map[string]string{
		"application/json":                  "JSON",
		"application/x-www-form-urlencoded": "Formdata",
		"multipart/form-data":               "Multipart",
		"multipart/mixed":                   "Multipart",
		"text/plain":                        "Text",
		// Vendored JSON isn't part of the mapping; callers derive its name
		// from the media type.
		"application/vnd.api+json": "",
		"text/csv":                 "",
	} {
		tag, err := m.TagFor(contentType)
		require.NoError(t, err)
		assert.Equal(t, want, tag, contentType)
	}

	// User entries replace default keys wholesale.
	m, err = compileContentTypeNameTags(map[string][]string{
		"JSON": {`^application/vnd\.foo\+json$`},
		"Text": {},
	})
	require.NoError(t, err)

	tag, err := m.TagFor("application/vnd.foo+json")
	require.NoError(t, err)
	assert.Equal(t, "JSON", tag)

	// application/json no longer matches the replaced JSON key...
	tag, err = m.TagFor("application/json")
	require.NoError(t, err)
	assert.Equal(t, "", tag)

	// ...and an empty list disables a key.
	tag, err = m.TagFor("text/plain")
	require.NoError(t, err)
	assert.Equal(t, "", tag)

	// Untouched defaults survive the merge.
	tag, err = m.TagFor("multipart/form-data")
	require.NoError(t, err)
	assert.Equal(t, "Multipart", tag)
}
