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

// HelperTemplate defines a template for a helper function that is conditionally included.
type HelperTemplate struct {
	Name     string   // Template name (e.g., "marshal_form")
	Imports  []Import // Required imports for this function
	Template string   // Template path in embedded FS (e.g., "helpers/marshal_form.go.tmpl")
}

// MarshalFormHelperTemplate is the template for the marshalForm helper function.
// This is included when any operation has a form-encoded typed request body.
var MarshalFormHelperTemplate = HelperTemplate{
	Name: "marshal_form",
	Imports: []Import{
		{Path: "errors"},
		{Path: "fmt"},
		{Path: "net/url"},
		{Path: "reflect"},
		{Path: "strings"},
	},
	Template: "helpers/marshal_form.go.tmpl",
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

// ReceiverTemplate defines a template for receiver (webhook/callback) generation.
type ReceiverTemplate struct {
	Name     string   // Template name (e.g., "receiver")
	Imports  []Import // Required imports for this template
	Template string   // Template path in embedded FS
}

// StdHTTPReceiverTemplates contains receiver templates for StdHTTP servers.
var StdHTTPReceiverTemplates = map[string]ReceiverTemplate{
	"receiver": {
		Name: "receiver",
		Imports: []Import{
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "net/http"},
			{Path: "net/url"},
		},
		Template: "server/stdhttp/receiver.go.tmpl",
	},
}

// ChiReceiverTemplates contains receiver templates for Chi servers.
var ChiReceiverTemplates = map[string]ReceiverTemplate{
	"receiver": {
		Name: "receiver",
		Imports: []Import{
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "net/http"},
			{Path: "net/url"},
		},
		Template: "server/chi/receiver.go.tmpl",
	},
}

// EchoReceiverTemplates contains receiver templates for Echo v5 servers.
var EchoReceiverTemplates = map[string]ReceiverTemplate{
	"receiver": {
		Name: "receiver",
		Imports: []Import{
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "net/http"},
			{Path: "net/url"},
			{Path: "github.com/labstack/echo/v5"},
		},
		Template: "server/echo/receiver.go.tmpl",
	},
}

// EchoV4ReceiverTemplates contains receiver templates for Echo v4 servers.
var EchoV4ReceiverTemplates = map[string]ReceiverTemplate{
	"receiver": {
		Name: "receiver",
		Imports: []Import{
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "net/http"},
			{Path: "net/url"},
			{Path: "github.com/labstack/echo/v4"},
		},
		Template: "server/echo-v4/receiver.go.tmpl",
	},
}

// GinReceiverTemplates contains receiver templates for Gin servers.
var GinReceiverTemplates = map[string]ReceiverTemplate{
	"receiver": {
		Name: "receiver",
		Imports: []Import{
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "net/http"},
			{Path: "net/url"},
			{Path: "github.com/gin-gonic/gin"},
		},
		Template: "server/gin/receiver.go.tmpl",
	},
}

// GorillaReceiverTemplates contains receiver templates for Gorilla servers.
var GorillaReceiverTemplates = map[string]ReceiverTemplate{
	"receiver": {
		Name: "receiver",
		Imports: []Import{
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "net/http"},
			{Path: "net/url"},
		},
		Template: "server/gorilla/receiver.go.tmpl",
	},
}

// FiberReceiverTemplates contains receiver templates for Fiber servers.
var FiberReceiverTemplates = map[string]ReceiverTemplate{
	"receiver": {
		Name: "receiver",
		Imports: []Import{
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "net/http"},
			{Path: "net/url"},
			{Path: "github.com/gofiber/fiber/v3"},
		},
		Template: "server/fiber/receiver.go.tmpl",
	},
}

// IrisReceiverTemplates contains receiver templates for Iris servers.
var IrisReceiverTemplates = map[string]ReceiverTemplate{
	"receiver": {
		Name: "receiver",
		Imports: []Import{
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "net/http"},
			{Path: "net/url"},
			{Path: "github.com/kataras/iris/v12"},
		},
		Template: "server/iris/receiver.go.tmpl",
	},
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
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "net/http"},
			{Path: "net/url"},
		},
		Template: "server/stdhttp/wrapper.go.tmpl",
	},
}

