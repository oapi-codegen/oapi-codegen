// Package issue_697 tests that properties alongside allOf are included.
// https://github.com/oapi-codegen/oapi-codegen/issues/697
package issue_697

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
