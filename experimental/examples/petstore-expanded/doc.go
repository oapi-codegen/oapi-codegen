// Package petstore provides generated types and client for the Petstore API.
//
//go:generate go run github.com/oapi-codegen/oapi-codegen/experimental/cmd/oapi-codegen -config models.config.yaml petstore-expanded.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/experimental/cmd/oapi-codegen -config client/client.config.yaml -output client/client.gen.go petstore-expanded.yaml
package petstore
