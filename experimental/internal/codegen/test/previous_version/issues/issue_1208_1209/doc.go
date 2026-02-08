// Package issue_1208_1209 tests multiple JSON content types.
// https://github.com/oapi-codegen/oapi-codegen/issues/1208
// https://github.com/oapi-codegen/oapi-codegen/issues/1209
package issue_1208_1209

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
