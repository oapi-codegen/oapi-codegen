// Package openapi31_content_keywords verifies that the OpenAPI 3.1
// `contentEncoding` and `contentMediaType` keywords -- the JSON-Schema-
// aligned replacements for the 3.0 `format: binary` / `format: byte`
// idioms -- generate the same Go shapes (`openapi_types.File`, `[]byte`)
// their 3.0 counterparts did. Without coverage, a user following the
// 3.1 upgrade guide's "Update file upload descriptions" step would
// silently lose their file/binary typing.
package openapi31_content_keywords

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
