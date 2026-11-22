package extensionsxgoname

import "testing"

// Compile-time guards that x-go-name produced the renamed identifiers: RenameMe was
// renamed to NewName, ReferenceToRenameMe.ToNewName resolves to that new name, and the
// response / request body were renamed. If x-go-name regressed, these names would not
// exist and the package would fail to compile.
func TestXGoNameRenames(t *testing.T) {
	var (
		_ NewName
		_ ReferenceToRenameMe
		_ RenamedResponseObject
		_ RenamedRequestBody
	)
}
