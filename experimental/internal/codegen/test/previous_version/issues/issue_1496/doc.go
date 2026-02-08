// Package issue_1496 tests that inline schemas generate valid Go identifiers.
// https://github.com/oapi-codegen/oapi-codegen/issues/1496
package issue_1496

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
