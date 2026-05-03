package externalref

import (
	"testing"

	packageA "github.com/oapi-codegen/oapi-codegen/v2/internal/test/externalref/packageA"
	packageB "github.com/oapi-codegen/oapi-codegen/v2/internal/test/externalref/packageB"
	petstore "github.com/oapi-codegen/oapi-codegen/v2/internal/test/externalref/petstore"
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
