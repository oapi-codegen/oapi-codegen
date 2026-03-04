package illegalenumnames

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
	"github.com/stretchr/testify/require"
)

func TestIllegalEnumNames(t *testing.T) {
	swagger, err := openapi3.NewLoader().LoadFromFile("spec.yaml")
	require.NoError(t, err)

	opts := codegen.Configuration{
		PackageName: "illegalenumnames",
		Generate: codegen.GenerateOptions{
			EchoServer:   true,
			Client:       true,
			Models:       true,
			EmbeddedSpec: true,
		},
	}

	code, err := codegen.Generate(swagger, opts)
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

	require.Equal(t, `""`, constDefs["BarEmpty"])
	require.Equal(t, `"Bar"`, constDefs["BarBar"])
	require.Equal(t, `"Foo"`, constDefs["BarFoo"])
	require.Equal(t, `"Foo Bar"`, constDefs["BarFooBar"])
	require.Equal(t, `"Foo-Bar"`, constDefs["BarFooBar1"])
	require.Equal(t, `"1Foo"`, constDefs["BarN1Foo"])
	require.Equal(t, `" Foo"`, constDefs["BarFoo1"])
	require.Equal(t, `" Foo "`, constDefs["BarFoo2"])
	require.Equal(t, `"_Foo_"`, constDefs["BarUnderscoreFoo"])
	require.Equal(t, `"1"`, constDefs["BarN1"])
}
