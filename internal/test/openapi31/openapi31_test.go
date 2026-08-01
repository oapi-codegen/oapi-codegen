// Package openapi31 tests detection of OpenAPI 3.1 authoring idioms. The
// assertions are structural -- they instantiate the generated types (a
// regression that drops a synthesis surfaces as a build failure) and, for
// doc-comment behavior, parse the generated source -- rather than
// string-matching the generated source wholesale.
package openapi31

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ----------------------------------------------------------------------------
// contentMediaType / contentEncoding keywords (from openapi31_content_keywords)
// ----------------------------------------------------------------------------

// contentMediaType: application/octet-stream -- raw binary -- must produce
// openapi_types.File, matching 3.0 `format: binary`.
func TestContentMediaTypeMapsToFile(t *testing.T) {
	var f openapi_types.File
	fields := FileUploadFields{
		OrderId:     1,
		RawFile:     f,
		Base64Field: []byte("data"),
	}
	assert.Equal(t, 1, fields.OrderId)
}

// A non-octet-stream contentMediaType (image/png) still routes to
// openapi_types.File: the keyword signals "this string is binary content,"
// regardless of the specific media type.
func TestContentMediaTypeImageMapsToFile(t *testing.T) {
	var f openapi_types.File
	fields := FileUploadFields{
		OrderId:     1,
		RawFile:     openapi_types.File{},
		Base64Field: []byte("data"),
		ImageFile:   &f,
	}
	assert.NotNil(t, fields.ImageFile)
}

// contentEncoding: base64 (standard padded base64) is the one RFC4648 variant
// Go's encoding/json handles natively for []byte, so it maps to []byte --
// matching 3.0 `format: byte`.
func TestContentEncodingBase64MapsToByteSlice(t *testing.T) {
	fields := FileUploadFields{
		OrderId:     1,
		RawFile:     openapi_types.File{},
		Base64Field: []byte("hello"),
	}
	assert.Equal(t, "hello", string(fields.Base64Field))
}

// contentEncoding: base64url (URL-safe base64) is intentionally NOT mapped to
// []byte: Go's JSON codec for []byte uses standard base64 unconditionally,
// which would silently corrupt URL-safe characters (`-`/`_`). The field stays
// `string`. Compile-time check: `*string` assignment.
func TestContentEncodingBase64UrlStaysString(t *testing.T) {
	v := "abc-def_ghi"
	fields := FileUploadFields{
		OrderId:        1,
		RawFile:        openapi_types.File{},
		Base64Field:    []byte("data"),
		Base64UrlField: &v,
	}
	require.NotNil(t, fields.Base64UrlField)
	assert.Equal(t, "abc-def_ghi", *fields.Base64UrlField)
}

// When both contentEncoding and contentMediaType are set, the file mapping wins
// -- the field is a binary blob of a specific media type. Matches the 3.0 model
// where `format: binary` dominated.
func TestContentMediaTypeWinsOverContentEncoding(t *testing.T) {
	var f openapi_types.File
	fields := FileUploadFields{
		OrderId:      1,
		RawFile:      openapi_types.File{},
		Base64Field:  []byte("data"),
		BothKeywords: &f,
	}
	assert.NotNil(t, fields.BothKeywords)
}

// An explicit `format` must continue to win over the contentMediaType-derived
// synthesis. Here `format: date` keeps its openapi_types.Date Go type even
// though contentMediaType: application/octet-stream would otherwise have routed
// to openapi_types.File.
func TestExplicitFormatOverridesContentMediaType(t *testing.T) {
	var d openapi_types.Date
	fields := FileUploadFields{
		OrderId:            1,
		RawFile:            openapi_types.File{},
		Base64Field:        []byte("data"),
		ExplicitFormatWins: &d,
	}
	assert.NotNil(t, fields.ExplicitFormatWins)
}

// ----------------------------------------------------------------------------
// enum-via-oneOf idiom (from enum_via_oneof)
// ----------------------------------------------------------------------------

