// Package serversstrictmulticontentiris is a compile-only check that strict-server +
// iris-server codegen handles a response with multiple content types that share a base
// media type but differ by media-type parameters (application/json vs
// application/json; profile="Foo"/"Bar").
//
// issue #1529 (strict-iris): client + models + embedded-spec + iris-server + strict-server.
package serversstrictmulticontentiris

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
