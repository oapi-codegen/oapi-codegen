// Package b provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.0.0-00010101000000-000000000000 DO NOT EDIT.
package b

// Element defines model for Element.
type Element struct {
	C *string `json:"c,omitempty"`
}

// Merge defines model for Merge.
type Merge struct {
	B *[]Element `json:"b,omitempty"`
}