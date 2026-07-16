// Package schemassecurity exercises OpenAPI security scheme handling: specifically
// that a global security reference naming a scheme not defined under
// components/securitySchemes is tolerated by the code generator and produces
// a package that compiles without error.
package schemassecurity

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
