package matrixv3allofmemberconstraints

// A member whose oneOf only adds constraints only annotates the $ref, so
// Subject is Contact. This fails to compile otherwise.
var _ Subject = Contact{}
