package codegen

// This file holds what the generator reads out of a Go type expression: the
// one a schema is generated as, or the one x-go-type carries.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
)

// parseGoType parses a Go type expression such as "[]Foo", "*string" or
// "struct { ... }". fset locates the expression's nodes in decl. ok is false
// when decl doesn't parse as one, which a caller treats as a type it can't
// see into.
func parseGoType(decl string) (expr ast.Expr, fset *token.FileSet, ok bool) {
	fset = token.NewFileSet()
	expr, err := parser.ParseExprFrom(fset, "", decl, 0)
	if err != nil {
		return nil, nil, false
	}
	return expr, fset, true
}

// unparen returns expr without the parentheses around it.
func unparen(expr ast.Expr) ast.Expr {
	for {
		paren, ok := expr.(*ast.ParenExpr)
		if !ok {
			return expr
		}
		expr = paren.X
	}
}

// typeIdents calls fn for each identifier in a type position of expr, in
// source order, until fn returns false: the T of []T, *T or map[K]T, the type
// of a field or a parameter, a type argument. Field and method names, the
// package and name of a qualified pkg.Name, and an array's length are not
// types.
func typeIdents(expr ast.Expr, fn func(*ast.Ident) bool) {
	more := true
	var visit func(ast.Node) bool
	visit = func(n ast.Node) bool {
		if !more {
			return false
		}
		switch n := n.(type) {
		case *ast.Field:
			ast.Inspect(n.Type, visit)
			return false
		case *ast.ArrayType:
			ast.Inspect(n.Elt, visit)
			return false
		case *ast.SelectorExpr:
			return false
		case *ast.Ident:
			more = fn(n)
		}
		return more
	}
	ast.Inspect(expr, visit)
}

// isPredeclaredType reports whether name is one of Go's predeclared types,
// such as "string", "any" or "error".
func isPredeclaredType(name string) bool {
	_, ok := types.Universe.Lookup(name).(*types.TypeName)
	return ok
}
