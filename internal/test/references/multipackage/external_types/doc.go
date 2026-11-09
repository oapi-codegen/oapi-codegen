// Package externaltypes is a regression fixture for types that come from an
// imported package through import-mapping but are not named types of it. A
// reusable response reached from a non-fixed status code, and an operation of
// a path item reused by $ref, carry the Go types of the imported document:
// builtins ("any"), pointers ("*string"), and composite types that name the
// imported package's models ("[]Thing"). Only the model names may be
// qualified with the imported package, not the whole type.
package externaltypes

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.common.yaml common/spec.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.api.yaml spec.yaml
