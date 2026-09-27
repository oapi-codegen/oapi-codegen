// Package serversstrictresponsebody verifies that OpenAPI 3.1 schemas lowered
// to any use compilable strict response bodies and preserve their JSON values.
package serversstrictresponsebody

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
