// Package issue_1182 tests external response refs across specs.
// https://github.com/oapi-codegen/oapi-codegen/issues/1182
package issue_1182

//go:generate go run ../../../../../../cmd/oapi-codegen -config pkg2/config.yaml pkg2.yaml
//go:generate go run ../../../../../../cmd/oapi-codegen -config pkg1/config.yaml pkg1.yaml
