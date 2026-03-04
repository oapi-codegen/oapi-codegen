//go:build go1.25

package server

import (
	"fmt"

	"github.com/labstack/echo/v5"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/echo-v5/api"
	mw "github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/echo-v5/middleware"
)

// NewEchoServer creates a fully configured *echo.Echo (v5) with the petstore
// handler and OpenAPI validation middleware.
func NewEchoServer() (*echo.Echo, error) {
	swagger, err := api.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("error loading swagger spec: %w", err)
	}

	swagger.Servers = nil

	petStore := NewPetStore()

	e := echo.New()
	e.Use(mw.OapiRequestValidator(swagger))
	api.RegisterHandlers(e, petStore)

	return e, nil
}
