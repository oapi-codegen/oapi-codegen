// Package namingoptions exercises the opt-in output-options.naming settings together:
// enum-values.keep-underscores (enum value "LOG_LEVEL_DEBUG" → LOG_LEVEL_DEBUG rather than
// LOGLEVELDEBUG), properties.leading-underscore: drop ("_id" → Id rather than UnderscoreId),
// and explicit-names: verbatim (x-go-name and x-enum-varnames emitted exactly as written,
// without the type name prefix an enum's other constants get).
package namingoptions

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
