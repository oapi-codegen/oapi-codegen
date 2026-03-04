package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/minimal-server/gorillamux/api"
)

func main() {
	// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code
	server := api.NewServer()

	r := mux.NewRouter()

	// get an `http.Handler` that we can use
	h := api.HandlerFromMux(server, r)

	s := &http.Server{
		Handler: h,
		Addr:    "0.0.0.0:8080",
	}

	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}
