// Regression fixture for:
//   - https://github.com/oapi-codegen/oapi-codegen/issues/1328
//   - https://github.com/oapi-codegen/oapi-codegen/issues/2010
//
// The base spec defines reusable responses with concrete, untyped, pointer,
// and header-bearing JSON bodies. The "other" specs reference those responses
// via external $refs. With strict-server enabled in both packages, the imported
// response envelope must remain embedded so cross-package response casts compile.
package responsecast

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.base.yaml spec-base.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.other.yaml spec-other.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.other-fiber.yaml spec-other.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.other-iris.yaml spec-other.yaml
