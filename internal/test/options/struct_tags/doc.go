// Package optionsstructtags exercises the struct-tags output option: struct
// tag generation is user-configurable via Go text/templates. The config for
// this package overrides the yaml tag template (superseding the yaml-tags
// flag), and adds validate and db tags, while keeping the default json and
// form templates.
package optionsstructtags

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
