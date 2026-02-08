// Package issue_52 tests that recursive types are handled properly.
// https://github.com/oapi-codegen/oapi-codegen/issues/52
package issue_52

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
