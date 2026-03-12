// Integration test program for the Iris petstore variant.
// Starts the server on a random port, runs the test client, and shuts down.
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/common/client/testclient"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/iris/server"
)

func main() {
	app, err := server.NewIrisApp()
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
	go func() {
		errCh <- app.Run(
			iris.Listener(ln),
			iris.WithoutBanner,
			iris.WithoutServerError(iris.ErrServerClosed),
		)
	}()

	serverURL := fmt.Sprintf("http://%s", ln.Addr().String())
	testErr := testclient.Run(serverURL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = app.Shutdown(ctx)

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
