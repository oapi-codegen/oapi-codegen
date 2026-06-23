// Package optionsnamenormalizer exercises the output-options.name-normalizer
// configuration flag across all supported values, plus additional-initialisms.
// Each of the five triples uses a distinct spec (with prefixed schema names to
// avoid collisions in the shared package) and a different normalizer value,
// so that the casing differences in generated type/field/operation names are
// purely a function of the name-normalizer flag.
//
// Triples:
//
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_unset.yaml spec.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_to_camel_case.yaml spec_to_camel_case.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_to_camel_case_with_initialisms.yaml spec_to_camel_case_with_initialisms.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_to_camel_case_with_digits.yaml spec_to_camel_case_with_digits.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_to_camel_case_with_additional_initialisms.yaml spec_to_camel_case_with_additional_initialisms.yaml
package optionsnamenormalizer
