package matrixallofuniqueitems

// allOf over an array schema keeps its items: Subject must be []string,
// not []any. This fails to compile otherwise.
var _ Subject = []string{}
