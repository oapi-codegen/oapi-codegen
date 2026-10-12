// Package matrix is the aggregate-schema (allOf / anyOf / oneOf) test matrix.
//
// Each file in shapes/ describes one schema shape: a subject schema, the
// components it needs, JSON samples that are valid instances of it, and the
// schema-merging-behavior versions it runs under. specgen turns every shape
// into one package per version, <version>/<shape>/, whose spec places the
// subject at each position we generate code for (component schema, property,
// array items, additionalProperties value, request and response bodies, inline
// or referenced, component request bodies and responses, a `default:` response
// and a response with headers), and whose config pins that version. The
// generated harness sends every sample through a generated client and strict
// server and checks that it comes back unchanged, both on the wire and after
// parsing.
//
// A version can reject a shape instead (the shape's rejects key). Its package
// then holds no generated code, and its test checks that generation fails with
// the expected error.
//
// Shape-specific checks (Go types, From/As behavior, discriminators) live in
// hand-written test files next to the harness.
//
// The tests in v1/ and v2/ guard those versions against regressions. Bug
// fixes may change what they generate, but the tests must not be changed to
// accept a regression, and a commit that changes them must say why. Editing a
// shape that v1 or v2 runs changes what their tests check, so add new cases
// as new shapes instead.
package matrix

//go:generate go run ./specgen
