// Package xenumnames provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.0.0-00010101000000-000000000000 DO NOT EDIT.
package xenumnames

// Defines values for ClientType.
const (
	ACT ClientType = "ACT"
	EXP ClientType = "EXP"
)

// Defines values for ClientTypeWithNamesExtension.
const (
	ClientTypeWithNamesExtensionActive  ClientTypeWithNamesExtension = "ACT"
	ClientTypeWithNamesExtensionExpired ClientTypeWithNamesExtension = "EXP"
)

// Defines values for ClientTypeWithVarNamesExtension.
const (
	ClientTypeWithVarNamesExtensionActive  ClientTypeWithVarNamesExtension = "ACT"
	ClientTypeWithVarNamesExtensionExpired ClientTypeWithVarNamesExtension = "EXP"
)

// ClientType defines model for ClientType.
type ClientType string

// ClientTypeWithNamesExtension defines model for ClientTypeWithNamesExtension.
type ClientTypeWithNamesExtension string

// ClientTypeWithVarNamesExtension defines model for ClientTypeWithVarNamesExtension.
type ClientTypeWithVarNamesExtension string
