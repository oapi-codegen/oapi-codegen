package server

import (
	"fmt"
	"net/http"

	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/stdhttp/api"
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

	r := http.NewServeMux()
	api.HandlerFromMux(petStore, r)

	h := middleware.OapiRequestValidator(swagger)(r)

	return &http.Server{Handler: h}, nil
}
