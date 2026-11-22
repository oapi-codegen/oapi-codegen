// Package multijson verifies that an operation declaring multiple
// application/*+json media types on its responses generates one client
// response field per media type, and that response parsing populates
// only the field matching the returned Content-Type and status code.
//
// From issue-1208-1209.
package multijson

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
