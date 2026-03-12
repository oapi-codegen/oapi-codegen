package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/chi/api"
)

// NewServer creates a fully configured *http.Server with the petstore handler
// and OpenAPI validation middleware. The caller should set Addr before calling
// ListenAndServe, or provide a net.Listener and call Serve.
func NewServer() (*http.Server, error) {
	swagger, err := api.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("error loading swagger spec: %w", err)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	petStore := NewPetStore()

	r := chi.NewRouter()
	r.Use(middleware.OapiRequestValidator(swagger))
	api.HandlerFromMux(petStore, r)

	return &http.Server{Handler: r}, nil
}
