//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --package=api --generate types,gin,spec,strict-server -o server.gen.go ../strict-schema.yaml

package api
