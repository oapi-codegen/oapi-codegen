package codegen

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
)

// StructTagInfo is the data made available to struct tag templates. The
// omit flags are fully computed before rendering: OmitEmpty folds in the
// required/readOnly/writeOnly logic, compatibility options and the
// x-omitempty extension; OmitZero folds in x-omitzero and
// prefer-skip-optional-pointer-with-omitzero. Templates therefore only
// need to decide whether to emit them, not re-derive them.
type StructTagInfo struct {
	// FieldName is the property name from the OpenAPI spec.
	FieldName string
	// IsOptional is true when the field is not listed as required.
	IsOptional bool
	// OmitEmpty is true when the field should carry ",omitempty".
	OmitEmpty bool
	// OmitZero is true when the field should carry ",omitzero".
	OmitZero bool
	// NeedsFormTag is true for fields bound from form-style parameters or
	// urlencoded request bodies.
	NeedsFormTag bool
}

// StructTagTemplate defines a single struct tag as a name plus a Go
// text/template rendered against StructTagInfo. A template that renders
// to the empty string suppresses the tag on that field.
type StructTagTemplate struct {
	// Name is the tag name (e.g. "json", "yaml", "db").
	Name string `yaml:"name"`
	// Template is a Go text/template producing the tag value.
	// Available fields: .FieldName, .IsOptional, .OmitEmpty, .OmitZero,
	// .NeedsFormTag.
	Template string `yaml:"template"`
}

// StructTagsConfig configures struct tag generation for generated fields.
type StructTagsConfig struct {
	// Tags lists the tags to generate. Entries are merged by name on top
	// of the defaults: a matching name replaces the default template, a
	// new name adds a tag. Output ordering is always alphabetical by name.
	Tags []StructTagTemplate `yaml:"tags,omitempty"`
}

const (
	defaultJSONTagTemplate = `{{.FieldName}}{{if .OmitEmpty}},omitempty{{end}}{{if .OmitZero}},omitzero{{end}}`
	defaultFormTagTemplate = `{{if .NeedsFormTag}}{{.FieldName}}{{if .OmitEmpty}},omitempty{{end}}{{end}}`
	defaultYamlTagTemplate = `{{.FieldName}}{{if .OmitEmpty}},omitempty{{end}}`
)

// defaultStructTagsConfig returns the built-in tag templates, which
// reproduce the historical hardcoded output. The yaml entry is only
// present when the legacy `output-options.yaml-tags` flag is enabled and
// applies only where it always did (schema properties); a user-supplied
// yaml entry in struct-tags supersedes the flag entirely.
func defaultStructTagsConfig(enableYamlTags bool) StructTagsConfig {
	tags := []StructTagTemplate{
		{Name: "json", Template: defaultJSONTagTemplate},
		{Name: "form", Template: defaultFormTagTemplate},
	}
	if enableYamlTags {
		tags = append(tags, StructTagTemplate{Name: "yaml", Template: defaultYamlTagTemplate})
	}
	return StructTagsConfig{Tags: tags}
}

// Merge merges user config on top of this config by tag name: matching
// names override the default template, new names are appended.
func (c StructTagsConfig) Merge(other StructTagsConfig) StructTagsConfig {
	if len(other.Tags) == 0 {
		return c
	}
	merged := make(map[string]StructTagTemplate, len(c.Tags)+len(other.Tags))
	order := make([]string, 0, len(c.Tags)+len(other.Tags))
	for _, t := range c.Tags {
		merged[t.Name] = t
		order = append(order, t.Name)
	}
	for _, t := range other.Tags {
		if _, exists := merged[t.Name]; !exists {
			order = append(order, t.Name)
		}
		merged[t.Name] = t
	}
	result := StructTagsConfig{Tags: make([]StructTagTemplate, 0, len(order))}
	for _, name := range order {
		result.Tags = append(result.Tags, merged[name])
	}
	return result
}

// structTagGenerator renders struct tag values from parsed templates.
type structTagGenerator struct {
	templates []tagTemplate
}

type tagTemplate struct {
	name string
	tmpl *template.Template
}

