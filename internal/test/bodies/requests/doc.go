// Package bodiesrequests exercises request bodies combined with parameters in the generated
// client, folded from the schemas.yaml kitchen sink.
//
//   - issue #9: client params type must not be incorrectly included for a request that has
//     both a body and parameters.
//   - issue #2194: typed clients can omit optional request bodies.
package bodiesrequests

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
