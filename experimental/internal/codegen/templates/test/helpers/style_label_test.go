package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStyleLabelParam_Primitive(t *testing.T) {
	result, err := StyleLabelParam("id", ParamLocationPath, 5)
	require.NoError(t, err)
	assert.Equal(t, ".5", result)
}

func TestStyleLabelParam_String(t *testing.T) {
	result, err := StyleLabelParam("color", ParamLocationPath, "blue")
	require.NoError(t, err)
	assert.Equal(t, ".blue", result)
}

func TestStyleLabelParam_StringSlice(t *testing.T) {
	result, err := StyleLabelParam("tags", ParamLocationPath, []string{"a", "b", "c"})
	require.NoError(t, err)
	assert.Equal(t, ".a,b,c", result)
}

func TestStyleLabelParam_Struct(t *testing.T) {
	type obj struct {
		Color string `json:"color"`
		Size  int    `json:"size"`
	}
	result, err := StyleLabelParam("filter", ParamLocationPath, obj{Color: "red", Size: 10})
	require.NoError(t, err)
	assert.Equal(t, ".color,red,size,10", result)
}

func TestStyleLabelParam_Roundtrip_Primitive(t *testing.T) {
	styled, err := StyleLabelParam("id", ParamLocationPath, 42)
	require.NoError(t, err)

	var result int
	err = BindLabelParam("id", ParamLocationPath, styled, &result)
	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestStyleLabelParam_Roundtrip_StringSlice(t *testing.T) {
	original := []string{"x", "y", "z"}
	styled, err := StyleLabelParam("items", ParamLocationPath, original)
	require.NoError(t, err)

	var result []string
	err = BindLabelParam("items", ParamLocationPath, styled, &result)
	require.NoError(t, err)
	assert.Equal(t, original, result)
}

func TestStyleLabelParam_Roundtrip_Struct(t *testing.T) {
	type obj struct {
		Color string `json:"color"`
		Size  string `json:"size"`
	}
	original := obj{Color: "blue", Size: "large"}
	styled, err := StyleLabelParam("filter", ParamLocationPath, original)
	require.NoError(t, err)

	var result obj
	err = BindLabelParam("filter", ParamLocationPath, styled, &result)
	require.NoError(t, err)
	assert.Equal(t, original, result)
}

func TestStyleLabelExplodeParam_Primitive(t *testing.T) {
	result, err := StyleLabelExplodeParam("id", ParamLocationPath, 5)
	require.NoError(t, err)
	assert.Equal(t, ".5", result)
}

func TestStyleLabelExplodeParam_StringSlice(t *testing.T) {
	result, err := StyleLabelExplodeParam("tags", ParamLocationPath, []string{"a", "b", "c"})
	require.NoError(t, err)
	assert.Equal(t, ".a.b.c", result)
}

func TestStyleLabelExplodeParam_Struct(t *testing.T) {
	type obj struct {
		Color string `json:"color"`
		Size  int    `json:"size"`
	}
	result, err := StyleLabelExplodeParam("filter", ParamLocationPath, obj{Color: "red", Size: 10})
	require.NoError(t, err)
	assert.Equal(t, ".color=red.size=10", result)
}

func TestStyleLabelExplodeParam_Roundtrip_Primitive(t *testing.T) {
	styled, err := StyleLabelExplodeParam("id", ParamLocationPath, 42)
	require.NoError(t, err)

	var result int
	err = BindLabelExplodeParam("id", ParamLocationPath, styled, &result)
	require.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestStyleLabelExplodeParam_Roundtrip_StringSlice(t *testing.T) {
	original := []string{"x", "y", "z"}
	styled, err := StyleLabelExplodeParam("items", ParamLocationPath, original)
	require.NoError(t, err)

	var result []string
	err = BindLabelExplodeParam("items", ParamLocationPath, styled, &result)
	require.NoError(t, err)
	assert.Equal(t, original, result)
}

func TestStyleLabelExplodeParam_Roundtrip_Struct(t *testing.T) {
	type obj struct {
		Color string `json:"color"`
		Size  string `json:"size"`
	}
	original := obj{Color: "blue", Size: "large"}
	styled, err := StyleLabelExplodeParam("filter", ParamLocationPath, original)
	require.NoError(t, err)

	var result obj
	err = BindLabelExplodeParam("filter", ParamLocationPath, styled, &result)
	require.NoError(t, err)
	assert.Equal(t, original, result)
}
