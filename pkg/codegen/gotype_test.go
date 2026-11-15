package codegen

import (
	"go/ast"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeIdents(t *testing.T) {
	for decl, want := range map[string][]string{
		"Foo":                     {"Foo"},
		"(Foo)":                   {"Foo"},
		"*[]map[Key]Value":        {"Key", "Value"},
		"[N]Elem":                 {"Elem"},
		"pkg.Name":                nil,
		"Generic[Arg, pkg.Other]": {"Generic", "Arg"},
		"struct {\n\tF *Foo `json:\"f\"`\n\tG string\n\tEmbedded\n}": {"Foo", "string", "Embedded"},
		"interface {\n\tM(In) Out\n\tBase\n}":                        {"In", "Out", "Base"},
		"func(A, ...B) (C, error)":                                   {"A", "B", "C", "error"},
		"chan<- Msg":                                                 {"Msg"},
	} {
		t.Run(decl, func(t *testing.T) {
			expr, _, ok := parseGoType(decl)
			require.True(t, ok)
			var got []string
			typeIdents(expr, func(id *ast.Ident) bool {
				got = append(got, id.Name)
				return true
			})
			assert.Equal(t, want, got)
		})
	}
}

func TestTypeIdentsStops(t *testing.T) {
	expr, _, ok := parseGoType("map[A]B")
	require.True(t, ok)
	var got []string
	typeIdents(expr, func(id *ast.Ident) bool {
		got = append(got, id.Name)
		return false
	})
	assert.Equal(t, []string{"A"}, got)
}

func TestParseGoType(t *testing.T) {
	_, _, ok := parseGoType("not a type (")
	assert.False(t, ok)

	expr, _, ok := parseGoType("((*Foo))")
	require.True(t, ok)
	_, isStar := unparen(expr).(*ast.StarExpr)
	assert.True(t, isStar)
}
