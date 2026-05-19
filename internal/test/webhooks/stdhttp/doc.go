// Package stdhttp verifies OpenAPI 3.1 webhook codegen end-to-end for
// the stdhttp server flavor: the generated WebhookInitiator fires a
// webhook against a httptest server, and the generated
// WebhookReceiverInterface handler receives it. The test asserts the
// payload round-trips intact.
package stdhttp

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml ../spec.yaml
