// Package issue_illegal_enum_names tests enum constant generation with edge cases.
// This tests various edge cases like empty strings, spaces, hyphens, leading digits, etc.
package issue_illegal_enum_names

//go:generate go run ../../../../../../cmd/oapi-codegen -package output -output output/types.gen.go spec.yaml
