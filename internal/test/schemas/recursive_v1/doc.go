// Package schemasrecursivev1 exercises the recursive allOf compositions of
// schemas/recursive under schema-merging-behavior v1: the shapes of issue
// #2542, whose members refer back into a body still being generated, and a
// composition that refers back to itself through a whole-document $ref
// (tree.yaml), which v1 alone used to overflow the stack on.
//
// v1 embeds a $ref member instead of merging it, so the compositions of
// #2542 close their recursion on the embedded type, inside the anonymous
// struct v1 has always generated. The whole-document $ref names no Go type,
// so its value is generated inline, and the composition becomes a named type
// that refers to itself through the slice, as it does under v2.
//
// This package records how v1 behaves. A bug fix may change it, but it must
// not be changed to accept a regression, and a commit that changes it must
// say why.
package schemasrecursivev1

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