// ChiServerTemplates contains templates for Chi server generation.
var ChiServerTemplates = map[string]ServerTemplate{
	"interface": {
		Name: "interface",
		Imports: []Import{
			{Path: "net/http"},
		},
		Template: "server/chi/interface.go.tmpl",
	},
	"handler": {
		Name: "handler",
		Imports: []Import{
			{Path: "net/http"},
			{Path: "github.com/go-chi/chi/v5"},
		},
		Template: "server/chi/handler.go.tmpl",
	},
	"wrapper": {
		Name: "wrapper",
		Imports: []Import{
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "net/http"},
			{Path: "net/url"},
			{Path: "github.com/go-chi/chi/v5"},
		},
		Template: "server/chi/wrapper.go.tmpl",
	},
}

// EchoServerTemplates contains templates for Echo v5 server generation.
var EchoServerTemplates = map[string]ServerTemplate{
	"interface": {
		Name: "interface",
		Imports: []Import{
			{Path: "net/http"},
			{Path: "github.com/labstack/echo/v5"},
		},
		Template: "server/echo/interface.go.tmpl",
	},
	"handler": {
		Name: "handler",
		Imports: []Import{
			{Path: "github.com/labstack/echo/v5"},
		},
		Template: "server/echo/handler.go.tmpl",
	},
	"wrapper": {
		Name: "wrapper",
		Imports: []Import{
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "net/http"},
			{Path: "net/url"},
			{Path: "github.com/labstack/echo/v5"},
		},
		Template: "server/echo/wrapper.go.tmpl",
	},
}

// EchoV4ServerTemplates contains templates for Echo v4 server generation.
var EchoV4ServerTemplates = map[string]ServerTemplate{
	"interface": {
		Name: "interface",
		Imports: []Import{
			{Path: "net/http"},
			{Path: "github.com/labstack/echo/v4"},
		},
		Template: "server/echo-v4/interface.go.tmpl",
	},
	"handler": {
		Name: "handler",
		Imports: []Import{
			{Path: "github.com/labstack/echo/v4"},
		},
		Template: "server/echo-v4/handler.go.tmpl",
	},
	"wrapper": {
		Name: "wrapper",
		Imports: []Import{
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "net/http"},
			{Path: "net/url"},
			{Path: "github.com/labstack/echo/v4"},
		},
		Template: "server/echo-v4/wrapper.go.tmpl",
	},
}

// GinServerTemplates contains templates for Gin server generation.
var GinServerTemplates = map[string]ServerTemplate{
	"interface": {
		Name: "interface",
		Imports: []Import{
			{Path: "net/http"},
			{Path: "github.com/gin-gonic/gin"},
		},
		Template: "server/gin/interface.go.tmpl",
	},
	"handler": {
		Name: "handler",
		Imports: []Import{
			{Path: "github.com/gin-gonic/gin"},
		},
		Template: "server/gin/handler.go.tmpl",
	},
	"wrapper": {
		Name: "wrapper",
		Imports: []Import{
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "net/http"},
			{Path: "net/url"},
			{Path: "github.com/gin-gonic/gin"},
		},
		Template: "server/gin/wrapper.go.tmpl",
	},
}

// GorillaServerTemplates contains templates for Gorilla server generation.
var GorillaServerTemplates = map[string]ServerTemplate{
	"interface": {
		Name: "interface",
		Imports: []Import{
			{Path: "net/http"},
		},
		Template: "server/gorilla/interface.go.tmpl",
	},
	"handler": {
		Name: "handler",
		Imports: []Import{
			{Path: "net/http"},
			{Path: "github.com/gorilla/mux"},
		},
		Template: "server/gorilla/handler.go.tmpl",
	},
	"wrapper": {
		Name: "wrapper",
		Imports: []Import{
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "net/http"},
			{Path: "net/url"},
			{Path: "github.com/gorilla/mux"},
		},
		Template: "server/gorilla/wrapper.go.tmpl",
	},
}

// FiberServerTemplates contains templates for Fiber server generation.
var FiberServerTemplates = map[string]ServerTemplate{
	"interface": {
		Name: "interface",
		Imports: []Import{
			{Path: "github.com/gofiber/fiber/v3"},
		},
		Template: "server/fiber/interface.go.tmpl",
	},
	"handler": {
		Name: "handler",
		Imports: []Import{
			{Path: "github.com/gofiber/fiber/v3"},
		},
		Template: "server/fiber/handler.go.tmpl",
	},
	"wrapper": {
		Name: "wrapper",
		Imports: []Import{
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "net/http"},
			{Path: "net/url"},
			{Path: "github.com/gofiber/fiber/v3"},
		},
		Template: "server/fiber/wrapper.go.tmpl",
	},
}

