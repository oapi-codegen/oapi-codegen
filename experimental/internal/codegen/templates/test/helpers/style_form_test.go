package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStyleFormParam_Primitive(t *testing.T) {
	result, err := StyleFormParam("color", ParamLocationQuery, "blue")
	require.NoError(t, err)
	assert.Equal(t, "color=blue", result)
}

func TestStyleFormParam_Int(t *testing.T) {
	result, err := StyleFormParam("count", ParamLocationQuery, 5)
	require.NoError(t, err)
	assert.Equal(t, "count=5", result)
}

func TestStyleFormParam_StringSlice(t *testing.T) {
	result, err := StyleFormParam("tags", ParamLocationQuery, []string{"a", "b", "c"})
	require.NoError(t, err)
	assert.Equal(t, "tags=a,b,c", result)
}

func TestStyleFormParam_Struct(t *testing.T) {
	type obj struct {
		Color string `json:"color"`
		Size  int    `json:"size"`
	}
	result, err := StyleFormParam("filter", ParamLocationQuery, obj{Color: "red", Size: 10})
	require.NoError(t, err)
	assert.Equal(t, "filter=color,red,size,10", result)
}

func TestStyleFormParam_Roundtrip_Primitive(t *testing.T) {
	styled, err := StyleFormParam("color", ParamLocationQuery, "blue")
	require.NoError(t, err)
	// Form style: "color=blue" — the value part is "blue"
	assert.Equal(t, "color=blue", styled)

	// BindFormParam takes just the value (after splitting on =)
	var result string
	err = BindFormParam("color", ParamLocationQuery, "blue", &result)
	require.NoError(t, err)
	assert.Equal(t, "blue", result)
}

func TestStyleFormParam_Roundtrip_StringSlice(t *testing.T) {
	original := []string{"x", "y", "z"}
	styled, err := StyleFormParam("items", ParamLocationQuery, original)
	require.NoError(t, err)
	assert.Equal(t, "items=x,y,z", styled)

	// BindFormParam takes the value part
	var result []string
	err = BindFormParam("items", ParamLocationQuery, "x,y,z", &result)
	require.NoError(t, err)
	assert.Equal(t, original, result)
}

func TestStyleFormParam_Roundtrip_Struct(t *testing.T) {
	type obj struct {
		Color string `json:"color"`
		Size  string `json:"size"`
	}
	original := obj{Color: "blue", Size: "large"}
	styled, err := StyleFormParam("filter", ParamLocationQuery, original)
	require.NoError(t, err)
	assert.Equal(t, "filter=color,blue,size,large", styled)

	// BindFormParam takes the value part
	var result obj
	err = BindFormParam("filter", ParamLocationQuery, "color,blue,size,large", &result)
	require.NoError(t, err)
	assert.Equal(t, original, result)
}
