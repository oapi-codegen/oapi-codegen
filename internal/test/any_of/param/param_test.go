package param_test

import (
	"testing"

	"github.com/deepmap/oapi-codegen/internal/test/any_of/param"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnyOfParameter(t *testing.T) {
	var p param.GetTestParams

	p.Test = new(param.Test)
	err := p.Test.FromTest0(param.Test0{
		Item1: "foo",
		Item2: "bar",
	})
	require.NoError(t, err)

	hp, err := param.NewGetTestRequest("", &p)
	assert.NoError(t, err)
	assert.Equal(t, "/test?item1=foo&item2=bar", hp.URL.String())
}

func TestArrayOfAnyOfParameter(t *testing.T) {
	var p param.GetTestParams

	p.Test2 = &[]param.Test2{
		{},
	}
	err := (*p.Test2)[0].FromTest20(100)
	require.NoError(t, err)

	hp, err := param.NewGetTestRequest("", &p)
	assert.NoError(t, err)
	assert.Equal(t, "/test?test2=100", hp.URL.String())
}
