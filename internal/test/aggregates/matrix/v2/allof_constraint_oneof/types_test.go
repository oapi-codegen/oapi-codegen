package matrixv2allofconstraintoneof

// This test records how schema-merging-behavior v2 handles this shape. A bug fix
// may change it, but it must not be changed to accept a regression: code that
// generated and worked before must keep doing so. A commit that changes it must
// say why.

// A oneOf of constraint-only branches next to allOf is still dropped, so the
// schema stays an alias of its allOf member. This fails to compile if Subject
// becomes a type of its own.
var _ Contact = Subject{}
