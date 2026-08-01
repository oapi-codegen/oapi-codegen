// Package openapi31 verifies that the generator DETECTS OpenAPI 3.1-specific
// authoring idioms and lowers each to the correct Go shape:
//
//   - contentEncoding / contentMediaType file-upload keywords (the
//     JSON-Schema-aligned replacements for 3.0's `format: binary` / `format:
//     byte`) -> openapi_types.File, []byte; an explicit `format` always wins.
//   - the enum-via-oneOf idiom: a scalar schema whose oneOf branches each carry
//     `title` + `const` -> a Go typed enum with named constants; a near-miss
//     (one branch missing `title`) falls through to a plain union type alias.
//   - the smaller 3.1 polish features: `const` on a scalar -> typed alias +
//     singleton constant; plural `examples` -> Go doc comments.
//
// Features that merely exist in both 3.0 and 3.1 (e.g. nullable) are NOT here --
// they live in their feature category with mixed-version specs.
package openapi31

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
