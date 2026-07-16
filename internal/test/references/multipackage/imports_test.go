package referencesmultipackage

import (
	"context"
	"testing"

	packageA "github.com/oapi-codegen/oapi-codegen/v2/internal/test/references/multipackage/packageA"
	packageB "github.com/oapi-codegen/oapi-codegen/v2/internal/test/references/multipackage/packageB"
	petstore "github.com/oapi-codegen/oapi-codegen/v2/internal/test/references/multipackage/petstore"
	"github.com/stretchr/testify/require"
)

func TestParameters(t *testing.T) {
	b := &packageB.ObjectB{}
	_ = Container{
		ObjectA: &packageA.ObjectA{ObjectB: b},
		ObjectB: b,
	}
}

func TestGetSwagger(t *testing.T) {
	_, err := packageB.GetSpec()
	require.Nil(t, err)

	_, err = packageA.GetSpec()
	require.Nil(t, err)

	_, err = petstore.GetSpec()
	require.Nil(t, err)

	_, err = GetSpec()
	require.Nil(t, err)
}

// TestSecuritySchemeScopesShared verifies that the scopes context key of a
// security scheme $ref'd from an import-mapped spec is shared across the
// generated packages: this package's BearerAuthScopes aliases packageA's, so
// a context value stored under one key is retrievable with the other.
// Reproduces https://github.com/oapi-codegen/oapi-codegen/issues/2383
func TestSecuritySchemeScopesShared(t *testing.T) {
	ctx := context.WithValue(context.Background(), packageA.BearerAuthScopes, []string{"read"})
	require.Equal(t, []string{"read"}, ctx.Value(BearerAuthScopes))
}
