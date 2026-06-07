// Package webhooks_echo verifies that the echo-server flag emits a
// compilable WebhookReceiverInterface with echo's (ctx echo.Context)
// error signature. Compile-time assertion only; runtime round-trip is
// covered by internal/test/webhooks (stdhttp).
package echo

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml ../spec.yaml
