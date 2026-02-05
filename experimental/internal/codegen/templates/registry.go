package templates

// Import represents a Go import with optional alias.
type Import struct {
	Path  string
	Alias string // empty if no alias
}

// TypeTemplate defines a template for a custom type along with its required imports.
type TypeTemplate struct {
	Name     string   // Type name (e.g., "Email", "Date")
	Imports  []Import // Required imports for this type
	Template string   // Template name in embedded FS (e.g., "types/email.tmpl")
}

// TypeTemplates maps type names to their template definitions.
var TypeTemplates = map[string]TypeTemplate{
	"Email": {
		Name: "Email",
		Imports: []Import{
			{Path: "encoding/json"},
			{Path: "errors"},
			{Path: "regexp"},
		},
		Template: "types/email.tmpl",
	},
	"Date": {
		Name: "Date",
		Imports: []Import{
			{Path: "encoding/json"},
			{Path: "time"},
		},
		Template: "types/date.tmpl",
	},
	"UUID": {
		Name: "UUID",
		Imports: []Import{
			{Path: "github.com/google/uuid"},
		},
		Template: "types/uuid.tmpl",
	},
	"File": {
		Name: "File",
		Imports: []Import{
			{Path: "bytes"},
			{Path: "encoding/json"},
			{Path: "io"},
			{Path: "mime/multipart"},
		},
		Template: "types/file.tmpl",
	},
	"Nullable": {
		Name: "Nullable",
		Imports: []Import{
			{Path: "encoding/json"},
			{Path: "errors"},
		},
		Template: "types/nullable.tmpl",
	},
}
