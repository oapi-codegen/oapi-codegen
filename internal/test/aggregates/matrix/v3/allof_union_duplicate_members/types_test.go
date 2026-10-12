package matrixv3allofunionduplicatemembers

// Cat and Dog are members of both the allOf's oneOf and the sibling anyOf;
// their accessors exist once. The package would not compile otherwise.
var (
	_ = Subject.AsCat
	_ = Subject.AsDog
	_ = (*Subject).FromCat
)
