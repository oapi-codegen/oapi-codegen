// Package optionsyamltags exercises the yaml-tags output option: when enabled,
// generated struct types include yaml:"..." struct tags alongside the json:"..."
// tags, making them compatible with YAML marshalling/unmarshalling without
// extra configuration.
package optionsyamltags

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
