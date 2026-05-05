// Package callbacks verifies OpenAPI callback codegen end-to-end. The
// spec is OpenAPI 3.0 to ensure callback support is NOT gated on 3.1
// (callbacks have been part of the spec since 3.0). The CallbackInitiator
// fires a callback against an httptest server that registers the
// CallbackReceiverInterface handler.
package callbacks

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
