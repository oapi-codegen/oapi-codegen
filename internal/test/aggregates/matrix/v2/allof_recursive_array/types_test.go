package matrixv2allofrecursivearray

// This test records how schema-merging-behavior v2 handles this shape. A bug fix
// may change it, but it must not be changed to accept a regression: code that
// generated and worked before must keep doing so. A commit that changes it must
// say why.

// A self-referential array is a defined type; as an alias it would be an
// invalid recursive type, and this would not compile.
var _ = Subject{Subject{}, Subject{Subject{}}}
