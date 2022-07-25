package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonMerge(t *testing.T) {
	t.Run("when object", func(t *testing.T) {
		t.Run("Merges properties defined in both objects", func(t *testing.T) {
			data := `{"foo": 1}`
			patch := `{"foo": null}`
			expected := `{"foo":null}`

			actual, err := JsonMerge([]byte(data), []byte(patch))
			assert.NoError(t, err)
			assert.Equal(t, expected, string(actual))
		})

		t.Run("Sets property defined in only src object", func(t *testing.T) {
			data := `{}`
			patch := `{"source":"merge-me"}`
			expected := `{"source":"merge-me"}`

			actual, err := JsonMerge([]byte(data), []byte(patch))
			assert.NoError(t, err)
			assert.Equal(t, expected, string(actual))
		})

		t.Run("Handles child objects", func(t *testing.T) {
			data := `{"channel":{"status":"valid"}}`
			patch := `{"channel":{"id":1}}`
			expected := `{"channel":{"id":1,"status":"valid"}}`

			actual, err := JsonMerge([]byte(data), []byte(patch))
			assert.NoError(t, err)
			assert.Equal(t, expected, string(actual))
		})

		t.Run("Handles empty objects", func(t *testing.T) {
			data := `{}`
			patch := `{}`
			expected := `{}`

			actual, err := JsonMerge([]byte(data), []byte(patch))
			assert.NoError(t, err)
			assert.Equal(t, expected, string(actual))
		})

		t.Run("Handles nil data", func(t *testing.T) {
			patch := `{"foo":"bar"}`
			expected := `{"foo":"bar"}`

			actual, err := JsonMerge(nil, []byte(patch))
			assert.NoError(t, err)
			assert.Equal(t, expected, string(actual))
		})

		t.Run("Handles nil patch", func(t *testing.T) {
			data := `{"foo":"bar"}`
			expected := `{"foo":"bar"}`

			actual, err := JsonMerge([]byte(data), nil)
			assert.NoError(t, err)
			assert.Equal(t, expected, string(actual))
		})
	})
	t.Run("when array", func(t *testing.T) {
		t.Run("it does not merge", func(t *testing.T) {
			data := `[{"foo": 1}]`
			patch := `[{"foo": null}]`
			expected := `[{"foo":1}]`

			actual, err := JsonMerge([]byte(data), []byte(patch))
			assert.NoError(t, err)
			assert.Equal(t, expected, string(actual))
		})
	})
}