// An integer enum-via-oneOf produces `type Severity int` with the right
// per-branch constant values.
func TestSeverityConstants(t *testing.T) {
	assert.Equal(t, 2, int(HIGH))
	assert.Equal(t, 1, int(MEDIUM))
	assert.Equal(t, 0, int(LOW))
}

// Severity marshals as its integer value, not a wrapped union or a string.
func TestSeverityJSONRoundTrip(t *testing.T) {
	data, err := json.Marshal(HIGH)
	require.NoError(t, err)
	assert.JSONEq(t, `2`, string(data))

	var got Severity
	require.NoError(t, json.Unmarshal([]byte(`1`), &got))
	assert.Equal(t, MEDIUM, got)
}

// A string enum-via-oneOf produces `type Color string` with the right
// per-branch constant values.
func TestColorConstants(t *testing.T) {
	assert.Equal(t, "r", string(Red))
	assert.Equal(t, "g", string(Green))
	assert.Equal(t, "b", string(Blue))
}

// Color marshals as its string value.
func TestColorJSONRoundTrip(t *testing.T) {
	data, err := json.Marshal(Red)
	require.NoError(t, err)
	assert.JSONEq(t, `"r"`, string(data))

	var got Color
	require.NoError(t, json.Unmarshal([]byte(`"b"`), &got))
	assert.Equal(t, Blue, got)
}

// Negative path: a oneOf where any branch lacks `title` must NOT trigger
// enum-via-oneOf detection. MixedOneOf is emitted by the standard handler as
// `type MixedOneOf = string` (an alias), so a plain string is directly
// assignable. If detection were over-eager, MixedOneOf would become a newtype
// and the assignment below would fail to compile.
//
// The explicit `var m MixedOneOf = s` declaration is intentional: staticcheck
// (ST1023) wants the type omitted because aliasing makes it redundant, but that
// redundancy IS the test -- it exercises the alias-vs-newtype property at
// compile time.
func TestMixedOneOfFallsThrough(t *testing.T) {
	var s = "anything"
	var m MixedOneOf = s //nolint:staticcheck // ST1023: explicit type is the compile-time alias check
	assert.Equal(t, "anything", string(m))
}

// ----------------------------------------------------------------------------
// 3.1 polish: const -> enum, examples -> doc comments (from openapi31_polish)
// ----------------------------------------------------------------------------

// A scalar `const` schema produces a typed alias and a singleton constant.
// Compile-time check: `Active` is declared as `const Active Status = "active"`,
// so type inference gives `s` the type `Status`.
func TestStatusConstSchema(t *testing.T) {
	s := Active
	assert.Equal(t, "active", string(s))
	assert.True(t, s.Valid(), "Active should be a valid Status enum member")
}

// `examples:` on a property surfaces in the generated field's Go doc comment.
// Doc comments aren't runtime-introspectable, so this parses the generated
// source file and asserts each field's doc-comment text.
func TestPetExampleComments(t *testing.T) {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "openapi31.gen.go", nil, parser.ParseComments)
	require.NoError(t, err, "could not parse generated file")

	fields := petFieldComments(t, f)

	// `name` had description="The pet's name." plus two examples; both
	// must appear in the field doc.
	require.Contains(t, fields, "Name")
	assert.Contains(t, fields["Name"], "The pet's name.",
		"Name field should preserve the original description")
	assert.Contains(t, fields["Name"], "Examples: Whiskers, Rex",
		"Name field should surface the plural examples on its own paragraph")

	// `lives` had no description, only an example value.
	require.Contains(t, fields, "Lives")
	assert.Contains(t, fields["Lives"], "Examples: 9",
		"Lives field should surface the integer example as a doc fragment")
}

// petFieldComments extracts the doc comment text for each field of the Pet
// struct from a parsed AST. Returns map[fieldName]commentText.
func petFieldComments(t *testing.T, f *ast.File) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name.Name != "Pet" {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, field := range st.Fields.List {
				if len(field.Names) == 0 || field.Doc == nil {
					continue
				}
				out[field.Names[0].Name] = strings.TrimSpace(field.Doc.Text())
			}
		}
	}
	return out
}
