package matrixallofoneofsiblings

// A oneOf next to allOf is merged into the type, so Subject has the
// union's accessors. This fails to compile otherwise.
var (
	_ = Subject.AsCat
	_ = (*Subject).FromDog
)
