// Package parameterspathedgecases covers special cases in path handling:
// colons in paths and URL escaping of reserved characters (issue #312), and
// operation-level parameter definitions overriding path-level ones (issue #1180).
package parameterspathedgecases

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
