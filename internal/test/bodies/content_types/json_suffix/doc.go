// Package jsonsuffix verifies that a request body declared with a
// +json content-type suffix (application/test+json) round-trips through
// the generated client and gin strict-server bindings.
//
// From issue-1298.
package jsonsuffix

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
