// Package aggregateshoisting exercises anonymous inner schema hoisting into named
// types. It covers two distinct behaviors:
//
//   - explicit hoisting: generate-types-for-anonymous-schemas enabled (OpenAPI 3.0).
//     Inline allOf + sibling properties are hoisted into named response-body types;
//     top-level array items become a named element type.
//
//   - implicit/default hoisting: no flag needed (OpenAPI 3.1). Inline oneOf/anyOf
//     schemas at every operation position (response root, array items, nested
//     properties, request body root, request body property, webhooks, callbacks)
//     are hoisted and receive As<Branch>()/From<Branch>()/Merge<Branch>() accessors.
package aggregateshoisting

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_explicit.yaml spec_explicit.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_implicit.yaml spec_implicit.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_schemas.yaml spec_schemas.yaml
