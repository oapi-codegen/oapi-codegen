package server

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	middleware "github.com/oapi-codegen/fiber-middleware"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/fiber/api"
)

// NewFiberApp creates a fully configured *fiber.App with the petstore handler
// and OpenAPI validation middleware.
func NewFiberApp() (*fiber.App, error) {
	swagger, err := api.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("error loading swagger spec: %w", err)
	}

	swagger.Servers = nil

	petStore := NewPetStore()

	app := fiber.New()
	app.Use(middleware.OapiRequestValidator(swagger))
	api.RegisterHandlers(app, petStore)

	return app, nil
}
