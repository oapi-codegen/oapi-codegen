// CLI test client for the petstore-expanded example.
// Start any server variant, then run:
//
//	go run ./common/client/ --port 8080
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/common/client/testclient"
)

func main() {
	port := flag.String("port", "8080", "Port of the running server")
	flag.Parse()

	serverURL := fmt.Sprintf("http://localhost:%s", *port)
	if err := testclient.Run(serverURL); err != nil {
		fmt.Fprintf(os.Stderr, "FAILED: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("\nPASSED: all checks passed")
}
