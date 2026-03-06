package server

import (
	"fmt"

	"github.com/kataras/iris/v12"
	middleware "github.com/oapi-codegen/iris-middleware"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/iris/api"
)

// NewIrisApp creates a fully configured *iris.Application with the petstore
// handler and OpenAPI validation middleware.
func NewIrisApp() (*iris.Application, error) {
	swagger, err := api.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("error loading swagger spec: %w", err)
	}

	swagger.Servers = nil

	petStore := NewPetStore()

	i := iris.Default()
	i.Use(middleware.OapiRequestValidator(swagger))
	api.RegisterHandlers(i, petStore)

	return i, nil
}
