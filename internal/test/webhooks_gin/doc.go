// Package webhooks_gin verifies that the gin-server flag emits a
// compilable WebhookReceiverInterface with gin's (c *gin.Context)
// signature. Compile-time assertion only.
package webhooks_gin

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
