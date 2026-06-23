// Package parametersheadernilcheck verifies that an optional array header parameter is
// omitted from the generated client request when nil, and sent when non-nil.
//
// issue #2238 (header half; the cookie half is parameters/cookie/nilcheck).
package parametersheadernilcheck

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
