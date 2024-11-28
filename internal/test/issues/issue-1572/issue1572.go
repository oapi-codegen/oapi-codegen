package issue1572

// This is an example of a spec that previously generated a different
// embedded spec each time code generation was run.
// See https://github.com/oapi-codegen/oapi-codegen/issues/1572

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config config.yaml openapi.yml
