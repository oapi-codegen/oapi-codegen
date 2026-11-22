package matrixv2allofpropertyrefine

// This test records how schema-merging-behavior v2 handles this shape. A bug fix
// may change it, but it must not be changed to accept a regression: code that
// generated and worked before must keep doing so. A commit that changes it must
// say why.

// In v2 the last member to declare a property decides its schema, so name is
// just `{nullable: true}`: any. This fails to compile otherwise.
var _ = Subject{Name: any(nil)}
