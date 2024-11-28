// This is an example of implementing the Pet Store from the OpenAPI documentation
// found at:
// https://github.com/OAI/OpenAPI-Specification/blob/master/examples/v3.0/petstore.yaml

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/gofiber/fiber/v2"

	middleware "github.com/oapi-codegen/fiber-middleware"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/fiber/api"
)

func main() {
	port := flag.String("port", "8080", "Port for test HTTP server")
	flag.Parse()

	// Create an instance of our handler which satisfies the generated interface
	petStore := api.NewPetStore()

	s := NewFiberPetServer(petStore)

	// And we serve HTTP until the world ends.
	log.Fatal(s.Listen(net.JoinHostPort("0.0.0.0", *port)))
}

func NewFiberPetServer(petStore *api.PetStore) *fiber.App {

	swagger, err := api.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	// This is how you set up a basic fiber router
	app := fiber.New()

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	app.Use(middleware.OapiRequestValidator(swagger))

	// We now register our petStore above as the handler for the interface
	api.RegisterHandlers(app, petStore)

	return app
}
