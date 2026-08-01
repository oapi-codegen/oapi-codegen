// Package optionscontenttypes exercises the content-types output option: a
// short name mapped to media-type regex patterns is used as the tag in
// generated type names (e.g. V1 instead of ApplicationVndMycompanyV1JSON),
// and matched media types get models generated even when they are not one of
// the built-in supported types (e.g. text/csv).
package optionscontenttypes

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
