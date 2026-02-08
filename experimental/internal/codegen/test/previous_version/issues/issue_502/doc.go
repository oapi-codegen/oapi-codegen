// Package issue_502 tests that anyOf with only one ref generates the referenced type.
// https://github.com/oapi-codegen/oapi-codegen/issues/502
package issue_502

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
