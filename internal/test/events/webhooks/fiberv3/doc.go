// Package webhooks_fiberv3 verifies that the fiber-v3-server flag
// emits a compilable WebhookReceiverInterface with fiber v3's
// (c fiber.Ctx) error signature. Compile-time assertion only.
package fiberv3

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml ../spec.yaml
