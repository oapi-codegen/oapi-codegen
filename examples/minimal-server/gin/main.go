package main

import (
	"log"
	"net/http"

	"github.com/deepmap/oapi-codegen/v2/examples/minimal-server/gin/api"
	"github.com/gin-gonic/gin"
)

func main() {
	// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code
	server := api.NewServer()

	r := gin.Default()

	api.RegisterHandlers(r, server)

	// And we serve HTTP until the world ends.

	s := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:8080",
	}

	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}
