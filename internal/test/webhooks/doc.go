// Package webhooks verifies OpenAPI 3.1 webhook codegen end-to-end:
// the generated WebhookInitiator fires a webhook against a httptest
// server, and the generated WebhookReceiverInterface handler receives
// it. The test asserts the payload round-trips intact.
package webhooks

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
