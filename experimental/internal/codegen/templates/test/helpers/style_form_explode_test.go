package helpers

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStyleFormExplodeParam_Primitive(t *testing.T) {
	result, err := StyleFormExplodeParam("color", ParamLocationQuery, "blue")
	require.NoError(t, err)
	assert.Equal(t, "color=blue", result)
}

func TestStyleFormExplodeParam_Int(t *testing.T) {
	result, err := StyleFormExplodeParam("count", ParamLocationQuery, 5)
	require.NoError(t, err)
	assert.Equal(t, "count=5", result)
}

func TestStyleFormExplodeParam_StringSlice(t *testing.T) {
	result, err := StyleFormExplodeParam("tags", ParamLocationQuery, []string{"a", "b", "c"})
	require.NoError(t, err)
	assert.Equal(t, "tags=a&tags=b&tags=c", result)
}

func TestStyleFormExplodeParam_Struct(t *testing.T) {
	type obj struct {
		Color string `json:"color"`
		Size  int    `json:"size"`
	}
	result, err := StyleFormExplodeParam("filter", ParamLocationQuery, obj{Color: "red", Size: 10})
	require.NoError(t, err)
	assert.Equal(t, "color=red&size=10", result)
}

func TestStyleFormExplodeParam_Roundtrip_Primitive(t *testing.T) {
	styled, err := StyleFormExplodeParam("color", ParamLocationQuery, "blue")
	require.NoError(t, err)

	// Parse to url.Values
	vals, err := url.ParseQuery(styled)
	require.NoError(t, err)

	var result string
	err = BindFormExplodeParam("color", true, vals, &result)
	require.NoError(t, err)
	assert.Equal(t, "blue", result)
}

func TestStyleFormExplodeParam_Roundtrip_StringSlice(t *testing.T) {
	original := []string{"a", "b", "c"}
	styled, err := StyleFormExplodeParam("items", ParamLocationQuery, original)
	require.NoError(t, err)

	// Parse to url.Values
	vals, err := url.ParseQuery(styled)
	require.NoError(t, err)

	var result []string
	err = BindFormExplodeParam("items", true, vals, &result)
	require.NoError(t, err)
	assert.Equal(t, original, result)
}

func TestStyleFormExplodeParam_Roundtrip_Struct(t *testing.T) {
	type obj struct {
		Color string `json:"color"`
		Size  string `json:"size"`
	}
	original := obj{Color: "blue", Size: "large"}
	styled, err := StyleFormExplodeParam("filter", ParamLocationQuery, original)
	require.NoError(t, err)

	// Parse to url.Values
	vals, err := url.ParseQuery(styled)
	require.NoError(t, err)

	var result obj
	err = BindFormExplodeParam("filter", true, vals, &result)
	require.NoError(t, err)
	assert.Equal(t, original, result)
}

func TestBindFormExplodeParam_OptionalMissing(t *testing.T) {
	vals := url.Values{}

	var result *string
	err := BindFormExplodeParam("missing", false, vals, &result)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestBindFormExplodeParam_RequiredMissing(t *testing.T) {
	vals := url.Values{}

	var result string
	err := BindFormExplodeParam("required", true, vals, &result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}
