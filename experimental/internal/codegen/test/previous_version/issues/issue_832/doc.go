// Package issue_832 tests x-go-type-name override for enum types.
// https://github.com/oapi-codegen/oapi-codegen/issues/832
package issue_832

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
