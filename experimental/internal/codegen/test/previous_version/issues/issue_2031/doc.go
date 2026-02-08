// Package issue_2031 tests skip-optional-pointer with arrays.
// https://github.com/oapi-codegen/oapi-codegen/issues/2031
package issue_2031

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
