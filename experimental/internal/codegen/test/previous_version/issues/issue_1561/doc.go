// Package issue_1561 tests skip-optional-pointer on containers.
// https://github.com/oapi-codegen/oapi-codegen/issues/1561
package issue_1561

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
