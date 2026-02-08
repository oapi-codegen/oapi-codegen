// Package issue_1029 tests that oneOf with multiple single-value string enums generates valid code.
// https://github.com/oapi-codegen/oapi-codegen/issues/1029
package issue_1029

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
