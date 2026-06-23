// Regression fixture for https://github.com/oapi-codegen/oapi-codegen/issues/2010.
//
// The base spec defines components/responses/400 with a JSON body. The "other"
// spec references that response via an external $ref. With strict-server
// enabled in both packages, the embedded response field name must agree across
// packages so cross-package response casts compile.
package issue_2010

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.base.yaml spec-base.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.other.yaml spec-other.yaml
