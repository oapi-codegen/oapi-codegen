package packageA

//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen -generate types,skip-prune,spec --package=packageA -o externalref.gen.go --import-mapping=../packageB/spec.yaml:github.com/deepmap/oapi-codegen/v2/internal/test/externalref/packageB spec.yaml
