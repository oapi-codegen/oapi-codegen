// Package name_conflict_resolution tests comprehensive type name collision resolution.
// Exercises patterns from issues #200, #254, #255, #292, #407, #899, #1357, #1450,
// #1474, #1713, #1881, #2097, #2213.
package name_conflict_resolution

//go:generate go run ../../../../cmd/oapi-codegen -config config.yaml spec.yaml
