package illegal_enum_names

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/codegen"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIllegalEnumNames(t *testing.T) {
	swagger, err := openapi3.NewLoader().LoadFromFile("spec.yaml")
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
	assert.Equal(t, `"1"`, constDefs["Bar1"])
	assert.Equal(t, `"Bar"`, constDefs["BarBar"])
	assert.Equal(t, `"Foo"`, constDefs["BarFoo"])
	assert.Equal(t, `"Foo Bar"`, constDefs["BarFooBar"])
	assert.Equal(t, `"Foo-Bar"`, constDefs["BarFooBar1"])
	assert.Equal(t, `"1Foo"`, constDefs["Bar1Foo"])
	assert.Equal(t, `" Foo"`, constDefs["BarFoo1"])
	assert.Equal(t, `" Foo "`, constDefs["BarFoo2"])
	assert.Equal(t, `"_Foo_"`, constDefs["BarFoo3"])
}
