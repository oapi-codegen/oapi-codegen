// Package aggregatesanyof exercises anyOf/oneOf/allOf union types in multiple forms:
// inline anyOf in a response body (no named schema), anyOf via $ref to named schemas,
// anyOf+allOf+oneOf combined in object properties (issue #1189), and anyOf/oneOf used
// as query-parameter types with serialization (any_of/param).
package aggregatesanyof

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
