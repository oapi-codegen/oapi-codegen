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

// ParamTemplate defines a template for a parameter styling/binding function.
type ParamTemplate struct {
	Name     string   // Function name (e.g., "StyleSimpleParam")
	Imports  []Import // Required imports for this function
	Template string   // Template name in embedded FS (e.g., "params/style_simple.go.tmpl")
}

// ParamHelpersTemplate is the template for shared helper functions.
// This is included whenever any param function is used.
var ParamHelpersTemplate = ParamTemplate{
	Name: "helpers",
	Imports: []Import{
		{Path: "bytes"},
		{Path: "encoding"},
		{Path: "encoding/json"},
		{Path: "errors"},
		{Path: "fmt"},
		{Path: "net/url"},
		{Path: "reflect"},
		{Path: "sort"},
		{Path: "strconv"},
		{Path: "strings"},
		{Path: "time"},
		{Path: "github.com/google/uuid"},
	},
	Template: "params/helpers.go.tmpl",
}

// ParamTemplates maps style/explode combinations to their template definitions.
// Keys follow the pattern: "style_{style}" or "style_{style}_explode" for styling,
// and "bind_{style}" or "bind_{style}_explode" for binding.
var ParamTemplates = map[string]ParamTemplate{
	// Style templates (serialization)
	"style_simple": {
		Name: "StyleSimpleParam",
		Imports: []Import{
			{Path: "bytes"},
			{Path: "encoding"},
			{Path: "encoding/json"},
			{Path: "errors"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
			{Path: "time"},
		},
		Template: "params/style_simple.go.tmpl",
	},
	"style_simple_explode": {
		Name: "StyleSimpleExplodeParam",
		Imports: []Import{
			{Path: "bytes"},
			{Path: "encoding"},
			{Path: "encoding/json"},
			{Path: "errors"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
			{Path: "time"},
		},
		Template: "params/style_simple_explode.go.tmpl",
	},
	"style_label": {
		Name: "StyleLabelParam",
		Imports: []Import{
			{Path: "bytes"},
			{Path: "encoding"},
			{Path: "encoding/json"},
			{Path: "errors"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
			{Path: "time"},
		},
		Template: "params/style_label.go.tmpl",
	},
	"style_label_explode": {
		Name: "StyleLabelExplodeParam",
		Imports: []Import{
			{Path: "bytes"},
			{Path: "encoding"},
			{Path: "encoding/json"},
			{Path: "errors"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
			{Path: "time"},
		},
		Template: "params/style_label_explode.go.tmpl",
	},
	"style_matrix": {
		Name: "StyleMatrixParam",
		Imports: []Import{
			{Path: "bytes"},
			{Path: "encoding"},
			{Path: "encoding/json"},
			{Path: "errors"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
			{Path: "time"},
		},
		Template: "params/style_matrix.go.tmpl",
	},
	"style_matrix_explode": {
		Name: "StyleMatrixExplodeParam",
		Imports: []Import{
			{Path: "bytes"},
			{Path: "encoding"},
			{Path: "encoding/json"},
			{Path: "errors"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
			{Path: "time"},
		},
		Template: "params/style_matrix_explode.go.tmpl",
	},
	"style_form": {
		Name: "StyleFormParam",
		Imports: []Import{
			{Path: "bytes"},
			{Path: "encoding"},
			{Path: "encoding/json"},
			{Path: "errors"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
			{Path: "time"},
		},
		Template: "params/style_form.go.tmpl",
	},
	"style_form_explode": {
		Name: "StyleFormExplodeParam",
		Imports: []Import{
			{Path: "bytes"},
			{Path: "encoding"},
			{Path: "encoding/json"},
			{Path: "errors"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
			{Path: "time"},
		},
		Template: "params/style_form_explode.go.tmpl",
	},
	"style_spaceDelimited": {
		Name: "StyleSpaceDelimitedParam",
		Imports: []Import{
			{Path: "encoding"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
			{Path: "time"},
		},
		Template: "params/style_space_delimited.go.tmpl",
	},
	"style_spaceDelimited_explode": {
		Name: "StyleSpaceDelimitedExplodeParam",
		Imports: []Import{
			{Path: "encoding"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
			{Path: "time"},
		},
		Template: "params/style_space_delimited_explode.go.tmpl",
	},
	"style_pipeDelimited": {
		Name: "StylePipeDelimitedParam",
		Imports: []Import{
			{Path: "encoding"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
			{Path: "time"},
		},
		Template: "params/style_pipe_delimited.go.tmpl",
	},
	"style_pipeDelimited_explode": {
		Name: "StylePipeDelimitedExplodeParam",
		Imports: []Import{
			{Path: "encoding"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
			{Path: "time"},
		},
		Template: "params/style_pipe_delimited_explode.go.tmpl",
	},
	"style_deepObject": {
		Name: "StyleDeepObjectParam",
		Imports: []Import{
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "sort"},
			{Path: "strconv"},
			{Path: "strings"},
		},
		Template: "params/style_deep_object.go.tmpl",
	},

	// Bind templates (deserialization)
	"bind_simple": {
		Name: "BindSimpleParam",
		Imports: []Import{
			{Path: "encoding"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
		},
		Template: "params/bind_simple.go.tmpl",
	},
	"bind_simple_explode": {
		Name: "BindSimpleExplodeParam",
		Imports: []Import{
			{Path: "encoding"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
		},
		Template: "params/bind_simple_explode.go.tmpl",
	},
	"bind_label": {
		Name: "BindLabelParam",
		Imports: []Import{
			{Path: "encoding"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
		},
		Template: "params/bind_label.go.tmpl",
	},
	"bind_label_explode": {
		Name: "BindLabelExplodeParam",
		Imports: []Import{
			{Path: "encoding"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
		},
		Template: "params/bind_label_explode.go.tmpl",
	},
	"bind_matrix": {
		Name: "BindMatrixParam",
		Imports: []Import{
			{Path: "encoding"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
		},
		Template: "params/bind_matrix.go.tmpl",
	},
	"bind_matrix_explode": {
		Name: "BindMatrixExplodeParam",
		Imports: []Import{
			{Path: "encoding"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
		},
		Template: "params/bind_matrix_explode.go.tmpl",
	},
	"bind_form": {
		Name: "BindFormParam",
		Imports: []Import{
			{Path: "encoding"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
		},
		Template: "params/bind_form.go.tmpl",
	},
	"bind_form_explode": {
		Name: "BindFormExplodeParam",
		Imports: []Import{
			{Path: "fmt"},
			{Path: "net/url"},
			{Path: "reflect"},
			{Path: "strings"},
			{Path: "time"},
		},
		Template: "params/bind_form_explode.go.tmpl",
	},
	"bind_spaceDelimited": {
		Name: "BindSpaceDelimitedParam",
		Imports: []Import{
			{Path: "encoding"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
		},
		Template: "params/bind_space_delimited.go.tmpl",
	},
	"bind_spaceDelimited_explode": {
		Name: "BindSpaceDelimitedExplodeParam",
		Imports: []Import{
			{Path: "net/url"},
		},
		Template: "params/bind_space_delimited_explode.go.tmpl",
	},
	"bind_pipeDelimited": {
		Name: "BindPipeDelimitedParam",
		Imports: []Import{
			{Path: "encoding"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "strings"},
		},
		Template: "params/bind_pipe_delimited.go.tmpl",
	},
	"bind_pipeDelimited_explode": {
		Name: "BindPipeDelimitedExplodeParam",
		Imports: []Import{
			{Path: "net/url"},
		},
		Template: "params/bind_pipe_delimited_explode.go.tmpl",
	},
	"bind_deepObject": {
		Name: "BindDeepObjectParam",
		Imports: []Import{
			{Path: "errors"},
			{Path: "fmt"},
			{Path: "net/url"},
			{Path: "reflect"},
			{Path: "sort"},
			{Path: "strconv"},
			{Path: "strings"},
			{Path: "time"},
		},
		Template: "params/bind_deep_object.go.tmpl",
	},
}

// ParamStyleKey returns the registry key for a style/explode combination.
// The prefix should be "style_" for serialization or "bind_" for binding.
func ParamStyleKey(prefix, style string, explode bool) string {
	key := prefix + style
	if explode {
		key += "_explode"
	}
	return key
}

// ServerTemplate defines a template for server generation.
type ServerTemplate struct {
	Name     string   // Template name (e.g., "interface", "handler")
	Imports  []Import // Required imports for this template
	Template string   // Template path in embedded FS
}

// StdHTTPServerTemplates contains templates for StdHTTP server generation.
var StdHTTPServerTemplates = map[string]ServerTemplate{
	"interface": {
		Name: "interface",
		Imports: []Import{
			{Path: "net/http"},
		},
		Template: "server/stdhttp/interface.go.tmpl",
	},
	"handler": {
		Name: "handler",
		Imports: []Import{
			{Path: "net/http"},
		},
		Template: "server/stdhttp/handler.go.tmpl",
	},
	"wrapper": {
		Name: "wrapper",
		Imports: []Import{
			{Path: "context"},
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "net/http"},
			{Path: "net/url"},
		},
		Template: "server/stdhttp/wrapper.go.tmpl",
	},
}

// SharedServerTemplates contains templates shared across all server implementations.
var SharedServerTemplates = map[string]ServerTemplate{
	"errors": {
		Name: "errors",
		Imports: []Import{
			{Path: "fmt"},
		},
		Template: "server/errors.go.tmpl",
	},
	"param_types": {
		Name: "param_types",
		Imports: []Import{},
		Template: "server/param_types.go.tmpl",
	},
}
