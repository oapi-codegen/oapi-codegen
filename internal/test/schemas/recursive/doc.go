// Package schemasrecursive exercises recursive and cyclic schema patterns:
// self-referencing types via additionalProperties (issue #52), cyclic oneOf
// references (issue #936, requires circular-reference-limit), and recursive
// $ref inside allOf (issue #1373). All cases are models-only; the point is
// that code generation and compilation succeed without infinite loops.
package schemasrecursive

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
