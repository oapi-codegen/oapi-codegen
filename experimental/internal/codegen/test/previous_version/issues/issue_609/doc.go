// Package issue_609 tests optional field with no type info.
// https://github.com/oapi-codegen/oapi-codegen/issues/609
package issue_609

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
