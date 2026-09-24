// Package matrix is the aggregate-schema (allOf / anyOf / oneOf) test matrix.
//
// Each file in shapes/ describes one schema shape: a subject schema, the
// components it needs, and JSON samples that are valid instances of it.
// specgen turns every shape into its own package, whose spec places the
// subject at each position we generate code for (component schema, property,
// array items, additionalProperties value, request and response bodies, inline
// or referenced, component request bodies and responses, a `default:` response
// and a response with headers). The generated harness sends every sample
// through a generated client and strict server and checks that it comes back
// unchanged, both on the wire and after parsing.
//
// Shape-specific checks (Go types, From/As behavior, discriminators) live in
// hand-written <shape>_test.go files next to the harness.
//
// The generated code doubles as a record of generator output: changes to it
// must be deliberate.
package matrix

//go:generate go run ./specgen
