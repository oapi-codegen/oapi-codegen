//go:generate go run github.com/oapi-codegen/oapi-codegen/experimental/cmd/oapi-codegen -config config.yaml door-badge-reader.yaml

// Package doorbadge provides an example of OpenAPI 3.1 webhooks.
// A door badge reader server generates random enter/exit events and
// notifies registered webhook listeners.
//
// You can run the example by running these two commands in parallel:
//
//	go run ./server --port 8080
//	go run ./client --server http://localhost:8080
//
// You can run multiple clients and they will all get the notifications
package doorbadge
