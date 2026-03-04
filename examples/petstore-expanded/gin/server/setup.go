package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	middleware "github.com/oapi-codegen/gin-middleware"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/gin/api"
)

// NewServer creates a fully configured *http.Server with the petstore handler
// and OpenAPI validation middleware. The caller should set Addr before calling
// ListenAndServe, or provide a net.Listener and call Serve.
func NewServer() (*http.Server, error) {
	swagger, err := api.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("error loading swagger spec: %w", err)
	}

	swagger.Servers = nil

	petStore := NewPetStore()

	r := gin.Default()
	r.Use(middleware.OapiRequestValidator(swagger))
	api.RegisterHandlers(r, petStore)

	return &http.Server{Handler: r}, nil
}
