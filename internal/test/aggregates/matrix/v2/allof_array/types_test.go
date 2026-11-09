package matrixv2allofarray

// This test records how schema-merging-behavior v2 handles this shape. A bug fix
// may change it, but it must not be changed to accept a regression: code that
// generated and worked before must keep doing so. A commit that changes it must
// say why.

// allOf over an array schema keeps its items: Subject must be []Item,
// not []any. This fails to compile otherwise.
var _ Subject = []Item{}
