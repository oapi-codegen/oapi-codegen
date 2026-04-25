// Package openapi31_polish verifies the OpenAPI 3.1 polish features:
// `examples` (plural array) propagating into Go doc comments, and
// scalar `const` synthesizing a typed alias + singleton constant via
// the existing enum-codegen path.
package openapi31_polish

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
