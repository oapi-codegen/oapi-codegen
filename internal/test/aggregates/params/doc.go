// Package aggregatesparams covers oneOf/anyOf parameters (issue #2560). An
// inline union parameter used to be an anonymous struct with an unexported
// field, which no caller outside the generated package could build, and no
// union parameter could be bound by a server: a path parameter failed every
// request and a query parameter was silently dropped. Scalar unions now get a
// named type and UnmarshalText.
package aggregatesparams

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
