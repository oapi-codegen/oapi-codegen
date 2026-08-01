// Package optionsnamenormalizertocamelcasewithadditionalinitialisms checks
// that name-normalizer: ToCamelCaseWithInitialisms combined with
// additional-initialisms: [NAME] renders the "name" field as NAME on top of
// the standard initialism expansions (uuid → UUID, getHttpPet → GetHTTPPet).
//
// outputoptions/name-normalizer/to-camel-case-with-additional-initialisms
package optionsnamenormalizertocamelcasewithadditionalinitialisms

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml ../spec.yaml
