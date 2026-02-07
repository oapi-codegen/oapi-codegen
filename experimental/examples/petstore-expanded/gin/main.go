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

	"github.com/gin-gonic/gin"
	"github.com/oapi-codegen/oapi-codegen/experimental/examples/petstore-expanded/gin/server"
)

func main() {
	port := flag.String("port", "8080", "Port for test HTTP server")
	flag.Parse()

	// Create an instance of our handler which satisfies the generated interface
	petStore := server.NewPetStore()

	r := gin.Default()

	// We now register our petStore above as the handler for the interface
	server.RegisterHandlers(r, petStore)

	s := &http.Server{
		Handler: r,
		Addr:    net.JoinHostPort("0.0.0.0", *port),
	}

	log.Printf("Server listening on %s", s.Addr)

	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}
