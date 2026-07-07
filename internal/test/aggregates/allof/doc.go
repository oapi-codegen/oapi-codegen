// Package aggregatesallof exercises allOf composition and additionalProperties merge
// semantics. It covers:
//   - allOf with additionalProperties on all constituent schemas (issue #193)
//   - additionalProperties merge-precedence rules across all combinations of
//     true/false/schema/default in allOf members (issue #1219)
//   - legacy allOf merge behavior (struct embedding) via compatibility.old-merge-schemas
//     (all_of/v1)
//   - modern allOf merge behavior (flat struct) without the compat flag (all_of/v2)
package aggregatesallof

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_old_merge.yaml spec_old_merge.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config_new_merge.yaml spec_new_merge.yaml
