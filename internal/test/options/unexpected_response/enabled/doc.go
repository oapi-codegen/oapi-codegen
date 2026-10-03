// Package optionsunexpectedresponseenabled checks that, with
// client-response-error-on-unexpected-response set to true, the generated
// Parse<Operation>Response functions gain a default case that returns the
// sentinel ErrUnexpectedResponse when the response matches none of the
// declared responses.
//
// outputoptions/unexpected-response/enabled
package optionsunexpectedresponseenabled

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml ../spec.yaml
