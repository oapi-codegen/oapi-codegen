package codegen

import (
	"encoding/json"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeSchemas(t *testing.T) {
	t.Run("with default", func(t *testing.T) {
		tests := []struct {
			name        string
			text        string
			wantDefault string
			mustFail    bool
		}{
			{
				name:        "rhs has default",
				text:        `[{"type": "string", "enum": ["a", "b"]}, {"default": "a"}]`,
				wantDefault: "a",
			},
			{
				name:        "lhs has default",
				text:        `[{"default": "a"}, {"type": "string", "enum": ["a", "b"]}]`,
				wantDefault: "a",
			},
			{
				name:     "both have default",
				text:     `[{"default": "a"}, {"type": "string", "enum": ["a", "b"], "default": "b"}]`,
				mustFail: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var schemas []*openapi3.SchemaRef
				require.NoError(t, json.Unmarshal([]byte(tt.text), &schemas), "unmarshal schemas")

				got, err := MergeSchemas(schemas, nil)
				if tt.mustFail {
					assert.Error(t, err, "conflicting default values")
					return
				}
				assert.NoError(t, err)
				assert.Equal(t, tt.wantDefault, got.OAPISchema.Default)
			})
		}
	})
}
