package codegen

type Configuration struct {
	// PackageName which will be used in all generated files
	PackageName string `yaml:"package"`
	// Output specifies the output file path
	Output string `yaml:"output"`
	// TypeMapping allows customizing OpenAPI type/format to Go type mappings
	TypeMapping TypeMapping `yaml:"type-mapping,omitempty"`
	// NameMangling configures how OpenAPI names are converted to Go identifiers
	NameMangling NameMangling `yaml:"name-mangling,omitempty"`
	// NameSubstitutions allows direct overrides of generated names
	NameSubstitutions NameSubstitutions `yaml:"name-substitutions,omitempty"`
}

// ApplyDefaults merges user configuration on top of default values.
func (c *Configuration) ApplyDefaults() {
	c.TypeMapping = DefaultTypeMapping.Merge(c.TypeMapping)
	c.NameMangling = DefaultNameMangling().Merge(c.NameMangling)
}
