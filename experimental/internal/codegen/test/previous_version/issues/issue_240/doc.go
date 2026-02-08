// Package issue_240 tests models with no type field.
// https://github.com/oapi-codegen/oapi-codegen/issues/240
package issue_240

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
