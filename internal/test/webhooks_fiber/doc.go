// Package webhooks_fiber verifies that the fiber-server flag emits a
// compilable WebhookReceiverInterface with fiber's (c *fiber.Ctx)
// error signature. Compile-time assertion only.
package webhooks_fiber

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
