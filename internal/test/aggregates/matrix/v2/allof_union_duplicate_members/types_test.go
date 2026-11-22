package matrixv2allofunionduplicatemembers

// This test records how schema-merging-behavior v2 handles this shape. A bug fix
// may change it, but it must not be changed to accept a regression: code that
// generated and worked before must keep doing so. A commit that changes it must
// say why.

// Cat and Dog are members of both the allOf's oneOf and the sibling anyOf;
// their accessors exist once. The package would not compile otherwise.
var (
	_ = Subject.AsCat
	_ = Subject.AsDog
	_ = (*Subject).FromCat
)
