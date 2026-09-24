package matrixv2allofoneofsiblings

// This test records how schema-merging-behavior v2 handles this shape. A bug fix
// may change it, but it must not be changed to accept a regression: code that
// generated and worked before must keep doing so. A commit that changes it must
// say why.

// A oneOf next to allOf is merged into the type, so Subject has the
// union's accessors. This fails to compile otherwise.
var (
	_ = Subject.AsCat
	_ = (*Subject).FromDog
)
