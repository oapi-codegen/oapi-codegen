package matrixv2allofnestedownpropertiesmap

// This test records how schema-merging-behavior v2 handles this shape. A bug fix
// may change it, but it must not be changed to accept a regression: code that
// generated and worked before must keep doing so. A commit that changes it must
// say why.

// v2 keeps the merge a map rather than add b as a field (#2571). This fails
// to compile otherwise.
var _ Subject = map[string]string{"b": "x"}
