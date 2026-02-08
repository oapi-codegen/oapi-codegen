// Package issue_1212 tests multi-package response schemas.
// https://github.com/oapi-codegen/oapi-codegen/issues/1212
package issue_1212

//go:generate go run ../../../../../../cmd/oapi-codegen -config pkg2/config.yaml pkg2.yaml
//go:generate go run ../../../../../../cmd/oapi-codegen -config pkg1/config.yaml pkg1.yaml