// newStructTagGenerator parses the configured templates. Each template is
// also rendered against every combination of StructTagInfo's boolean
// fields so that execute-time errors (e.g. references to unknown fields),
// including ones gated behind conditionals, are reported at configuration
// time rather than silently dropping tags during generation.
func newStructTagGenerator(config StructTagsConfig) (*structTagGenerator, error) {
	g := &structTagGenerator{
		templates: make([]tagTemplate, 0, len(config.Tags)),
	}
	// Keep numFlags in sync with the boolean fields of StructTagInfo.
	const numFlags = 4
	probes := make([]StructTagInfo, 0, 1<<numFlags)
	for mask := 0; mask < 1<<numFlags; mask++ {
		probes = append(probes, StructTagInfo{
			FieldName:    "probe",
			IsOptional:   mask&1 != 0,
			OmitEmpty:    mask&2 != 0,
			OmitZero:     mask&4 != 0,
			NeedsFormTag: mask&8 != 0,
		})
	}
	for _, tag := range config.Tags {
		tmpl, err := template.New(tag.Name).Parse(tag.Template)
		if err != nil {
			return nil, fmt.Errorf("invalid struct tag template for %q: %w", tag.Name, err)
		}
		for _, probe := range probes {
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, probe); err != nil {
				return nil, fmt.Errorf("struct tag template for %q failed to render: %w", tag.Name, err)
			}
		}
		g.templates = append(g.templates, tagTemplate{name: tag.Name, tmpl: tmpl})
	}
	return g, nil
}

// generateTagsMap renders every configured template for the given field
// and returns the non-empty results as tag name -> tag value. Extension
// driven overrides (x-go-json-ignore, x-oapi-codegen-extra-tags) are
// applied by the callers on top of this map.
func (g *structTagGenerator) generateTagsMap(info StructTagInfo) map[string]string {
	result := make(map[string]string, len(g.templates))
	for _, t := range g.templates {
		var buf bytes.Buffer
		if err := t.tmpl.Execute(&buf, info); err != nil {
			// Templates are validated at construction time against every
			// combination of boolean fields, so a failure here is
			// unexpected. The tag is skipped rather than emitting a partial
			// value, with a warning so the omission is not silent.
			fmt.Fprintf(os.Stderr, "Warning: struct tag template %q failed for field %q, omitting the tag: %v\n", t.name, info.FieldName, err)
			continue
		}
		if value := buf.String(); value != "" {
			result[t.name] = value
		}
	}
	return result
}

// schemaFieldTagGenerator returns the generator used for schema property
// fields, building it from the current configuration on first use. This
// lazy path keeps direct callers of GenFieldsFromProperties (tests)
// working without running Generate; Generate itself constructs the
// generators eagerly so configuration errors are reported.
func schemaFieldTagGenerator() *structTagGenerator {
	if globalState.schemaFieldTagGenerator == nil {
		cfg := defaultStructTagsConfig(globalState.options.OutputOptions.EnableYamlTags).
			Merge(globalState.options.OutputOptions.StructTags)
		g, err := newStructTagGenerator(cfg)
		if err != nil {
			// Generate() reports this error to the user; fall back to the
			// defaults so lazy callers still produce standard tags.
			g, _ = newStructTagGenerator(defaultStructTagsConfig(globalState.options.OutputOptions.EnableYamlTags))
		}
		globalState.schemaFieldTagGenerator = g
	}
	return globalState.schemaFieldTagGenerator
}

// paramFieldTagGenerator returns the generator used for path parameter
// fields on strict RequestObject structs (ParameterDefinition.JsonTag).
// It differs from the schema one in a single way: the legacy `yaml-tags`
// flag never applied to those fields, so its injected default is excluded
// here. Tags configured explicitly via struct-tags apply to them too.
func paramFieldTagGenerator() *structTagGenerator {
	if globalState.paramFieldTagGenerator == nil {
		cfg := defaultStructTagsConfig(false).
			Merge(globalState.options.OutputOptions.StructTags)
		g, err := newStructTagGenerator(cfg)
		if err != nil {
			g, _ = newStructTagGenerator(defaultStructTagsConfig(false))
		}
		globalState.paramFieldTagGenerator = g
	}
	return globalState.paramFieldTagGenerator
}
