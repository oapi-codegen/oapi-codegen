// Package issue_1710 tests that fields are not lost in nested allOf oneOf structures.
// https://github.com/oapi-codegen/oapi-codegen/issues/1710
package issue_1710

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
