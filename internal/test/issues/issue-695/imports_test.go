package externalref

import (
	"testing"

	"github.com/stretchr/testify/require"

	packageA "github.com/deepmap/oapi-codegen/v2/internal/test/issues/issue-695/packageA"
	packageB "github.com/deepmap/oapi-codegen/v2/internal/test/issues/issue-695/packageB"
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

	_, err = GetSwagger()
	require.Nil(t, err)
}

func TestExternalRefIsResolvedCorrectly(t *testing.T) {
	resp := packageA.TestApiResponse{}

	expected := packageB.ResponseC{}

	resp.JSON400 = &expected
}
