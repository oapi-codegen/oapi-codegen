// Package schemasdeprecated verifies deprecation handling across every surface:
//
//   - issue #975: deprecated FIELDS get a `// Deprecated:` doc comment, with or
//     without a description, and x-deprecated-reason is surfaced.
//   - deprecation/ (comprehensive): deprecated operations, parameters
//     (query/path/header/cookie), request bodies, response headers, and whole
//     deprecated objects; plus the negative case where x-deprecated-reason is
//     set without deprecated:true (must NOT generate).
//
// Compile-only: a regression would fail to generate or build.
package schemasdeprecated

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
