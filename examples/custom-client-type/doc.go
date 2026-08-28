package customclienttype

// This is an example of how to add a prefix to the name of the generated Client struct
// See https://github.com/oapi-codegen/oapi-codegen/issues/785 for why this might be necessary
//
// NOTE that `client-type-name` is deprecated in favour of
// `output-options.component-names.client`, which renames the whole client
// family (ClientInterface, NewClient, ClientWithResponses, ...) rather than
// only the Client struct. This example is kept as-is to pin the behaviour of
// the deprecated option.

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config cfg.yaml api.yaml
