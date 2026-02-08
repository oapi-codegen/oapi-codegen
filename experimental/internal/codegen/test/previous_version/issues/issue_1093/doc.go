// Package issue_1093 tests multi-spec cross-package imports.
// https://github.com/oapi-codegen/oapi-codegen/issues/1093
package issue_1093

//go:generate go run ../../../../../../cmd/oapi-codegen -config parent.config.yaml parent.api.yaml
//go:generate go run ../../../../../../cmd/oapi-codegen -config child.config.yaml child.api.yaml
