package codegen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultStructTagsConfig(t *testing.T) {
	t.Run("without yaml-tags", func(t *testing.T) {
		cfg := defaultStructTagsConfig(false)
		require.Len(t, cfg.Tags, 2)
		assert.Equal(t, "json", cfg.Tags[0].Name)
		assert.Equal(t, "form", cfg.Tags[1].Name)
	})

	t.Run("with yaml-tags", func(t *testing.T) {
		cfg := defaultStructTagsConfig(true)
		require.Len(t, cfg.Tags, 3)
		assert.Equal(t, "yaml", cfg.Tags[2].Name)
	})
}

func TestStructTagsConfigMerge(t *testing.T) {
	t.Run("empty user config keeps defaults", func(t *testing.T) {
		merged := defaultStructTagsConfig(false).Merge(StructTagsConfig{})
		assert.Equal(t, defaultStructTagsConfig(false), merged)
	})

	t.Run("matching name overrides default", func(t *testing.T) {
		merged := defaultStructTagsConfig(false).Merge(StructTagsConfig{
			Tags: []StructTagTemplate{{Name: "json", Template: `{{.FieldName}}`}},
		})
		require.Len(t, merged.Tags, 2)
		assert.Equal(t, "json", merged.Tags[0].Name)
		assert.Equal(t, `{{.FieldName}}`, merged.Tags[0].Template)
		assert.Equal(t, defaultFormTagTemplate, merged.Tags[1].Template)
	})

	t.Run("new name is appended", func(t *testing.T) {
		merged := defaultStructTagsConfig(false).Merge(StructTagsConfig{
			Tags: []StructTagTemplate{{Name: "db", Template: `{{.FieldName}}`}},
		})
		require.Len(t, merged.Tags, 3)
		assert.Equal(t, "db", merged.Tags[2].Name)
	})

	t.Run("user yaml entry clobbers the yaml-tags injected default", func(t *testing.T) {
		merged := defaultStructTagsConfig(true).Merge(StructTagsConfig{
			Tags: []StructTagTemplate{{Name: "yaml", Template: `{{.FieldName}}`}},
		})
		require.Len(t, merged.Tags, 3)
		assert.Equal(t, "yaml", merged.Tags[2].Name)
		assert.Equal(t, `{{.FieldName}}`, merged.Tags[2].Template)
	})
}

func TestNewStructTagGenerator(t *testing.T) {
	t.Run("defaults parse", func(t *testing.T) {
		_, err := newStructTagGenerator(defaultStructTagsConfig(true))
		require.NoError(t, err)
	})

	t.Run("parse error is reported", func(t *testing.T) {
		_, err := newStructTagGenerator(StructTagsConfig{
			Tags: []StructTagTemplate{{Name: "json", Template: `{{.FieldName`}},
		})
		require.ErrorContains(t, err, `invalid struct tag template for "json"`)
	})

	t.Run("execute error is reported", func(t *testing.T) {
		_, err := newStructTagGenerator(StructTagsConfig{
			Tags: []StructTagTemplate{{Name: "json", Template: `{{.NoSuchField}}`}},
		})
		require.ErrorContains(t, err, `struct tag template for "json" failed to render`)
	})

	t.Run("execute error inside a conditional branch is reported", func(t *testing.T) {
		_, err := newStructTagGenerator(StructTagsConfig{
			Tags: []StructTagTemplate{{Name: "json", Template: `{{if .IsOptional}}{{.NoSuchField}}{{end}}`}},
		})
		require.ErrorContains(t, err, `struct tag template for "json" failed to render`)

		_, err = newStructTagGenerator(StructTagsConfig{
			Tags: []StructTagTemplate{{Name: "json", Template: `{{if not .NeedsFormTag}}{{.NoSuchField}}{{end}}`}},
		})
		require.ErrorContains(t, err, `struct tag template for "json" failed to render`)
	})
}

func TestStructTagGeneratorGenerateTagsMap(t *testing.T) {
	defaults, err := newStructTagGenerator(defaultStructTagsConfig(true))
	require.NoError(t, err)

	t.Run("required field", func(t *testing.T) {
		tags := defaults.generateTagsMap(StructTagInfo{FieldName: "name"})
		assert.Equal(t, map[string]string{
			"json": "name",
			"yaml": "name",
		}, tags)
	})

	t.Run("optional field gets omitempty", func(t *testing.T) {
		tags := defaults.generateTagsMap(StructTagInfo{
			FieldName:  "name",
			IsOptional: true,
			OmitEmpty:  true,
		})
		assert.Equal(t, map[string]string{
			"json": "name,omitempty",
			"yaml": "name,omitempty",
		}, tags)
	})

	t.Run("omitzero only affects json", func(t *testing.T) {
		tags := defaults.generateTagsMap(StructTagInfo{
			FieldName: "name",
			OmitEmpty: true,
			OmitZero:  true,
		})
		assert.Equal(t, map[string]string{
			"json": "name,omitempty,omitzero",
			"yaml": "name,omitempty",
		}, tags)
	})

	t.Run("form tag only rendered when NeedsFormTag", func(t *testing.T) {
		tags := defaults.generateTagsMap(StructTagInfo{
			FieldName:    "name",
			NeedsFormTag: true,
		})
		assert.Equal(t, map[string]string{
			"json": "name",
			"yaml": "name",
			"form": "name",
		}, tags)
	})

	t.Run("empty render suppresses the tag", func(t *testing.T) {
		g, err := newStructTagGenerator(StructTagsConfig{
			Tags: []StructTagTemplate{
				{Name: "validate", Template: `{{if not .IsOptional}}required{{end}}`},
			},
		})
		require.NoError(t, err)

		assert.Equal(t, map[string]string{"validate": "required"},
			g.generateTagsMap(StructTagInfo{FieldName: "name"}))
		assert.Empty(t, g.generateTagsMap(StructTagInfo{FieldName: "name", IsOptional: true}))
	})
}

func TestOutputOptionsValidateStructTags(t *testing.T) {
	oo := OutputOptions{
		StructTags: StructTagsConfig{
			Tags: []StructTagTemplate{{Name: "json", Template: `{{.FieldName`}},
		},
	}
	problems := oo.Validate()
	require.Contains(t, problems, "struct-tags")
	assert.Contains(t, problems["struct-tags"], "invalid struct tag template")
}
