//go:build go1.25

// Integration test program for the Echo v5 petstore variant.
// Starts the server on a random port, runs the test client, and shuts down.
package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/common/client/testclient"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/echo-v5/server"
)

func main() {
	e, err := server.NewEchoServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "setup failed: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addrCh := make(chan net.Addr, 1)
	sc := echo.StartConfig{
		Address:          "127.0.0.1:0",
		HideBanner:       true,
		HidePort:         true,
		ListenerAddrFunc: func(addr net.Addr) { addrCh <- addr },
	}

	errCh := make(chan error, 1)
	go func() { errCh <- sc.Start(ctx, e) }()

	addr := <-addrCh
	serverURL := fmt.Sprintf("http://%s", addr.String())
	testErr := testclient.Run(serverURL)

	cancel() // triggers graceful shutdown

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
