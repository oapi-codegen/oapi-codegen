// Package issue_1825 tests overlay and external refs.
// https://github.com/oapi-codegen/oapi-codegen/issues/1825
package issue_1825

//go:generate go run ../../../../../../cmd/oapi-codegen -config packageA/config.yaml packageA/spec.yaml
//go:generate go run ../../../../../../cmd/oapi-codegen -config spec/config.yaml spec/spec.yaml
