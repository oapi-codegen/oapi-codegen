package matrixv2allofxgotypeannotated

// This test records how schema-merging-behavior v2 handles this shape. A bug fix
// may change it, but it must not be changed to accept a regression: code that
// generated and worked before must keep doing so. A commit that changes it must
// say why.

// v2 drops the member's x-go-type and copies the schema it replaces into a
// struct of its own. This fails to compile otherwise.
var _ = Subject{Name: "rex"}
