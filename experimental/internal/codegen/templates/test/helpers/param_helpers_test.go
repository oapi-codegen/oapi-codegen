package helpers

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrimitiveToString(t *testing.T) {
	testCases := map[string]struct {
		value    any
		expected string
	}{
		"int": {
			value:    42,
			expected: "42",
		},
		"float64": {
			value:    3.14,
			expected: "3.14",
		},
		"bool true": {
			value:    true,
			expected: "true",
		},
		"bool false": {
			value:    false,
			expected: "false",
		},
		"string": {
			value:    "hello",
			expected: "hello",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result, err := primitiveToString(tc.value)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestPrimitiveToString_Time(t *testing.T) {
	tm := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	result, err := primitiveToString(tm)
	require.NoError(t, err)
	assert.Equal(t, "2024-06-15T10:30:00Z", result)
}

func TestPrimitiveToString_UUID(t *testing.T) {
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	result, err := primitiveToString(id)
	require.NoError(t, err)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", result)
}

func TestPrimitiveToString_Date(t *testing.T) {
	d := Date{Time: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)}
	result, err := primitiveToString(d)
	require.NoError(t, err)
	assert.Equal(t, "2024-06-15", result)
}

func TestBindStringToObject(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		var s string
		err := BindStringToObject("hello", &s)
		require.NoError(t, err)
		assert.Equal(t, "hello", s)
	})

	t.Run("int", func(t *testing.T) {
		var i int
		err := BindStringToObject("42", &i)
		require.NoError(t, err)
		assert.Equal(t, 42, i)
	})

	t.Run("bool", func(t *testing.T) {
		var b bool
		err := BindStringToObject("true", &b)
		require.NoError(t, err)
		assert.True(t, b)
	})

	t.Run("float64", func(t *testing.T) {
		var f float64
		err := BindStringToObject("3.14", &f)
		require.NoError(t, err)
		assert.InDelta(t, 3.14, f, 0.001)
	})

	t.Run("invalid int returns error", func(t *testing.T) {
		var i int
		err := BindStringToObject("not_a_number", &i)
		assert.Error(t, err)
	})
}

func TestDate_MarshalUnmarshalRoundtrip(t *testing.T) {
	d := Date{Time: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)}

	text, err := d.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "2024-03-15", string(text))

	var d2 Date
	err = d2.UnmarshalText(text)
	require.NoError(t, err)
	assert.Equal(t, d.Time, d2.Time)
}

func TestDate_Format(t *testing.T) {
	d := Date{Time: time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC)}
	assert.Equal(t, "2024-12-25", d.Format(DateFormat))
	assert.Equal(t, "25 Dec 2024", d.Format("02 Jan 2006"))
}

func TestDate_UnmarshalText_InvalidFormat(t *testing.T) {
	var d Date
	err := d.UnmarshalText([]byte("not-a-date"))
	assert.Error(t, err)
}

func TestEscapeParameterString(t *testing.T) {
	t.Run("query location escapes", func(t *testing.T) {
		result := escapeParameterString("hello world", ParamLocationQuery)
		assert.Equal(t, "hello+world", result)
	})

	t.Run("path location escapes", func(t *testing.T) {
		result := escapeParameterString("hello world", ParamLocationPath)
		assert.Equal(t, "hello%20world", result)
	})

	t.Run("header location no escaping", func(t *testing.T) {
		result := escapeParameterString("hello world", ParamLocationHeader)
		assert.Equal(t, "hello world", result)
	})

	t.Run("cookie location no escaping", func(t *testing.T) {
		result := escapeParameterString("hello world", ParamLocationCookie)
		assert.Equal(t, "hello world", result)
	})
}

func TestUnescapeParameterString(t *testing.T) {
	t.Run("query location unescapes", func(t *testing.T) {
		result, err := unescapeParameterString("hello+world", ParamLocationQuery)
		require.NoError(t, err)
		assert.Equal(t, "hello world", result)
	})

	t.Run("path location unescapes", func(t *testing.T) {
		result, err := unescapeParameterString("hello%20world", ParamLocationPath)
		require.NoError(t, err)
		assert.Equal(t, "hello world", result)
	})

	t.Run("header location no change", func(t *testing.T) {
		result, err := unescapeParameterString("hello world", ParamLocationHeader)
		require.NoError(t, err)
		assert.Equal(t, "hello world", result)
	})
}
