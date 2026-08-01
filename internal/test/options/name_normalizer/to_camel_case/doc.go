// Package optionsnamenormalizertocamelcase checks that name-normalizer:
// ToCamelCase produces the same output as unset for this spec: uuid stays
// Uuid, operationId getHttpPet → GetHttpPet (on both client and server), and
// a digit is not a word boundary so OneOf2things stays OneOf2things.
//
// outputoptions/name-normalizer/to-camel-case
package optionsnamenormalizertocamelcase

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml ../spec.yaml
