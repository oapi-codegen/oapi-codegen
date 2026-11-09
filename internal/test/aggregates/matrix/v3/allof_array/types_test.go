package matrixv3allofarray

// allOf over an array schema keeps its items: Subject must be []Item,
// not []any. This fails to compile otherwise.
var _ Subject = []Item{}
