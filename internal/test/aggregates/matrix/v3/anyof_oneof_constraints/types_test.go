package matrixv3anyofoneofconstraints

// An anyOf and a oneOf whose branches only add constraints make no union:
// Subject is a plain struct with its four fields and nothing else. This fails
// to compile otherwise.
var _ = Subject{nil, nil, nil, nil}
