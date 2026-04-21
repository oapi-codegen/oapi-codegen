package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/oapi-codegen/oapi-codegen/v2/examples/streaming/client/sse"
)

func main() {
	serverURL := flag.String("url", "http://localhost:8080", "server base URL")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	client, err := sse.NewClient(*serverURL)
	if err != nil {
		slog.Error("NewClient failed", "error", err)
		os.Exit(1)
	}

	// Use the plain Client (not ClientWithResponses) so the response body stays
	// an open io.Reader — ClientWithResponses would io.ReadAll the stream.
	resp, err := client.GetStream(ctx)
	if err != nil {
		slog.Error("GetStream failed", "error", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("unexpected status", "status", resp.Status)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		slog.Error("scan failed", "error", err)
		os.Exit(1)
	}
}
