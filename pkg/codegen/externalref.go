package codegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"slices"
	"strings"
)

// ensureExternalRefsInRequestBodyDefinitions ensures that when an externalRef (`$ref` that points to a file that isn't the current spec) is encountered, we make sure we update our underlying `RefType` to make sure that we point to that type.
// This only happens if we have a non-empty `ref` passed in, and that `ref` isn't pointing to something in our file
// NOTE that the pointer here allows us to pass in a reference and edit in-place
func ensureExternalRefsInRequestBodyDefinitions(defs *[]RequestBodyDefinition, ref string) {
	if ref == "" {
		return
	}

	for i, rbd := range *defs {
		ensureExternalRefsInSchema(&rbd.Schema, ref)

		// make sure we then update it in-place
		(*defs)[i] = rbd
	}
}

// ensureExternalRefsInResponseDefinitions ensures that when an externalRef (`$ref` that points to a file that isn't the current spec) is encountered, we make sure we update our underlying `RefType` to make sure that we point to that type.
// This only happens if we have a non-empty `ref` passed in, and that `ref` isn't pointing to something in our file
// NOTE that the pointer here allows us to pass in a reference and edit in-place
func ensureExternalRefsInResponseDefinitions(defs *[]ResponseDefinition, ref string) {
	if ref == "" {
		return
	}

	for i, rd := range *defs {
		for j, rcd := range rd.Contents {
			ensureExternalRefsInSchema(&rcd.Schema, ref)

			// make sure we then update it in-place
			rd.Contents[j] = rcd
		}
		for j, rhd := range rd.Headers {
			// A header that is itself a $ref into another document was
			// already qualified with that document's package.
			if rhd.Schema.IsExternalRef() {
				continue
			}
			ensureExternalRefsInSchema(&rhd.Schema, ref)
			rd.Headers[j] = rhd
		}

		// make sure we then update it in-place
		(*defs)[i] = rd
	}
}

// ensureExternalRefsInParameterDefinitions ensures that when an externalRef (`$ref` that points to a file that isn't the current spec) is encountered, we make sure we update our underlying `RefType` to make sure that we point to that type.
// This only happens if we have a non-empty `ref` passed in, and that `ref` isn't pointing to something in our file
// NOTE that the pointer here allows us to pass in a reference and edit in-place
func ensureExternalRefsInParameterDefinitions(defs *[]ParameterDefinition, ref string) {
	if ref == "" {
		return
	}

	for i, pd := range *defs {
		ensureExternalRefsInSchema(&pd.Schema, ref)

		// make sure we then update it in-place
		(*defs)[i] = pd
	}
}

// externalPackageFor returns the imported Go package name for a `$ref` that
// targets a file outside the current spec. Returns an empty string when the
// ref is empty, points within the current spec, or has no import-mapping.
func externalPackageFor(ref string) string {
	if ref == "" {
		return ""
	}
	parts := strings.SplitN(ref, "#", 2)
	if pack, ok := globalState.importMapping[parts[0]]; ok {
		return pack.Name
	}
	return ""
}

