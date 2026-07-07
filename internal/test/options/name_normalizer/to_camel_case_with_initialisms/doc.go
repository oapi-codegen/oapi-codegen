// Package optionsnamenormalizertocamelcasewithinitialisms checks that
// name-normalizer: ToCamelCaseWithInitialisms expands common Go initialisms:
// uuid → UUID and operationId getHttpPet → GetHTTPPet (on both client and
// server); a digit becomes a word boundary so OneOf2things → OneOf2Things.
//
// outputoptions/name-normalizer/to-camel-case-with-initialisms
package optionsnamenormalizertocamelcasewithinitialisms

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml ../spec.yaml
