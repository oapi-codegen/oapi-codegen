// Package textandjson verifies strict-server codegen for an operation
// mixing text/plain and JSON responses while nullable-type is enabled:
// the generated Visit* method for the text/plain response must compile
// and write the plain-text body.
//
// From issue-2190.
package textandjson

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
