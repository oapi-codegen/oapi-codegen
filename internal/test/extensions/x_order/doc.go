// Package extensionsxorder exercises the x-order extension, which controls
// the ordering of struct fields in generated Go types. Properties annotated
// with x-order are emitted before unannotated ones, and properties with
// lower x-order values appear before those with higher values.
package extensionsxorder

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
