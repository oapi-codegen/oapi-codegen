// Package extensionsskipoptionalpointer exercises x-go-type-skip-optional-pointer:
// on scalars via $ref, with property-level true/false overrides, on container
// types (slices/maps/[]byte) both explicitly and via
// prefer-skip-optional-pointer-on-container-types, and in the client
// request-builder + custom MarshalJSON paths (nil omitted, never `null`).
package extensionsskipoptionalpointer

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
