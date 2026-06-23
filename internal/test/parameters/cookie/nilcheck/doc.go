// Package parameterscookienilcheck verifies that an optional array cookie parameter is
// omitted from the generated client request when nil, and sent when non-nil.
//
// issue #2238 (cookie half; the header half is parameters/header/nilcheck).
package parameterscookienilcheck

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
