package matrixv3allofconstraintoneofrefs

// A oneOf of $refs to constraint-only schemas next to allOf is dropped, so the
// schema stays an alias of its allOf member. This fails to compile if Subject
// becomes a type of its own.
var _ Contact = Subject{}
