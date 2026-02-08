// Package issue_775 tests that allOf with format specification works correctly.
// https://github.com/oapi-codegen/oapi-codegen/issues/775
package issue_775

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
