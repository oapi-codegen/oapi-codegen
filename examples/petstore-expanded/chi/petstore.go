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

	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/chi/server"
)

func main() {
	port := flag.String("port", "8080", "Port for test HTTP server")
	flag.Parse()

	s, err := server.NewServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up server: %s\n", err)
		os.Exit(1)
	}
	s.Addr = net.JoinHostPort("0.0.0.0", *port)

	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}
