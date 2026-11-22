// Package optionsnamenormalizertocamelcasewithdigits checks that
// name-normalizer: ToCamelCaseWithDigits treats digit sequences as word
// boundaries so OneOf2things → OneOf2Things, but does NOT expand initialisms
// (uuid stays Uuid, operationId getHttpPet → GetHttpPet on both client and
// server).
//
// outputoptions/name-normalizer/to-camel-case-with-digits
package optionsnamenormalizertocamelcasewithdigits

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml ../spec.yaml
