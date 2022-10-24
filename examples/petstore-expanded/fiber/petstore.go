// This is an example of implementing the Pet Store from the OpenAPI documentation
// found at:
// https://github.com/OAI/OpenAPI-Specification/blob/master/examples/v3.0/petstore.yaml

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/deepmap/oapi-codegen/examples/petstore-expanded/fiber/api"
	middleware "github.com/deepmap/oapi-codegen/pkg/fiber-middleware"
	"github.com/gofiber/fiber/v2"
)

func main() {

	var port = flag.Int("port", 8080, "Port for test HTTP server")
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

	// This is how you set up a basic fiber router
	app := fiber.New()

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	app.Use(middleware.OapiRequestValidator(swagger))

	// We now register our petStore above as the handler for the interface
	api.HandlerFromMux(petStore, app)

	// And we serve HTTP until the world ends.
	log.Fatal(app.Listen(fmt.Sprintf("0.0.0.0:%d", *port)))
}
