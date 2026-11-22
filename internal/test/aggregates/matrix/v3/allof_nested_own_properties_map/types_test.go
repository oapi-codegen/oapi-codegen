package matrixv3allofnestedownpropertiesmap

// b is a field next to the additional properties. This fails to compile
// otherwise.
var _ = Subject{B: new(string), AdditionalProperties: map[string]string{"k": "v"}}