// IrisServerTemplates contains templates for Iris server generation.
var IrisServerTemplates = map[string]ServerTemplate{
	"interface": {
		Name: "interface",
		Imports: []Import{
			{Path: "net/http"},
			{Path: "github.com/kataras/iris/v12"},
		},
		Template: "server/iris/interface.go.tmpl",
	},
	"handler": {
		Name: "handler",
		Imports: []Import{
			{Path: "github.com/kataras/iris/v12"},
		},
		Template: "server/iris/handler.go.tmpl",
	},
	"wrapper": {
		Name: "wrapper",
		Imports: []Import{
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "net/http"},
			{Path: "net/url"},
			{Path: "github.com/kataras/iris/v12"},
		},
		Template: "server/iris/wrapper.go.tmpl",
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

// InitiatorTemplate defines a template for initiator (webhook/callback sender) generation.
type InitiatorTemplate struct {
	Name     string   // Template name (e.g., "initiator_base", "initiator_interface")
	Imports  []Import // Required imports for this template
	Template string   // Template path in embedded FS
}

// InitiatorTemplates contains templates for initiator generation.
// These are shared between webhook and callback initiators (parameterized by prefix).
var InitiatorTemplates = map[string]InitiatorTemplate{
	"initiator_base": {
		Name: "initiator_base",
		Imports: []Import{
			{Path: "context"},
			{Path: "net/http"},
		},
		Template: "initiator/base.go.tmpl",
	},
	"initiator_interface": {
		Name: "initiator_interface",
		Imports: []Import{
			{Path: "context"},
			{Path: "io"},
			{Path: "net/http"},
		},
		Template: "initiator/interface.go.tmpl",
	},
	"initiator_methods": {
		Name: "initiator_methods",
		Imports: []Import{
			{Path: "context"},
			{Path: "io"},
			{Path: "net/http"},
		},
		Template: "initiator/methods.go.tmpl",
	},
	"initiator_request_builders": {
		Name: "initiator_request_builders",
		Imports: []Import{
			{Path: "bytes"},
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "io"},
			{Path: "net/http"},
			{Path: "net/url"},
			{Path: "strings"},
		},
		Template: "initiator/request_builders.go.tmpl",
	},
	"initiator_simple": {
		Name: "initiator_simple",
		Imports: []Import{
			{Path: "context"},
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "io"},
			{Path: "net/http"},
		},
		Template: "initiator/simple.go.tmpl",
	},
}

// ClientTemplate defines a template for client generation.
type ClientTemplate struct {
	Name     string   // Template name (e.g., "base", "interface")
	Imports  []Import // Required imports for this template
	Template string   // Template path in embedded FS
}

// ClientTemplates contains templates for client generation.
var ClientTemplates = map[string]ClientTemplate{
	"base": {
		Name: "base",
		Imports: []Import{
			{Path: "context"},
			{Path: "net/http"},
			{Path: "net/url"},
			{Path: "strings"},
		},
		Template: "client/base.go.tmpl",
	},
	"interface": {
		Name: "interface",
		Imports: []Import{
			{Path: "context"},
			{Path: "io"},
			{Path: "net/http"},
		},
		Template: "client/interface.go.tmpl",
	},
	"methods": {
		Name: "methods",
		Imports: []Import{
			{Path: "context"},
			{Path: "io"},
			{Path: "net/http"},
		},
		Template: "client/methods.go.tmpl",
	},
	"request_builders": {
		Name: "request_builders",
		Imports: []Import{
			{Path: "bytes"},
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "io"},
			{Path: "net/http"},
			{Path: "net/url"},
			{Path: "strings"},
		},
		Template: "client/request_builders.go.tmpl",
	},
	"simple": {
		Name: "simple",
		Imports: []Import{
			{Path: "context"},
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "io"},
			{Path: "net/http"},
		},
		Template: "client/simple.go.tmpl",
	},
}
