// Package issue_1767 tests underscore field name mapping.
// https://github.com/oapi-codegen/oapi-codegen/issues/1767
package issue_1767

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
