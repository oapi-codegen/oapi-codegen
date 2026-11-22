// Package schemasnullable exercises nullable type generation across multiple
// scenarios: the required/optional x nullable matrix and client/server stubs
// with nullable-type:true (issue #1039); array of nullable items using
// nullable.Nullable[T] (issue #2185); equivalence of the OpenAPI 3.0
// nullable:true idiom vs. the OpenAPI 3.1 type-array idiom (openapi31_nullable);
// and bare OpenAPI 3.1 `type: "null"` schemas mapping to Go `any` (issue #2430).
package schemasnullable

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_defaultbehaviour.yaml spec.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_30.yaml spec_30.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_31.yaml spec_31.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_schemas.yaml spec_schemas.yaml
