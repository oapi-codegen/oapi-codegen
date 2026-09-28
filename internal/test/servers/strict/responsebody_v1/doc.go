// Package serversstrictresponsebodyv1 exercises strict responses composed from
// an untyped schema under schema-merging-behavior v1.
//
// v1 embeds an allOf member in a struct instead of lowering the allOf to the
// member's type, and a struct is a valid method receiver, so these responses
// keep the direct type they have always had. A nullable union collapsed to the
// untyped member is an interface type under every behavior, so it carries its
// value in a Body field.
//
// This package records how v1 behaves. A bug fix may change it, but it must
// not be changed to accept a regression, and a commit that changes it must
// say why.
package serversstrictresponsebodyv1

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
