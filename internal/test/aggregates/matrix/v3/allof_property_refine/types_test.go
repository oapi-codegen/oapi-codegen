package matrixv3allofpropertyrefine

// The member refines Base's name rather than replacing it: name stays a
// string, and becomes nullable. This fails to compile otherwise.
var _ = Subject{Name: new(string)}
