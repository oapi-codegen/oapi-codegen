// Package extensionsxgotype exercises x-go-type and x-go-type-import for
// overriding generated Go types with external package types; x-go-type-name
// for renaming inline enum types and nested object types; and x-go-type
// combined with x-go-type-skip-optional-pointer so that an overridden optional
// field is emitted as a non-pointer value type.
package extensionsxgotype

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
