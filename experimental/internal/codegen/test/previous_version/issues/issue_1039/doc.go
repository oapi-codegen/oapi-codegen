// Package issue_1039 tests nullable type generation.
// https://github.com/oapi-codegen/oapi-codegen/issues/1039
package issue_1039

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
