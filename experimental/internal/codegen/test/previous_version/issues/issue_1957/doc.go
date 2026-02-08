// Package issue_1957 tests x-go-type with skip-optional-pointer.
// https://github.com/oapi-codegen/oapi-codegen/issues/1957
package issue_1957

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
