// Package webhooks_iris verifies that the iris-server flag emits a
// compilable WebhookReceiverInterface with iris's (ctx iris.Context)
// signature. Compile-time assertion only.
package webhooks_iris

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
