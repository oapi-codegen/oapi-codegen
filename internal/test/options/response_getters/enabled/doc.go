// Package optionsresponsegettersenabled checks that, with skip-response-body-getters
// unset (the default), the generated client response wrapper exposes GetBody() and a
// typed getter per response field.
//
// outputoptions/response-body-getters/enabled
package optionsresponsegettersenabled

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml ../spec.yaml
