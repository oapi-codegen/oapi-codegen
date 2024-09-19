package main

import (
	"context"
	"github.com/oapi-codegen/oapi-codegen/v2/examples/streaming/stdhttp/sse"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code
	server := sse.NewServer()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	err := server.Run(ctx)
	if err != nil {
		slog.Error("server run failed", "error:", err)
		os.Exit(1)
	}
}
