package matrixv3allofnestedownpropertiesempty

// b is a field. This fails to compile otherwise.
var _ = Subject{B: new(string)}
