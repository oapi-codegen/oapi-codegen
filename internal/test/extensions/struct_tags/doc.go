// Package extensionsstructtags exercises struct-tag-related extensions:
// x-oapi-codegen-extra-tags (extra struct tags on schema properties and on query
// parameters, both at the parameter level and at the schema level within a parameter);
// x-omitempty (force-omit/non-omit on optional fields, including sibling override
// next to a $ref); and param-ref sibling x-omitempty (x-omitempty: false on a
// query param schema $ref keeps the pointer field but removes omitempty from the
// generated JSON tag).
package extensionsstructtags

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_extra_tags_params.yaml spec_extra_tags_params.yaml
