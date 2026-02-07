//go:build go1.22

// This is an example of implementing the Pet Store from the OpenAPI documentation
// found at:
// https://github.com/OAI/OpenAPI-Specification/blob/master/examples/v3.0/petstore.yaml

package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/oapi-codegen/oapi-codegen-exp/experimental/examples/petstore-expanded/chi/server"
)

func main() {
	port := flag.String("port", "8080", "Port for test HTTP server")
	flag.Parse()

	// Create an instance of our handler which satisfies the generated interface
	petStore := server.NewPetStore()

	r := chi.NewRouter()

	// We now register our petStore above as the handler for the interface
	server.HandlerFromMux(petStore, r)

	s := &http.Server{
		Handler: r,
		Addr:    net.JoinHostPort("0.0.0.0", *port),
	}

	log.Printf("Server listening on %s", s.Addr)

	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}
