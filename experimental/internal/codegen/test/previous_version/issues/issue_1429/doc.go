// Package issue_1429 tests that enums inside anyOf members are generated.
// https://github.com/oapi-codegen/oapi-codegen/issues/1429
package issue_1429

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
