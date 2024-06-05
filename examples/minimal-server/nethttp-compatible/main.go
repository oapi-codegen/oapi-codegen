package main

import (
	"log"
	"net/http"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/minimal-server/gorillamux/api"
)

func main() {
	// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code
	server := api.NewServer()

	// get an `http.Handler` that we can use, but notice that we don't need to specify the router (although under-the-hood it'll use the generated router)
	h := api.Handler(server)

	s := &http.Server{
		Handler: h,
		Addr:    "0.0.0.0:8080",
	}

	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}
