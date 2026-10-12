package matrixv3allofconstraintoneof

// A oneOf of constraint-only branches next to allOf is still dropped, so the
// schema stays an alias of its allOf member. This fails to compile if Subject
// becomes a type of its own.
var _ Contact = Subject{}
