// Package responsescontent verifies that a components/responses entry
// exposing the same schema under multiple content types (application/json
// and application/xml) generates client response fields that all point at
// the single declared component type rather than undefined per-content-type
// names.
//
// From issue-2389.
package responsescontent

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
