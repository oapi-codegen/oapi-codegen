package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseInputMapping(t *testing.T) {
	var src string
	var expected, parsed map[string]string
	var err error

	src = "key1:value1,key2:value2"
	expected = map[string]string{"key1": "value1", "key2": "value2"}
	parsed, err = ParseCommandlineMap(src)
	require.NoError(t, err)
	assert.Equal(t, expected, parsed)

	src = `key1:"value1,value2",key2:value3`
	expected = map[string]string{"key1": "value1,value2", "key2": "value3"}
	parsed, err = ParseCommandlineMap(src)
	require.NoError(t, err)
	assert.Equal(t, expected, parsed)

	src = `key1:"value1,value2,key2:value3"`
	expected = map[string]string{"key1": "value1,value2,key2:value3"}
	parsed, err = ParseCommandlineMap(src)
	require.NoError(t, err)
	assert.Equal(t, expected, parsed)

	src = `"key1,key2":value1`
	expected = map[string]string{"key1,key2": "value1"}
	parsed, err = ParseCommandlineMap(src)
	require.NoError(t, err)
	assert.Equal(t, expected, parsed)
}

func TestSplitString(t *testing.T) {
	var src string
	var expected, result []string
	var err error

	src = "1,2,3"
	expected = []string{"1", "2", "3"}
	result = splitString(src, ',')
	require.NoError(t, err)
	assert.Equal(t, expected, result)

	src = `"1,2",3`
	expected = []string{`"1,2"`, "3"}
	result = splitString(src, ',')
	require.NoError(t, err)
	assert.Equal(t, expected, result)

	src = `1,"2,3",`
	expected = []string{"1", `"2,3"`, ""}
	result = splitString(src, ',')
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}
