package issue1825

// We place the spec in a subdirectory, as this requires us to initialize the resolver kin-openapi's loader
// If this is not done, the generator would fail with an `encountered disallowed external reference` error.
//go:generate go run github.com/ascendsoftware/oapi-codegen/cmd/oapi-codegen --config=config.yaml spec/spec.yaml
