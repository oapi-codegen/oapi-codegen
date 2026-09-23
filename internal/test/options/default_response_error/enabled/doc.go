// Package optionsdefaultresponseerrorenabled checks that, with
// default-response-error enabled, the generated Parse<Op>Response function
// returns an error for a response whose status code and content type are not
// matched by any declared case clause.
//
// outputoptions/default-response-error/enabled
package optionsdefaultresponseerrorenabled

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml ../spec.yaml
