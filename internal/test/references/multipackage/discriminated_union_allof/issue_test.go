package discriminatedunionallof

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oapi-codegen/oapi-codegen/v2/internal/test/references/multipackage/discriminated_union_allof/gen/api"
	"github.com/oapi-codegen/oapi-codegen/v2/internal/test/references/multipackage/discriminated_union_allof/gen/common"
)

// TestValueByDiscriminatorExternalRef verifies that PRFile.ValueByDiscriminator()
// dispatches to the correct As* helper when the discriminated union is pulled in
// from another package via an external $ref inside an allOf.
//
// Before the fix, ValueByDiscriminator() interpolated the raw discriminator
// mapping value (externalRef0.GitDiffFile) into the method name, generating
// t.AsexternalRef0.GitDiffFile() — which does not compile. This package simply
// existing (its generated code compiling) is the primary regression guard; the
// round-trip below additionally exercises the dispatch at runtime.
//
// See https://github.com/oapi-codegen/oapi-codegen/issues/2470
func TestValueByDiscriminatorExternalRef(t *testing.T) {
	patch := "@@ -1 +1 @@"
	var f api.PRFile
	require.NoError(t, f.FromExternalRef0GitDiffFile(common.GitDiffFile{
		DiffType: "GitDiffFile",
		Patch:    &patch,
	}))

	v, err := f.ValueByDiscriminator()
	require.NoError(t, err)

	got, ok := v.(common.GitDiffFile)
	require.True(t, ok, "expected common.GitDiffFile, got %T", v)
	require.NotNil(t, got.Patch)
	assert.Equal(t, patch, *got.Patch)
}
