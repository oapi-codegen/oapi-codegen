// Package optionsunexpectedresponseskipped checks that, with
// client-response-error-on-unexpected-response unset (the default), the
// generated Parse<Operation>Response functions have no default case and an
// unexpected response yields a response object with a nil error.
//
// outputoptions/unexpected-response/skipped
package optionsunexpectedresponseskipped

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml ../spec.yaml
