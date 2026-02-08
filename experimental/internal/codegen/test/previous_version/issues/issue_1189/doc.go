// Package issue_1189 tests anyOf/allOf/oneOf composition.
// https://github.com/oapi-codegen/oapi-codegen/issues/1189
package issue_1189

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
