// Package issue_312 tests proper escaping of paths with special characters.
// https://github.com/oapi-codegen/oapi-codegen/issues/312
// This tests paths with colons like /pets:validate
package issue_312

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
