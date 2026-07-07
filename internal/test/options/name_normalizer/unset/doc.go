// Package optionsnamenormalizerunset checks that, with name-normalizer unset,
// the default Go casing is applied: "uuid" → Uuid, operationId getHttpPet →
// GetHttpPet (on both client and server), and a digit is not a word boundary
// so OneOf2things stays OneOf2things.
//
// outputoptions/name-normalizer/unset
package optionsnamenormalizerunset

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml ../spec.yaml
