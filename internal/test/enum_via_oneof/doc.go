// Package enum_via_oneof verifies the OpenAPI 3.1 enum-via-oneOf idiom:
// a scalar schema whose oneOf branches each carry `title` + `const`
// renders as a Go typed enum, while a near-miss schema (one branch
// missing `title`) falls through to the standard union generator.
package enum_via_oneof

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
