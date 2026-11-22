// Package schemasenums exercises enum sanitization: codegen must produce valid Go
// identifiers for enum values that are empty strings, contain spaces, hyphens,
// leading digits, leading/trailing underscores, or duplicate after normalization.
package schemasenums

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
