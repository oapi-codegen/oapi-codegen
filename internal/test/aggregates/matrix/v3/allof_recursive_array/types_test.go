package matrixv3allofrecursivearray

// A self-referential array is a defined type; as an alias it would be an
// invalid recursive type, and this would not compile.
var _ = Subject{Subject{}, Subject{Subject{}}}
