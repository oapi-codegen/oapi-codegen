// Integration test program for the Fiber petstore variant.
// Starts the server on a random port, runs the test client, and shuts down.
package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/common/client/testclient"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/fiber/server"
)

func main() {
	app, err := server.NewFiberApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "setup failed: %v\n", err)
		os.Exit(1)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "listen failed: %v\n", err)
		os.Exit(1)
	}

	errCh := make(chan error, 1)
	go func() { errCh <- app.Listener(ln) }()

	serverURL := fmt.Sprintf("http://%s", ln.Addr().String())
	testErr := testclient.Run(serverURL)

	_ = app.ShutdownWithContext(context.Background())

	if srvErr := <-errCh; srvErr != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", srvErr)
		os.Exit(1)
	}
	if testErr != nil {
		fmt.Fprintf(os.Stderr, "FAILED: %v\n", testErr)
		os.Exit(1)
	}
	fmt.Println("\nPASSED: all checks passed")
}
