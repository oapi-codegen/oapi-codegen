// Package issue_1298 tests custom content-type schemas.
// https://github.com/oapi-codegen/oapi-codegen/issues/1298
package issue_1298

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
