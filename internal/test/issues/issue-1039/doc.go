package issue1039

//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=types-config.yaml spec.yaml
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=type-config-defaultbehaviour.yaml spec.yaml
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=client-config.yaml spec.yaml
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=server-config.yaml spec.yaml
