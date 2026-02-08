// Package issue_1219 tests additionalProperties merge with allOf.
// https://github.com/oapi-codegen/oapi-codegen/issues/1219
package issue_1219

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
