// Package issue_1168 tests additionalProperties: true.
// https://github.com/oapi-codegen/oapi-codegen/issues/1168
package issue_1168

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
