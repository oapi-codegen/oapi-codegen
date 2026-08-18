// Package openapi31params verifies that OpenAPI 3.1 multi-type union
// parameters (`type: [string, integer]`) bind end-to-end. The models-only
// lowering of unions to `any` is covered in the sibling openapi31 suite;
// this suite generates a std-http server and client so the emitted
// bind-option Types lists are exercised against the real runtime binder
// (runtime >= v1.7.0), in path, query, and header positions:
//
//   - the value binds to the first union member that parses, trying
//     boolean, integer, number, then string — so "42" arrives as int64(42)
//     and "abc" as "abc" in the same `any` parameter.
//   - a "null" entry in the type list is the nullability marker, not a
//     member, and is stripped from the emitted Types list.
//   - single-type parameters emit no Types field at all, keeping their
//     generated code identical to before.
package openapi31params

//go:generate go run github.com/wayleadr/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
