package customclienttype

// This is an example of how to add a prefix to the name of the generated Client struct
// See https://github.com/oapi-codegen/oapi-codegen/issues/785 for why this might be necessary

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config cfg.yaml api.yaml
