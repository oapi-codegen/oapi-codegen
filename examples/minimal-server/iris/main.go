package main

import (
	"log"

	"github.com/kataras/iris/v12"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/minimal-server/iris/api"
)

func main() {
	// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code
	server := api.NewServer()

	i := iris.Default()

	api.RegisterHandlers(i, server)

	// And we serve HTTP until the world ends.
	log.Fatal(i.Listen("0.0.0.0:8080"))
}
