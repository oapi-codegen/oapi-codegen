// Package webhooks_chi verifies that the chi-server flag emits a
// compilable WebhookReceiverInterface alongside chi's path-server
// interface. Chi shares stdhttp's (w, r) handler signature, so the
// receiver shape is structurally identical to internal/test/webhooks
// (which already round-trip-tests the runtime behavior). This package
// is a compile-time assertion that the chi/chi-receiver.tmpl renders
// valid Go.
package webhooks_chi

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
