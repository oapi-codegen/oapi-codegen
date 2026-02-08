// Package issue_193 tests allOf with additionalProperties merging.
// https://github.com/oapi-codegen/oapi-codegen/issues/193
package issue_193

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
