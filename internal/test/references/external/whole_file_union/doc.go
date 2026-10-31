// Package wholefileunion exercises oneOf/anyOf branches that $ref a whole file
// (issue #2639): a response body, an anyOf mixing a whole-file branch with an
// inline one, an object property, and a component schema. Such a branch names
// no Go type, so it is generated like an inline branch, named after its path.
package wholefileunion

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
