// Package issue_2102 tests that properties defined at the same level as allOf are included.
// https://github.com/oapi-codegen/oapi-codegen/issues/2102
package issue_2102

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
