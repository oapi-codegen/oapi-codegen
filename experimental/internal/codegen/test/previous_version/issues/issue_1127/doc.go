// Package issue_1127 tests multiple content types.
// https://github.com/oapi-codegen/oapi-codegen/issues/1127
package issue_1127

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
