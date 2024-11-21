// Package issues provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.0.0-00010101000000-000000000000 DO NOT EDIT.
package issues

import (
	externalRef0 "github.com/oapi-codegen/oapi-codegen/v2/internal/test/issues/issue-1841/b"
)

// Example defines model for Example.
type Example struct {
	A *string               `json:"a,omitempty"`
	C *externalRef0.Element `json:"c,omitempty"`
}

// Merged defines model for Merged.
type Merged struct {
	A *string                 `json:"a,omitempty"`
	B *[]externalRef0.Element `json:"b,omitempty"`
	C *externalRef0.Element   `json:"c,omitempty"`
}
