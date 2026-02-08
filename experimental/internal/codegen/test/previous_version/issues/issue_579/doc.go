// Package issue_579 tests aliased types with date format.
// https://github.com/oapi-codegen/oapi-codegen/issues/579
package issue_579

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
