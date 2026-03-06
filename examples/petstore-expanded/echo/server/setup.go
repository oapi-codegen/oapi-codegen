package server

import (
	"fmt"

	"github.com/labstack/echo/v4"
	middleware "github.com/oapi-codegen/echo-middleware"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/echo/api"
)

// NewEchoServer creates a fully configured *echo.Echo with the petstore handler
// and OpenAPI validation middleware.
func NewEchoServer() (*echo.Echo, error) {
	swagger, err := api.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("error loading swagger spec: %w", err)
	}

	swagger.Servers = nil

	petStore := NewPetStore()

	e := echo.New()
	e.Use(middleware.OapiRequestValidator(swagger))
	api.RegisterHandlers(e, petStore)

	return e, nil
}
