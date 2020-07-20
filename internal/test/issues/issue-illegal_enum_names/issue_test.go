package illegal_enum_names

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/codegen"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestIllegalEnumNames(t *testing.T) {
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile("spec.yaml")
	require.NoError(t, err)

	opts := codegen.Options{
		GenerateClient:     true,
		GenerateEchoServer: true,
		GenerateTypes:      true,
		EmbedSpec:          true,
	}

	code, err := codegen.Generate(swagger, "illegal_enum_names", opts)
	require.NoError(t, err)

	f, err := parser.ParseFile(token.NewFileSet(), "", code, parser.AllErrors)
	require.NoError(t, err)

	constDefs := make(map[string]string)
	for _, d := range f.Decls {
		switch decl := d.(type) {
		case *ast.GenDecl:
			if token.CONST == decl.Tok {
				for _, s := range decl.Specs {
					switch spec := s.(type) {
					case *ast.ValueSpec:
						constDefs[spec.Names[0].Name] = spec.Names[0].Obj.Decl.(*ast.ValueSpec).Values[0].(*ast.BasicLit).Value
					}
				}
			}
		}
	}

	require.Equal(t, `"Bar"`, constDefs["Bar_Bar"])
	require.Equal(t, `"Foo"`, constDefs["Bar_Foo"])
	require.Equal(t, `"Foo Bar"`, constDefs["Bar_Foo_Bar"])
	require.Equal(t, `"Foo-Bar"`, constDefs["Bar_Foo_Bar1"])
	require.Equal(t, `"1Foo"`, constDefs["Bar__Foo"])
	require.Equal(t, `" Foo"`, constDefs["Bar__Foo1"])
	require.Equal(t, `" Foo "`, constDefs["Bar__Foo_"])
	require.Equal(t, `"_Foo_"`, constDefs["Bar__Foo_1"])
}
