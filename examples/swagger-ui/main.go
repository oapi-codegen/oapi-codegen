package main

import (
	"log"
	"net/http"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/swagger-ui/api"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func main() {
	server := api.NewServer()

	r := http.NewServeMux()

	r.Handle("/swagger-ui/", httpSwagger.Handler())

	spec, _ := api.RawSpec()
	r.HandleFunc("/swagger-ui/doc.json", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(spec)
	})

	// get an `http.Handler` that we can use
	h := api.HandlerFromMux(server, r)

	s := &http.Server{
		Handler: h,
		Addr:    "localhost:8080",
	}

	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}
