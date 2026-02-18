package typemapping

// The configuration in this directory overrides the default handling of
// "type: number" from producing an `int` to producing an `int64`, and we
// override `type: string, format: date` to be a custom type in this package.
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
