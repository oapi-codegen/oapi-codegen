package matrixv3allofnestedownproperties

// Subject keeps the property Mid declares next to its allOf. This fails to
// compile otherwise.
var _ = Subject{Name: "rex", B: new(string), C: new(string)}
