package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStyleSimpleParam_Primitive(t *testing.T) {
	result, err := StyleSimpleParam("id", ParamLocationPath, 5)
	require.NoError(t, err)
	assert.Equal(t, "5", result)
}

func TestStyleSimpleParam_String(t *testing.T) {
	result, err := StyleSimpleParam("name", ParamLocationPath, "hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", result)
}

func TestStyleSimpleParam_StringSlice(t *testing.T) {
	result, err := StyleSimpleParam("tags", ParamLocationPath, []string{"a", "b", "c"})
	require.NoError(t, err)
	assert.Equal(t, "a,b,c", result)
}

func TestStyleSimpleParam_Struct(t *testing.T) {
	type obj struct {
		Color string `json:"color"`
		Size  int    `json:"size"`
	}
	result, err := StyleSimpleParam("filter", ParamLocationPath, obj{Color: "red", Size: 10})
	require.NoError(t, err)
	assert.Equal(t, "color,red,size,10", result)
}

func TestStyleSimpleParam_Roundtrip_Primitive(t *testing.T) {
	styled, err := StyleSimpleParam("id", ParamLocationPath, 42)
	require.NoError(t, err)

	var result int
	err = BindSimpleParam("id", ParamLocationPath, styled, &result)
	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestStyleSimpleParam_Roundtrip_StringSlice(t *testing.T) {
	original := []string{"x", "y", "z"}
	styled, err := StyleSimpleParam("items", ParamLocationPath, original)
	require.NoError(t, err)

	var result []string
	err = BindSimpleParam("items", ParamLocationPath, styled, &result)
	require.NoError(t, err)
	assert.Equal(t, original, result)
}

func TestStyleSimpleParam_Roundtrip_Struct(t *testing.T) {
	type obj struct {
		Color string `json:"color"`
		Size  string `json:"size"`
	}
	original := obj{Color: "blue", Size: "large"}
	styled, err := StyleSimpleParam("filter", ParamLocationPath, original)
	require.NoError(t, err)

	var result obj
	err = BindSimpleParam("filter", ParamLocationPath, styled, &result)
	require.NoError(t, err)
	assert.Equal(t, original, result)
}

func TestStyleSimpleExplodeParam_Primitive(t *testing.T) {
	result, err := StyleSimpleExplodeParam("id", ParamLocationPath, 5)
	require.NoError(t, err)
	assert.Equal(t, "5", result)
}

func TestStyleSimpleExplodeParam_StringSlice(t *testing.T) {
	result, err := StyleSimpleExplodeParam("tags", ParamLocationPath, []string{"a", "b", "c"})
	require.NoError(t, err)
	assert.Equal(t, "a,b,c", result)
}

func TestStyleSimpleExplodeParam_Struct(t *testing.T) {
	type obj struct {
		Color string `json:"color"`
		Size  int    `json:"size"`
	}
	result, err := StyleSimpleExplodeParam("filter", ParamLocationPath, obj{Color: "red", Size: 10})
	require.NoError(t, err)
	assert.Equal(t, "color=red,size=10", result)
}

func TestStyleSimpleExplodeParam_Roundtrip_Struct(t *testing.T) {
	type obj struct {
		Color string `json:"color"`
		Size  string `json:"size"`
	}
	original := obj{Color: "blue", Size: "large"}
	styled, err := StyleSimpleExplodeParam("filter", ParamLocationPath, original)
	require.NoError(t, err)

	var result obj
	err = BindSimpleExplodeParam("filter", ParamLocationPath, styled, &result)
	require.NoError(t, err)
	assert.Equal(t, original, result)
}

func TestStyleSimpleExplodeParam_Roundtrip_StringSlice(t *testing.T) {
	original := []string{"a", "b", "c"}
	styled, err := StyleSimpleExplodeParam("items", ParamLocationPath, original)
	require.NoError(t, err)

	var result []string
	err = BindSimpleExplodeParam("items", ParamLocationPath, styled, &result)
	require.NoError(t, err)
	assert.Equal(t, original, result)
}
