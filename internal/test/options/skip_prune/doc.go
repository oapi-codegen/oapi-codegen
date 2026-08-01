// Package optionsskipprune exercises the output-options skip-prune flag:
// by default the code generator prunes schemas that are not referenced by
// any path operation; with skip-prune:true every schema in components/schemas
// is emitted regardless of whether it is referenced. Also covers the
// client-response-bytes-function output option in the same generation run.
package optionsskipprune

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_default_prune.yaml spec_default_prune.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
