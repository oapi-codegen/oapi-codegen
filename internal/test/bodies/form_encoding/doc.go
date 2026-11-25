// Package bodiesformencoding exercises the encoding of
// application/x-www-form-urlencoded request body properties.
//
//   - issue #2604: an array of primitives which declares style: form,
//     explode: true is sent as repeated keys (topic=a&topic=b) rather than
//     indexed ones (topic[0]=a&topic[1]=b).
//
// Every other declared encoding is ignored by the client, as it always has
// been, because generated servers don't read those serializations back. The
// tests pin both halves: the wire format, and the round trip through a
// generated strict server.
package bodiesformencoding

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
