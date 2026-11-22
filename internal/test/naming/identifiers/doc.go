// Package namingidentifiers exercises identifier normalisation for names that would be
// invalid or ambiguous as Go identifiers: operation IDs that begin with a digit (e.g.
// "3GPPFoo"), HTTP response-header / component names that begin with a digit ("000-foo" →
// N000Foo, "200" → N200Response), struct-field names beginning with an underscore
// ("_id" → UnderscoreId), reserved keywords as path params ("fallthrough"), digit-leading
// path params ("1param") and schema names ("5StartsWithNumber"), $ref-declared params, and
// hyphenated security-scheme names (access-token). The schemas-fold cases are compile-only.
package namingidentifiers

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_schemas.yaml spec_schemas.yaml
