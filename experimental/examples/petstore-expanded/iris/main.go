//go:build go1.22

// This is an example of implementing the Pet Store from the OpenAPI documentation
// found at:
// https://github.com/OAI/OpenAPI-Specification/blob/master/examples/v3.0/petstore.yaml

package main

import (
	"flag"
	"log"
	"net"

	"github.com/kataras/iris/v12"
	"github.com/oapi-codegen/oapi-codegen-exp/experimental/examples/petstore-expanded/iris/server"
)

func main() {
	port := flag.String("port", "8080", "Port for test HTTP server")
	flag.Parse()

	// Create an instance of our handler which satisfies the generated interface
	petStore := server.NewPetStore()

	app := iris.New()

	// We now register our petStore above as the handler for the interface
	server.RegisterHandlers(app, petStore)

	addr := net.JoinHostPort("0.0.0.0", *port)
	log.Printf("Server listening on %s", addr)

	// And we serve HTTP until the world ends.
	log.Fatal(app.Listen(addr))
}
