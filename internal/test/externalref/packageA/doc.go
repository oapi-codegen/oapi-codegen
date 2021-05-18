package packageA

//go:generate go run github.com/dmhdeveloper/oapi-codegen/cmd/oapi-codegen -generate types,skip-prune,spec --package=packageA -o externalref.gen.go --import-mapping=../packageB/spec.yaml:github.com/dmhdeveloper/oapi-codegen/internal/test/externalref/packageB spec.yaml
