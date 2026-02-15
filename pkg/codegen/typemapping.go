package codegen

// SimpleTypeSpec defines the Go type for an OpenAPI type/format combination,
// along with any import required to use it.
type SimpleTypeSpec struct {
	Type   string `yaml:"type" json:"type"`
	Import string `yaml:"import,omitempty" json:"import,omitempty"`
}

// FormatMapping defines the default Go type and format-specific overrides
// for an OpenAPI type.
type FormatMapping struct {
	Default SimpleTypeSpec            `yaml:"default" json:"default"`
	Formats map[string]SimpleTypeSpec `yaml:"formats,omitempty" json:"formats,omitempty"`
}

// TypeMapping defines the mapping from OpenAPI types to Go types.
type TypeMapping struct {
	Integer FormatMapping `yaml:"integer,omitempty" json:"integer,omitempty"`
	Number  FormatMapping `yaml:"number,omitempty" json:"number,omitempty"`
	Boolean FormatMapping `yaml:"boolean,omitempty" json:"boolean,omitempty"`
	String  FormatMapping `yaml:"string,omitempty" json:"string,omitempty"`
}

// Merge returns a new TypeMapping with user overrides applied on top of base.
func (base TypeMapping) Merge(user TypeMapping) TypeMapping {
	return TypeMapping{
		Integer: base.Integer.merge(user.Integer),
		Number:  base.Number.merge(user.Number),
		Boolean: base.Boolean.merge(user.Boolean),
		String:  base.String.merge(user.String),
	}
}

func (base FormatMapping) merge(user FormatMapping) FormatMapping {
	result := FormatMapping{
		Default: base.Default,
		Formats: make(map[string]SimpleTypeSpec),
	}

	// Copy base formats
	for k, v := range base.Formats {
		result.Formats[k] = v
	}

	// Override with user default if specified
	if user.Default.Type != "" {
		result.Default = user.Default
	}

	// Override/add user formats
	for k, v := range user.Formats {
		result.Formats[k] = v
	}

	return result
}

// Resolve returns the SimpleTypeSpec for a given format string.
// If the format has a specific mapping, that is returned; otherwise the default is used.
func (fm FormatMapping) Resolve(format string) SimpleTypeSpec {
	if format != "" {
		if spec, ok := fm.Formats[format]; ok {
			return spec
		}
	}
	return fm.Default
}

// DefaultTypeMapping provides the default OpenAPI type/format to Go type mappings.
var DefaultTypeMapping = TypeMapping{
	Integer: FormatMapping{
		Default: SimpleTypeSpec{Type: "int"},
		Formats: map[string]SimpleTypeSpec{
			"int":    {Type: "int"},
			"int8":   {Type: "int8"},
			"int16":  {Type: "int16"},
			"int32":  {Type: "int32"},
			"int64":  {Type: "int64"},
			"uint":   {Type: "uint"},
			"uint8":  {Type: "uint8"},
			"uint16": {Type: "uint16"},
			"uint32": {Type: "uint32"},
			"uint64": {Type: "uint64"},
		},
	},
	Number: FormatMapping{
		Default: SimpleTypeSpec{Type: "float32"},
		Formats: map[string]SimpleTypeSpec{
			"float":  {Type: "float32"},
			"double": {Type: "float64"},
		},
	},
	Boolean: FormatMapping{
		Default: SimpleTypeSpec{Type: "bool"},
	},
	String: FormatMapping{
		Default: SimpleTypeSpec{Type: "string"},
		Formats: map[string]SimpleTypeSpec{
			"byte":      {Type: "[]byte"},
			"email":     {Type: "openapi_types.Email"},
			"date":      {Type: "openapi_types.Date"},
			"date-time": {Type: "time.Time", Import: "time"},
			"json":      {Type: "json.RawMessage", Import: "encoding/json"},
			"uuid":      {Type: "openapi_types.UUID"},
			"binary":    {Type: "openapi_types.File"},
		},
	},
}
