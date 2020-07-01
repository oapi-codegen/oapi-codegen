package duplicate_type_names

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/codegen"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestDuplicateTypeNames(t *testing.T) {
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile("spec.yaml")
	require.NoError(t, err)

	opts := codegen.Options{
		GenerateClient:     true,
		GenerateEchoServer: true,
		GenerateTypes:      true,
		EmbedSpec:          true,
	}

	code, err := codegen.Generate(swagger, "duplicate_type_names", opts)
	require.NoError(t, err)

	f, err := parser.ParseFile(token.NewFileSet(), "", code, parser.AllErrors)
	require.NoError(t, err)

	dupCheck := make(map[string]int)
	for _, d := range f.Decls {
		switch decl := d.(type) {
		case *ast.GenDecl:
			if token.TYPE == decl.Tok {
				for _, s := range decl.Specs {
					switch spec := s.(type) {
					case *ast.TypeSpec:
						dupCheck[spec.Name.Name]++

						require.Equal(t, 1, dupCheck[spec.Name.Name])
					}
				}
			}
		}
	}
}
