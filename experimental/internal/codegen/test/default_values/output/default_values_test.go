package output

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T {
	return &v
}

// TestSimpleDefaults tests ApplyDefaults on basic primitive types
func TestSimpleDefaults(t *testing.T) {
	t.Run("applies all defaults to empty struct", func(t *testing.T) {
		s := SimpleDefaults{}
		s.ApplyDefaults()

		require.NotNil(t, s.StringField)
		assert.Equal(t, "hello", *s.StringField)

		require.NotNil(t, s.IntField)
		assert.Equal(t, 42, *s.IntField)

		require.NotNil(t, s.BoolField)
		assert.Equal(t, true, *s.BoolField)

		require.NotNil(t, s.FloatField)
		assert.Equal(t, float32(3.14), *s.FloatField)

		require.NotNil(t, s.Int64Field)
		assert.Equal(t, int64(9223372036854775807), *s.Int64Field)
	})

	t.Run("does not overwrite existing values", func(t *testing.T) {
		s := SimpleDefaults{
			StringField: ptr("custom"),
			IntField:    ptr(100),
			BoolField:   ptr(false),
			FloatField:  ptr(float32(1.5)),
			Int64Field:  ptr(int64(123)),
		}
		s.ApplyDefaults()

		assert.Equal(t, "custom", *s.StringField)
		assert.Equal(t, 100, *s.IntField)
		assert.Equal(t, false, *s.BoolField)
		assert.Equal(t, float32(1.5), *s.FloatField)
		assert.Equal(t, int64(123), *s.Int64Field)
	})

	t.Run("applies defaults after unmarshaling empty object", func(t *testing.T) {
		input := `{}`
		var s SimpleDefaults
		err := json.Unmarshal([]byte(input), &s)
		require.NoError(t, err)

		s.ApplyDefaults()

		assert.Equal(t, "hello", *s.StringField)
		assert.Equal(t, 42, *s.IntField)
	})

	t.Run("applies defaults after unmarshaling partial object", func(t *testing.T) {
		input := `{"stringField": "from-json"}`
		var s SimpleDefaults
		err := json.Unmarshal([]byte(input), &s)
		require.NoError(t, err)

		s.ApplyDefaults()

		assert.Equal(t, "from-json", *s.StringField) // from JSON
		assert.Equal(t, 42, *s.IntField)              // from default
	})
}

// TestNestedDefaults tests ApplyDefaults recursion into nested structs
func TestNestedDefaults(t *testing.T) {
	t.Run("applies defaults to parent and recurses to children", func(t *testing.T) {
		n := NestedDefaults{
			Child:       &SimpleDefaults{},
			InlineChild: &NestedDefaultsInlineChild{},
		}
		n.ApplyDefaults()

		// Parent defaults
		require.NotNil(t, n.Name)
		assert.Equal(t, "parent", *n.Name)

		// Child defaults (recursion)
		require.NotNil(t, n.Child.StringField)
		assert.Equal(t, "hello", *n.Child.StringField)
		require.NotNil(t, n.Child.IntField)
		assert.Equal(t, 42, *n.Child.IntField)

		// Inline child defaults (recursion)
		require.NotNil(t, n.InlineChild.Label)
		assert.Equal(t, "inline-default", *n.InlineChild.Label)
		require.NotNil(t, n.InlineChild.Value)
		assert.Equal(t, 100, *n.InlineChild.Value)
	})

	t.Run("does not recurse into nil children", func(t *testing.T) {
		n := NestedDefaults{}
		n.ApplyDefaults()

		// Parent defaults applied
		require.NotNil(t, n.Name)
		assert.Equal(t, "parent", *n.Name)

		// Children are still nil (not created)
		assert.Nil(t, n.Child)
		assert.Nil(t, n.InlineChild)
	})

	t.Run("applies defaults after unmarshaling nested JSON", func(t *testing.T) {
		input := `{"child": {"stringField": "from-child"}, "inlineChild": {}}`
		var n NestedDefaults
		err := json.Unmarshal([]byte(input), &n)
		require.NoError(t, err)

		n.ApplyDefaults()

		// Parent defaults
		assert.Equal(t, "parent", *n.Name)

		// Child - one field from JSON, others from defaults
		assert.Equal(t, "from-child", *n.Child.StringField)
		assert.Equal(t, 42, *n.Child.IntField)

		// Inline child - all defaults
		assert.Equal(t, "inline-default", *n.InlineChild.Label)
		assert.Equal(t, 100, *n.InlineChild.Value)
	})
}

// TestMapWithDefaults tests ApplyDefaults on structs with additionalProperties
func TestMapWithDefaults(t *testing.T) {
	t.Run("applies defaults to known fields", func(t *testing.T) {
		m := MapWithDefaults{}
		m.ApplyDefaults()

		require.NotNil(t, m.Prefix)
		assert.Equal(t, "map-", *m.Prefix)
	})

	t.Run("does not affect additional properties", func(t *testing.T) {
		m := MapWithDefaults{
			AdditionalProperties: map[string]string{
				"extra": "value",
			},
		}
		m.ApplyDefaults()

		assert.Equal(t, "map-", *m.Prefix)
		assert.Equal(t, "value", m.AdditionalProperties["extra"])
	})
}

// TestArrayDefaults tests ApplyDefaults on structs with array fields
func TestArrayDefaults(t *testing.T) {
	t.Run("applies defaults to non-array fields", func(t *testing.T) {
		a := ArrayDefaults{}
		a.ApplyDefaults()

		require.NotNil(t, a.Count)
		assert.Equal(t, 0, *a.Count)
		// Array field is not touched (no default generation for arrays currently)
		assert.Nil(t, a.Items)
	})
}

