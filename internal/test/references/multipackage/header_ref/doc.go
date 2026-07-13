// Package headerref is a regression test for issue-2060: a response header
// that is an external `$ref` whose schema is itself a `$ref` to a named type
// must qualify that type with the imported package (e.g.
// externalRef0.ETagSchema), both in the strict-server response headers struct
// and in the typed client response wrappers. Before the fix the header field
// referenced an undefined local type and the generated package failed to
// compile.
package headerref

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.common.yaml common/spec.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.api.yaml spec.yaml