// ensureExternalRefsInSchema ensures that when an externalRef (`$ref` that points to a file that isn't the current spec) is encountered, we make sure we update our underlying `RefType` to make sure that we point to that type.
//
// This only happens if we have a non-empty `ref` passed in, and that `ref` isn't pointing to something in our file
//
// The schema's Go type was generated in the context of the external document,
// so the type names in it are declared by the imported package. Only those
// names are qualified: builtins, names already qualified with a package, and
// the syntax around them (`[]`, `*`, `map[...]`, struct literals) are kept.
//
// NOTE that the pointer here allows us to pass in a reference and edit in-place
func ensureExternalRefsInSchema(schema *Schema, ref string) {
	pkg := externalPackageFor(ref)
	if pkg == "" {
		return
	}

	switch {
	case isExternalTypeName(schema.GoType):
		// A named type of the imported package is referred to as a $ref
		// to it would be. This wins over a RefType the schema already
		// carries, as it always has, so generated code keeps its types.
		schema.RefType = fmt.Sprintf("%s.%s", pkg, schema.GoType)
	case strings.HasPrefix(schema.GoType, "struct {"):
		// A struct literal is spelled out in place, so its fields must
		// name the imported package's types. One that the schema declares
		// as a named type is referred to by that name, unchanged.
		if schema.RefType == "" {
			schema.GoType = qualifyExternalTypeNames(schema.GoType, pkg)
		}
	case schema.RefType != "":
		schema.RefType = qualifyExternalTypeNames(schema.RefType, pkg)
	default:
		schema.GoType = qualifyExternalTypeNames(schema.GoType, pkg)
	}

	// Qualify union branch types. Each UnionElement was resolved as a
	// local reference by the union generator (e.g. "Bionicle"); when the
	// enclosing schema came from an externally-ref'd file the branches
	// actually live in the imported package, so the As/From/Merge
	// methods generated by union.tmpl must use "<pkg>.Bionicle" instead.
	for i, elem := range schema.UnionElements {
		s := string(elem)
		if !strings.Contains(s, ".") {
			schema.UnionElements[i] = UnionElement(fmt.Sprintf("%s.%s", pkg, s))
		}
	}
}

// isExternalTypeName reports whether goType is a single unqualified type name
// that is not a Go builtin, such as the name of a generated model.
func isExternalTypeName(goType string) bool {
	expr, err := parser.ParseExpr(goType)
	if err != nil {
		return false
	}
	ident, ok := expr.(*ast.Ident)
	return ok && !isPredeclaredType(ident.Name)
}

// qualifyExternalTypeNames qualifies with pkg every unqualified type name in
// the Go type expression typeExpr that is not a Go builtin, for example
// "[]Foo" becomes "[]pkg.Foo" and "struct { F *Foo `json:...` }" becomes
// "struct { F *pkg.Foo `json:...` }". Builtins ("any", "*string"), names that
// already carry a package ("time.Time") and everything else in typeExpr,
// including field names, tags and comments, are left as they are. An
// expression that does not parse is returned unchanged.
func qualifyExternalTypeNames(typeExpr, pkg string) string {
	fset := token.NewFileSet()
	expr, err := parser.ParseExprFrom(fset, "", typeExpr, 0)
	if err != nil {
		return typeExpr
	}

	var offsets []int
	var visit func(ast.Expr)
	visitFields := func(fields *ast.FieldList) {
		if fields == nil {
			return
		}
		for _, field := range fields.List {
			visit(field.Type)
		}
	}
	visit = func(e ast.Expr) {
		switch e := e.(type) {
		case *ast.Ident:
			if !isPredeclaredType(e.Name) {
				offsets = append(offsets, fset.Position(e.Pos()).Offset)
			}
		case *ast.SelectorExpr:
			// Already qualified with a package.
		case *ast.StarExpr:
			visit(e.X)
		case *ast.ParenExpr:
			visit(e.X)
		case *ast.ArrayType:
			visit(e.Elt)
		case *ast.Ellipsis:
			visit(e.Elt)
		case *ast.MapType:
			visit(e.Key)
			visit(e.Value)
		case *ast.ChanType:
			visit(e.Value)
		case *ast.IndexExpr:
			visit(e.X)
			visit(e.Index)
		case *ast.IndexListExpr:
			visit(e.X)
			for _, index := range e.Indices {
				visit(index)
			}
		case *ast.StructType:
			visitFields(e.Fields)
		case *ast.InterfaceType:
			visitFields(e.Methods)
		case *ast.FuncType:
			visitFields(e.Params)
			visitFields(e.Results)
		}
	}
	visit(expr)
	slices.Sort(offsets)

	qualified := typeExpr
	for _, offset := range slices.Backward(offsets) {
		qualified = qualified[:offset] + pkg + "." + qualified[offset:]
	}
	return qualified
}

// isPredeclaredType reports whether name is one of Go's predeclared types,
// such as "string", "any" or "error".
func isPredeclaredType(name string) bool {
	_, ok := types.Universe.Lookup(name).(*types.TypeName)
	return ok
}
