//go:build go1.22

// This is an example of implementing the Pet Store from the OpenAPI documentation
// found at:
// https://github.com/OAI/OpenAPI-Specification/blob/master/examples/v3.0/petstore.yaml

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/stdhttp/api"
)

func main() {
	port := flag.String("port", "8080", "Port for test HTTP server")
	flag.Parse()

	swagger, err := api.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	// Create an instance of our handler which satisfies the generated interface
	petStore := api.NewPetStore()

	r := http.NewServeMux()

	// We now register our petStore above as the handler for the interface
	api.HandlerFromMux(petStore, r)

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	h := middleware.OapiRequestValidator(swagger)(r)

	s := &http.Server{
		Handler: h,
		Addr:    net.JoinHostPort("0.0.0.0", *port),
	}

	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}
