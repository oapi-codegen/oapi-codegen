// Package optionscompatibility exercises the compatibility output options.
// It tests that preserve-original-operation-id-casing-in-embedded-spec
// keeps the raw operationId string from the spec in the embedded spec blob,
// verifying that mixed-case and kebab+camel+snake operationIds are preserved
// exactly as written in the OpenAPI spec.
package optionscompatibility

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
