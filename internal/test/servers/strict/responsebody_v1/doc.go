// Package serversstrictresponsebodyv1 exercises strict responses composed from
// an untyped schema under schema-merging-behavior v1.
//
// An allOf of a single untyped $ref, next to members that add no fields, is
// the untyped member's type under v1 as under v2 (issue #2590), and so is a
// nullable union collapsed to it. That type is an interface, so each response
// carries its value in a Body field, and the value goes on the wire as it is.
// v1 used to embed the member in a struct instead, which encoded the value
// under a key named after the type ({"Untyped": ...}).
//
// This package records how v1 behaves. A bug fix may change it, but it must
// not be changed to accept a regression, and a commit that changes it must
// say why.
package serversstrictresponsebodyv1

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
