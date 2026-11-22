// Package responseheaders verifies that response headers declared on an
// operation are emitted as a Headers struct on the strict-server
// (gorilla) response object and written by the generated Visit* method.
// Compile-time assertion only.
//
// From issue-1676.
package responseheaders

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
