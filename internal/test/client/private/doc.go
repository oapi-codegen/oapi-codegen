package pkg

//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --package=pkg --make-private=client -o client.gen.go ../client.yaml
