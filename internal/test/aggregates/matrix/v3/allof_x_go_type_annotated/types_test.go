package matrixv3allofxgotypeannotated

// A member that x-go-type replaces, and that the other members only annotate,
// is the composition's type: Subject must be the x-go-type, map[string]any.
// This fails to compile otherwise.
var _ Subject = map[string]any{"name": "rex"}
