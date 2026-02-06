//go:generate go run github.com/oapi-codegen/oapi-codegen/experimental/cmd/oapi-codegen -config config.yaml tree-farm.yaml

// Package treefarm provides an example of how to handle OpenAPI callbacks.
// We create a server which plants trees. The client asks the server to plant
// a tree and requests a callback when the planting is done.
//
// The server program will wait 1-5 seconds before notifying the client that a
// tree has been planted.
//
// You can run the example by running these two commands in parallel
// go run ./server --port 8080
// go run ./client --server http://localhost:8080
package treefarm
