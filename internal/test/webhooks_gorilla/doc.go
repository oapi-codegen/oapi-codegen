// Package webhooks_gorilla verifies that the gorilla/mux server flag
// emits a compilable WebhookReceiverInterface alongside gorilla's
// path-server interface. Same (w, r) signature as stdhttp/chi, so the
// receiver shape is identical; this is a compile-time assertion.
package webhooks_gorilla

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
