package enumconflicts

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/codegen"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestEnumConflicts(t *testing.T) {
	swagger, err := openapi3.NewLoader().LoadFromFile("spec.yaml")
	require.NoError(t, err)

	opts := codegen.Configuration{
		PackageName: "enumconflicts",
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

	require.Equal(t, `"Orange"`, constDefs["ColorOrange"])
	require.Equal(t, `"Red"`, constDefs["ColorRed"])
	require.Equal(t, `"Fruit"`, constDefs["FoodGroupFruit"])
	require.Equal(t, `"Vegetable"`, constDefs["FoodGroupVegetable"])
	require.Equal(t, `"Apple"`, constDefs["FruitApple"])
	require.Equal(t, `"Banana"`, constDefs["FruitBanana"])
	require.Equal(t, `"Orange"`, constDefs["FruitOrange"])
	require.Equal(t, `"Carrot"`, constDefs["Carrot"])
	require.Equal(t, `"Potato"`, constDefs["Potato"])
	require.Equal(t, `"Goulburn"`, constDefs["TownGoulburn"])
	require.Equal(t, `"Orange"`, constDefs["TownOrange"])
	require.Equal(t, `"Parks"`, constDefs["TownParks"])
}
