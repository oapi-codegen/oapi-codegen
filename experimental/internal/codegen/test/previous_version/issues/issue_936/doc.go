// Package issue_936 tests recursive/circular schema references.
// https://github.com/oapi-codegen/oapi-codegen/issues/936
package issue_936

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
