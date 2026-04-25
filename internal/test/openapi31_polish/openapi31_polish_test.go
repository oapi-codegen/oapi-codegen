// Package openapi31_polish tests two OpenAPI 3.1 polish features:
//
//   - `examples:` (plural array) folding into Go doc comments. Doc
//     comments aren't runtime-introspectable, so the test parses the
//     generated source file and asserts the expected text appears in
//     the type's field comments.
//
//   - Scalar `const:` becoming a typed alias plus a singleton constant
//     via the existing enum-codegen path. The test exercises this by
//     instantiating the typed value (compile-time guarantee that the
//     type and constant exist) and asserting the constant's value.
package openapi31_polish

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStatusConstSchema verifies that a scalar `const` schema produces a
// typed alias and a singleton constant. Compile-time check: `Active` is
// declared as `const Active Status = "active"`, so type inference here
// gives `s` the type `Status` -- if the codegen had emitted Active as
// untyped, this would not compile.
func TestStatusConstSchema(t *testing.T) {
	s := Active
	assert.Equal(t, "active", string(s))
	assert.True(t, s.Valid(), "Active should be a valid Status enum member")
}

// TestPetExampleComments verifies that `examples:` on a property
// surfaces in the generated field's Go doc comment. We can't introspect
// doc comments at runtime, so this parses the generated source file and
// asserts each field's doc-comment text.
func TestPetExampleComments(t *testing.T) {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "openapi31_polish.gen.go", nil, parser.ParseComments)
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

// petFieldComments extracts the doc comment text for each field of the
// Pet struct from a parsed AST. Returns map[fieldName]commentText.
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
