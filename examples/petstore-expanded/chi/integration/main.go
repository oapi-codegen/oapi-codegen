// Integration test program for the Chi petstore variant.
// Starts the server on a random port, runs the test client, and shuts down.
package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/chi/server"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/common/client/testclient"
)

func main() {
	s, err := server.NewServer()
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
	go func() { errCh <- s.Serve(ln) }()

	serverURL := fmt.Sprintf("http://%s", ln.Addr().String())
	testErr := testclient.Run(serverURL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.Shutdown(ctx)

	if srvErr := <-errCh; srvErr != nil && !errors.Is(srvErr, http.ErrServerClosed) {
		fmt.Fprintf(os.Stderr, "server error: %v\n", srvErr)
		os.Exit(1)
	}
	if testErr != nil {
		fmt.Fprintf(os.Stderr, "FAILED: %v\n", testErr)
		os.Exit(1)
	}
	fmt.Println("\nPASSED: all checks passed")
}
