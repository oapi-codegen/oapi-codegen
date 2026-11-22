// Package schemasrecursive exercises recursive and cyclic schema patterns:
// self-referencing types via additionalProperties (issue #52), cyclic oneOf
// references (issue #936, requires circular-reference-limit), recursive $ref
// inside allOf (issue #1373), and allOf compositions whose members refer back
// into a body still being generated (issue #2542, in its object, union,
// transitive, mutual and bystander forms). All cases are models-only; the
// point is that code generation and compilation succeed without infinite
// loops, and that a recursive composition keeps both halves of the allOf.
package schemasrecursive

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml spec.yaml
