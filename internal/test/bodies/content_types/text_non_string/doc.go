// Package textnonstring verifies strict-server codegen for text/plain
// responses whose schema is a non-string primitive (integer, boolean).
// The generated Visit* method must compile and write the value's decimal /
// literal text form. Before the fix, the generated code did
// []byte(response) on a non-string-underlying type, which either failed to
// compile (integer, boolean) or, on the fiber/iris path, string(response)
// silently produced a single rune instead of the number.
//
// From issue-1897.
package textnonstring

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
