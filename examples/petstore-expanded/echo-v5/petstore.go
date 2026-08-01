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
	"net"
	"os"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/echo-v5/server"
)

func main() {
	port := flag.String("port", "8080", "Port for test HTTP server")
	flag.Parse()

	e, err := server.NewEchoServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up server: %s\n", err)
		os.Exit(1)
	}

	// And we serve HTTP until the world ends.
	log.Fatal(e.Start(net.JoinHostPort("0.0.0.0", *port)))
}
