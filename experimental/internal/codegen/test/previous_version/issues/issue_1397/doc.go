// Package issue_1397 tests basic type generation with x-go-type-name.
// https://github.com/oapi-codegen/oapi-codegen/issues/1397
package issue_1397

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
