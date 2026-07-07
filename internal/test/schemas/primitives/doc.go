// Package schemasprimitives exercises primitive schema behaviours:
// aliased scalar types (e.g. Date aliased from openapi_types.Date, so that a
// $ref field and an inline field share the same underlying type) and untyped
// schemas (empty schema `{}`) that must generate interface{} without an
// optional pointer.
package schemasprimitives

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
