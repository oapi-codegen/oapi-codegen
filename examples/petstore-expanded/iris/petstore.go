// This is an example of implementing the Pet Store from the OpenAPI documentation
// found at:
// https://github.com/OAI/OpenAPI-Specification/blob/master/examples/v3.0/petstore.yaml
//
// The code under api/petstore/ has been generated from that specification.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/deepmap/oapi-codegen/examples/petstore-expanded/iris/api"
	middleware "github.com/oapi-codegen/iris-middleware"
	"github.com/kataras/iris/v12"
)

func NewIrisPetServer(petStore *api.PetStore, port int) *iris.Application {
	swagger, err := api.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	i := iris.Default()

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	i.Use(middleware.OapiRequestValidator(swagger))

	api.RegisterHandlers(i, petStore)

	return i
}

func main() {
	port := flag.Int("port", 8080, "Port for test HTTP server")
	flag.Parse()
	// Create an instance of our handler which satisfies the generated interface
	petStore := api.NewPetStore()
	s := NewIrisPetServer(petStore, *port)

	// And we serve HTTP until the world ends.
	log.Fatal(s.Listen(fmt.Sprintf("localhost:%d", *port)))
}
