// Package schemasobjects exercises object schemas with additionalProperties, including
// additionalProperties:true inside an array item's inline schema (the strict-server path,
// issue-1277), plus the additionalProperties / readOnly+writeOnly / raw-JSON object cases
// folded from the components.yaml kitchen sink.
package schemasobjects

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
