// Package optionsresponsegettersskipped checks that, with skip-response-body-getters
// set to true, the generated client response wrapper exposes NEITHER GetBody() nor any
// typed getter — only the response struct itself is generated.
//
// outputoptions/response-body-getters/skipped
package optionsresponsegettersskipped

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml ../spec.yaml