// TestAnyOfWithDefaults tests ApplyDefaults on anyOf variant members
func TestAnyOfWithDefaults(t *testing.T) {
	t.Run("applies defaults to anyOf variant 0", func(t *testing.T) {
		v0 := AnyOfWithDefaultsValueAnyOf0{}
		v0.ApplyDefaults()

		require.NotNil(t, v0.StringVal)
		assert.Equal(t, "default-string", *v0.StringVal)
	})

	t.Run("applies defaults to anyOf variant 1", func(t *testing.T) {
		v1 := AnyOfWithDefaultsValueAnyOf1{}
		v1.ApplyDefaults()

		require.NotNil(t, v1.IntVal)
		assert.Equal(t, 999, *v1.IntVal)
	})
}

// TestOneOfWithDefaults tests ApplyDefaults on oneOf variant members
func TestOneOfWithDefaults(t *testing.T) {
	t.Run("applies defaults to oneOf variant 0", func(t *testing.T) {
		v0 := OneOfWithDefaultsVariantOneOf0{}
		v0.ApplyDefaults()

		require.NotNil(t, v0.OptionA)
		assert.Equal(t, "option-a-default", *v0.OptionA)
	})

	t.Run("applies defaults to oneOf variant 1", func(t *testing.T) {
		v1 := OneOfWithDefaultsVariantOneOf1{}
		v1.ApplyDefaults()

		require.NotNil(t, v1.OptionB)
		assert.Equal(t, 123, *v1.OptionB)
	})
}

// TestAllOfWithDefaults tests ApplyDefaults on allOf merged structs
func TestAllOfWithDefaults(t *testing.T) {
	t.Run("applies defaults from all merged schemas", func(t *testing.T) {
		a := AllOfWithDefaults{}
		a.ApplyDefaults()

		// Default from allOf/0
		require.NotNil(t, a.Base)
		assert.Equal(t, "base-value", *a.Base)

		// Default from allOf/1
		require.NotNil(t, a.Extended)
		assert.Equal(t, 50, *a.Extended)
	})
}

// TestDeepNesting tests ApplyDefaults recursion through multiple levels
func TestDeepNesting(t *testing.T) {
	t.Run("recurses through all levels", func(t *testing.T) {
		d := DeepNesting{
			Level1: &DeepNestingLevel1{
				Level2: &DeepNestingLevel1Level2{
					Level3: &DeepNestingLevel1Level2Level3{},
				},
			},
		}
		d.ApplyDefaults()

		// Level 1 defaults
		require.NotNil(t, d.Level1.Name)
		assert.Equal(t, "level1-name", *d.Level1.Name)

		// Level 2 defaults
		require.NotNil(t, d.Level1.Level2.Count)
		assert.Equal(t, 2, *d.Level1.Level2.Count)

		// Level 3 defaults
		require.NotNil(t, d.Level1.Level2.Level3.Enabled)
		assert.Equal(t, false, *d.Level1.Level2.Level3.Enabled)
	})

	t.Run("stops at nil levels", func(t *testing.T) {
		d := DeepNesting{
			Level1: &DeepNestingLevel1{
				// Level2 is nil
			},
		}
		d.ApplyDefaults()

		assert.Equal(t, "level1-name", *d.Level1.Name)
		assert.Nil(t, d.Level1.Level2)
	})
}

// TestRequiredAndOptional tests ApplyDefaults behavior with required fields
func TestRequiredAndOptional(t *testing.T) {
	t.Run("applies defaults only to optional pointer fields", func(t *testing.T) {
		r := RequiredAndOptional{
			RequiredWithDefault: "set-by-user",
			RequiredNoDefault:   "also-set",
		}
		r.ApplyDefaults()

		// Required fields are value types, not pointers, so they don't get defaults applied
		assert.Equal(t, "set-by-user", r.RequiredWithDefault)
		assert.Equal(t, "also-set", r.RequiredNoDefault)

		// Optional fields with defaults get defaults applied
		require.NotNil(t, r.OptionalWithDefault)
		assert.Equal(t, "optional-default", *r.OptionalWithDefault)

		// Optional fields without defaults stay nil
		assert.Nil(t, r.OptionalNoDefault)
	})
}

// TestApplyDefaultsIdempotent tests that ApplyDefaults can be called multiple times
func TestApplyDefaultsIdempotent(t *testing.T) {
	t.Run("multiple calls have same effect", func(t *testing.T) {
		s := SimpleDefaults{}
		s.ApplyDefaults()
		s.ApplyDefaults()
		s.ApplyDefaults()

		assert.Equal(t, "hello", *s.StringField)
		assert.Equal(t, 42, *s.IntField)
	})
}

// TestApplyDefaultsChain tests typical usage pattern: unmarshal then apply defaults
func TestApplyDefaultsChain(t *testing.T) {
	t.Run("unmarshal partial JSON then apply defaults", func(t *testing.T) {
		input := `{
			"level1": {
				"level2": {
					"level3": {}
				}
			}
		}`

		var d DeepNesting
		err := json.Unmarshal([]byte(input), &d)
		require.NoError(t, err)

		d.ApplyDefaults()

		// All defaults applied at all levels
		assert.Equal(t, "level1-name", *d.Level1.Name)
		assert.Equal(t, 2, *d.Level1.Level2.Count)
		assert.Equal(t, false, *d.Level1.Level2.Level3.Enabled)
	})
}
