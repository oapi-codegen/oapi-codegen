// Package optionsdefaultresponseerrordefault checks that, with
// default-response-error left at its default, the generated Parse<Op>Response
// function returns an unmatched response with no error, preserving the existing
// behaviour.
//
// outputoptions/default-response-error/default
package optionsdefaultresponseerrordefault

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml ../spec.yaml
