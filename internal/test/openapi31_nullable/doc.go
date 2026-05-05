// Package openapi31_nullable verifies that the OpenAPI 3.1 type-array
// nullable idiom (`type: ["X","null"]`) produces the same Go shape as the
// OpenAPI 3.0 `nullable: true` idiom. The generated types are emitted in
// the spec_3_0/ and spec_3_1/ subpackages and exercised by the
// instantiation tests in this directory.
package openapi31_nullable

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_3_0.yaml spec_3_0.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_3_1.yaml spec_3_1.yaml
