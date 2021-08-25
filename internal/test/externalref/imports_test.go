package externalref

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/deepmap/oapi-codegen/internal/test/externalref/packageA"
	"github.com/deepmap/oapi-codegen/internal/test/externalref/packageB"
)

func TestParameters(t *testing.T) {
	b := &packageB.ObjectB{}
	_ = Container{
		ObjectA: &packageA.ObjectA{ObjectB: b},
		ObjectB: b,
	}
}

func TestGetSpec(t *testing.T) {
	_, err := packageB.GetSpec()
	require.Nil(t, err)

	_, err = packageA.GetSpec()
	require.Nil(t, err)

	_, err = GetSpec()
	require.Nil(t, err)
}
