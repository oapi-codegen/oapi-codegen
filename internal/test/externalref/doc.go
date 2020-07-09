package externalref

//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen -generate types,skip-prune --package=externalref -o externalref.gen.go --import-mapping={"./packageA/spec.yaml":"github.com/deepmap/oapi-codegen/internal/test/externalref/packageA","./packageB/spec.yaml":"github.com/deepmap/oapi-codegen/internal/test/externalref/packageB"} spec.yaml
