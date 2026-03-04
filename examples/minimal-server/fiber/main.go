package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/minimal-server/fiber/api"
)

func main() {
	// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code
	server := api.NewServer()

	app := fiber.New()

	api.RegisterHandlers(app, server)

	// And we serve HTTP until the world ends.
	log.Fatal(app.Listen("0.0.0.0:8080"))
}
