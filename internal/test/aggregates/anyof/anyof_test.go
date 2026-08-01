package aggregatesanyof_test

import (
	"testing"

	aggregatesanyof "github.com/oapi-codegen/oapi-codegen/v2/internal/test/aggregates/anyof"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// any_of/codegen/inline: compile-only — no assertions in original source.
// Verifies that inline anyOf in a response body array (no named union schema)
// generates valid Go code that compiles.
var _ = aggregatesanyof.GetInlinePets200JSONResponseBody_Data_Item{}

// any_of/codegen/ref_schema: compile-only — no assertions in original source.
// Verifies that anyOf via a named $ref schema generates valid Go code that compiles.
var _ = aggregatesanyof.GetRefPetsDto{}
var _ = aggregatesanyof.GetRefPetsDto_Data{}

// issues/issue-1189: compile-only — no assertions in original source.
// Verifies that a schema with combined anyOf+allOf+oneOf fields generates
// valid Go code with proper union accessor methods that compile.
var _ = aggregatesanyof.Issue1189Test{}

// any_of/param: assertions ported from any_of/param/param_test.go.
// Tests that anyOf/oneOf query-parameter types serialize correctly.

// TestAnyOfParameter verifies that a ParamAnyOf query parameter serializes
// its struct variant via the union accessor (ported from any_of/param).
func TestAnyOfParameter(t *testing.T) {
	var p aggregatesanyof.GetParamTestParams

	p.Test = new(aggregatesanyof.ParamAnyOf)
	err := p.Test.FromParamAnyOf0(aggregatesanyof.ParamAnyOf0{
		Item1: "foo",
		Item2: "bar",
	})
	require.NoError(t, err)

	hp, err := aggregatesanyof.NewGetParamTestRequest("", &p)
	assert.NoError(t, err)
	assert.Equal(t, "/param/test?item1=foo&item2=bar", hp.URL.String())
}

// TestArrayOfAnyOfParameter verifies that a []ParamOneOf query parameter
// serializes its integer variant (ported from any_of/param).
func TestArrayOfAnyOfParameter(t *testing.T) {
	var p aggregatesanyof.GetParamTestParams

	p.Test2 = &[]aggregatesanyof.ParamOneOf{
		{},
	}
	err := (*p.Test2)[0].FromParamOneOf0(100)
	require.NoError(t, err)

	hp, err := aggregatesanyof.NewGetParamTestRequest("", &p)
	assert.NoError(t, err)
	assert.Equal(t, "/param/test?test2=100", hp.URL.String())
}
