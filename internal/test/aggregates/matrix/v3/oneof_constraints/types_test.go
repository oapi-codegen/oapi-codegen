package matrixv3oneofconstraints

// A oneOf whose branches only add constraints makes no union: Subject is a
// plain struct with its two fields and nothing else. This fails to compile
// otherwise.
var _ = Subject{nil, nil}
