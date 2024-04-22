package issue1362

// This is an example of a spec that uses $ref to reference the same schema
// multiple times, that would otherwise gennerate anonymous structs.
// See https://github.com/oapi-codegen/oapi-codegen/issues/1362

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config config.yaml api.yaml
