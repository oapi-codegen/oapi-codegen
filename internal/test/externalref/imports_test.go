package externalref

import (
	"testing"

	packageA "github.com/deepmap/oapi-codegen/v2/internal/test/externalref/packageA"
	packageB "github.com/deepmap/oapi-codegen/v2/internal/test/externalref/packageB"
	petstore "github.com/deepmap/oapi-codegen/v2/internal/test/externalref/petstore"
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
	_, err := packageB.GetSwagger()
	require.Nil(t, err)

	_, err = packageA.GetSwagger()
	require.Nil(t, err)

	_, err = petstore.GetSwagger()
	require.Nil(t, err)

	_, err = GetSwagger()
	require.Nil(t, err)
}
