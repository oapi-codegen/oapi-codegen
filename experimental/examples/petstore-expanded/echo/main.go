//go:build go1.22

// This is an example of implementing the Pet Store from the OpenAPI documentation
// found at:
// https://github.com/OAI/OpenAPI-Specification/blob/master/examples/v3.0/petstore.yaml

package main

import (
	"flag"
	"log"
	"net"

	"github.com/labstack/echo/v5"
	"github.com/oapi-codegen/oapi-codegen-exp/experimental/examples/petstore-expanded/echo/server"
)

func main() {
	port := flag.String("port", "8080", "Port for test HTTP server")
	flag.Parse()

	// Create an instance of our handler which satisfies the generated interface
	petStore := server.NewPetStore()

	e := echo.New()

	// We now register our petStore above as the handler for the interface
	server.RegisterHandlers(e, petStore)

	log.Printf("Server listening on %s", net.JoinHostPort("0.0.0.0", *port))

	// And we serve HTTP until the world ends.
	log.Fatal(e.Start(net.JoinHostPort("0.0.0.0", *port)))
}
