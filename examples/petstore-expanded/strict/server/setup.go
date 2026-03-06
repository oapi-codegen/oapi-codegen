package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/strict/api"
)

// NewServer creates a fully configured *http.Server with the strict petstore
// handler and OpenAPI validation middleware. The caller should set Addr before
// calling ListenAndServe, or provide a net.Listener and call Serve.
func NewServer() (*http.Server, error) {
	swagger, err := api.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("error loading swagger spec: %w", err)
	}

	swagger.Servers = nil

	petStore := NewPetStore()
	petStoreStrictHandler := api.NewStrictHandler(petStore, nil)

	r := chi.NewRouter()
	r.Use(middleware.OapiRequestValidator(swagger))
	api.HandlerFromMux(petStoreStrictHandler, r)

	return &http.Server{Handler: r}, nil
}
