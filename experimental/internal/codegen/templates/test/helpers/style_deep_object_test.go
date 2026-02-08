package helpers

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStyleDeepObjectParam_Struct(t *testing.T) {
	type obj struct {
		Color string `json:"color"`
		Size  int    `json:"size"`
	}
	result, err := StyleDeepObjectParam("filter", ParamLocationQuery, obj{Color: "red", Size: 10})
	require.NoError(t, err)
	assert.Equal(t, "filter[color]=red&filter[size]=10", result)
}

func TestStyleDeepObjectParam_NestedStruct(t *testing.T) {
	type inner struct {
		City string `json:"city"`
	}
	type outer struct {
		Name    string `json:"name"`
		Address inner  `json:"address"`
	}
	result, err := StyleDeepObjectParam("user", ParamLocationQuery, outer{
		Name:    "alice",
		Address: inner{City: "NYC"},
	})
	require.NoError(t, err)
	assert.Equal(t, "user[address][city]=NYC&user[name]=alice", result)
}

func TestStyleDeepObjectParam_Roundtrip_Struct(t *testing.T) {
	type obj struct {
		Color string `json:"color"`
		Size  int    `json:"size"`
	}
	original := obj{Color: "blue", Size: 42}
	styled, err := StyleDeepObjectParam("filter", ParamLocationQuery, original)
	require.NoError(t, err)

	// Parse to url.Values
	vals, err := url.ParseQuery(styled)
	require.NoError(t, err)

	var result obj
	err = BindDeepObjectParam("filter", vals, &result)
	require.NoError(t, err)
	assert.Equal(t, original, result)
}

func TestStyleDeepObjectParam_Roundtrip_NestedStruct(t *testing.T) {
	type inner struct {
		City string `json:"city"`
	}
	type outer struct {
		Name    string `json:"name"`
		Address inner  `json:"address"`
	}
	original := outer{
		Name:    "alice",
		Address: inner{City: "NYC"},
	}
	styled, err := StyleDeepObjectParam("user", ParamLocationQuery, original)
	require.NoError(t, err)

	// Parse to url.Values
	vals, err := url.ParseQuery(styled)
	require.NoError(t, err)

	var result outer
	err = BindDeepObjectParam("user", vals, &result)
	require.NoError(t, err)
	assert.Equal(t, original, result)
}

func TestStyleDeepObjectParam_WithSlice(t *testing.T) {
	type obj struct {
		Tags []string `json:"tags"`
	}
	result, err := StyleDeepObjectParam("filter", ParamLocationQuery, obj{Tags: []string{"a", "b"}})
	require.NoError(t, err)
	assert.Equal(t, "filter[tags][0]=a&filter[tags][1]=b", result)
}

func TestStyleDeepObjectParam_Roundtrip_WithSlice(t *testing.T) {
	type obj struct {
		Tags []string `json:"tags"`
	}
	original := obj{Tags: []string{"a", "b"}}
	styled, err := StyleDeepObjectParam("filter", ParamLocationQuery, original)
	require.NoError(t, err)

	vals, err := url.ParseQuery(styled)
	require.NoError(t, err)

	var result obj
	err = BindDeepObjectParam("filter", vals, &result)
	require.NoError(t, err)
	assert.Equal(t, original, result)
}
