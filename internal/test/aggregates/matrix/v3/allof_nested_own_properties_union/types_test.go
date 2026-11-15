package matrixv3allofnestedownpropertiesunion

// b is a field next to the union. This fails to compile otherwise.
var _ = Subject{B: new(string)}
