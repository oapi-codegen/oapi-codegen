// Package openapi31enumviaoneof — 3.1 idiom: scalar schema + oneOf branches with title+const -> Go typed enum; falls through to standard union when a branch lacks title.
package openapi31enumviaoneof

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
