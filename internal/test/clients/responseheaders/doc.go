// Package responseheaders verifies that response headers declared in the
// spec are parsed into typed Headers<StatusCode> fields on the generated
// client response wrappers. The header structs are client-local types named
// after the response wrapper (e.g. GetFooResponse200Headers), emitted and
// consumed entirely within the client output; specs without response
// headers generate no additional code.
//
// From issue-2011.
package responseheaders

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
