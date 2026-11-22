// Package optionsstructtagsunions exercises the struct-tags output option
// with schema-merging-behavior v3 unions. The json template gives every
// field a key other than its property name, which the As* of a variant with
// additionalProperties: false must allow, and From* of a union combined with
// another must replace.
package optionsstructtagsunions

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
