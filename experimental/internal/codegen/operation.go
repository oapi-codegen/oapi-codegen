package codegen

import (
	"strings"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// OperationSource indicates where an operation was defined in the spec.
type OperationSource string

const (
	OperationSourcePath     OperationSource = "path"
	OperationSourceWebhook  OperationSource = "webhook"
	OperationSourceCallback OperationSource = "callback"
)

// OperationDescriptor describes a single API operation from an OpenAPI spec.
type OperationDescriptor struct {
	OperationID   string // Normalized operation ID for function names
	GoOperationID string // Go-safe identifier (handles leading digits, keywords)
	Method        string // HTTP method: GET, POST, PUT, DELETE, etc.
	Path          string // Original path: /users/{id}
	Summary       string // For generating comments
	Description   string // Longer description

	// Source indicates where this operation was defined (path, webhook, or callback)
	Source       OperationSource
	WebhookName  string // Webhook name (for Source=webhook)
	CallbackName string // Callback key (for Source=callback)
	ParentOpID   string // Parent operation ID (for Source=callback)

	PathParams   []*ParameterDescriptor
	QueryParams  []*ParameterDescriptor
	HeaderParams []*ParameterDescriptor
	CookieParams []*ParameterDescriptor

	Bodies    []*RequestBodyDescriptor
	Responses []*ResponseDescriptor

	Security []SecurityRequirement

	// Precomputed for templates
	HasBody        bool   // Has at least one request body
	HasParams      bool   // Has non-path params (needs Params struct)
	ParamsTypeName string // "{OperationID}Params"

	// Reference to the underlying spec
	Spec *v3.Operation
}

// Params returns all non-path parameters (query, header, cookie).
// These are bundled into a Params struct.
func (o *OperationDescriptor) Params() []*ParameterDescriptor {
	result := make([]*ParameterDescriptor, 0, len(o.QueryParams)+len(o.HeaderParams)+len(o.CookieParams))
	result = append(result, o.QueryParams...)
	result = append(result, o.HeaderParams...)
	result = append(result, o.CookieParams...)
	return result
}

// AllParams returns all parameters including path params.
func (o *OperationDescriptor) AllParams() []*ParameterDescriptor {
	result := make([]*ParameterDescriptor, 0, len(o.PathParams)+len(o.QueryParams)+len(o.HeaderParams)+len(o.CookieParams))
	result = append(result, o.PathParams...)
	result = append(result, o.QueryParams...)
	result = append(result, o.HeaderParams...)
	result = append(result, o.CookieParams...)
	return result
}

// SummaryAsComment returns the summary formatted as a Go comment.
func (o *OperationDescriptor) SummaryAsComment() string {
	if o.Summary == "" {
		return ""
	}
	trimmed := strings.TrimSuffix(o.Summary, "\n")
	parts := strings.Split(trimmed, "\n")
	for i, p := range parts {
		parts[i] = "// " + p
	}
	return strings.Join(parts, "\n")
}

// DefaultBody returns the default request body (typically application/json), or nil.
func (o *OperationDescriptor) DefaultBody() *RequestBodyDescriptor {
	for _, b := range o.Bodies {
		if b.IsDefault {
			return b
		}
	}
	if len(o.Bodies) > 0 {
		return o.Bodies[0]
	}
	return nil
}

// DefaultTypedBody returns the first request body with GenerateTyped=true,
// preferring the default body. Returns nil if no typed body exists.
func (o *OperationDescriptor) DefaultTypedBody() *RequestBodyDescriptor {
	for _, b := range o.Bodies {
		if b.GenerateTyped && b.IsDefault {
			return b
		}
	}
	for _, b := range o.Bodies {
		if b.GenerateTyped {
			return b
		}
	}
	return nil
}

// HasTypedBody returns true if at least one request body has GenerateTyped=true.
func (o *OperationDescriptor) HasTypedBody() bool {
	return o.DefaultTypedBody() != nil
}

// ParameterDescriptor describes a parameter in any location.
type ParameterDescriptor struct {
	Name     string // Original name from spec (e.g., "user_id")
	GoName   string // Go-safe name for struct fields (e.g., "UserId")
	Location string // "path", "query", "header", "cookie"
	Required bool

	// Serialization style
	Style   string // "simple", "form", "label", "matrix", etc.
	Explode bool

	// Type information
	Schema   *SchemaDescriptor
	TypeDecl string // Go type declaration (e.g., "string", "[]int", "*MyType")

	// Precomputed function names for templates
	StyleFunc string // "StyleSimpleParam", "StyleFormExplodeParam", etc.
	BindFunc  string // "BindSimpleParam", "BindFormExplodeParam", etc.

	// Encoding modes
	IsStyled      bool // Uses style/explode serialization (most common)
	IsPassThrough bool // No styling, just pass the string through
	IsJSON        bool // Parameter uses JSON content encoding

	Spec *v3.Parameter
}

// GoVariableName returns a Go-safe variable name for this parameter.
// Used for local variables in generated code.
func (p *ParameterDescriptor) GoVariableName() string {
	name := LowercaseFirstCharacter(p.GoName)
	if IsGoKeyword(name) {
		name = "p" + p.GoName
	}
	// Handle leading digits
	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = "n" + name
	}
	return name
}

// HasOptionalPointer returns true if this parameter should be a pointer
// (optional parameters that aren't required).
func (p *ParameterDescriptor) HasOptionalPointer() bool {
	if p.Required {
		return false
	}
	// Check if schema has skip-optional-pointer extension
	if p.Schema != nil && p.Schema.Extensions != nil &&
		p.Schema.Extensions.SkipOptionalPointer != nil && *p.Schema.Extensions.SkipOptionalPointer {
		return false
	}
	return true
}

// RequestBodyDescriptor describes a request body for a specific content type.
type RequestBodyDescriptor struct {
	ContentType string // "application/json", "multipart/form-data", etc.
	Required    bool
	Schema      *SchemaDescriptor

	// Precomputed for templates
	NameTag    string // "JSON", "Formdata", "Multipart", "Text", etc.
	GoTypeName string // "{OperationID}JSONBody", etc.
	FuncSuffix string // "", "WithJSONBody", "WithFormBody" (empty for default)
	IsDefault     bool // Is this the default body type?
	IsFormEncoded bool // Is this application/x-www-form-urlencoded?
	GenerateTyped bool // Generate typed methods for this body (based on content-types config)

	// Encoding options for form data
	Encoding map[string]RequestBodyEncoding
}

// RequestBodyEncoding describes encoding options for a form field.
type RequestBodyEncoding struct {
	ContentType string
	Style       string
	Explode     *bool
}

// ResponseDescriptor describes a response for a status code.
type ResponseDescriptor struct {
	StatusCode  string // "200", "404", "default", "2XX"
	Description string
	Contents    []*ResponseContentDescriptor
	Headers     []*ResponseHeaderDescriptor
	Ref         string // If this is a reference to a named response
}

// GoName returns a Go-safe name for this response (e.g., "200" -> "200", "default" -> "Default").
func (r *ResponseDescriptor) GoName() string {
	return ToCamelCase(r.StatusCode)
}

// HasFixedStatusCode returns true if the status code is a specific number (not "default" or "2XX").
func (r *ResponseDescriptor) HasFixedStatusCode() bool {
	if r.StatusCode == "default" {
		return false
	}
	// Check for wildcard patterns like "2XX"
	if strings.HasSuffix(strings.ToUpper(r.StatusCode), "XX") {
		return false
	}
	return true
}

// ResponseContentDescriptor describes response content for a content type.
type ResponseContentDescriptor struct {
	ContentType string
	Schema      *SchemaDescriptor
	NameTag     string // "JSON", "XML", etc.
	IsJSON      bool
}

// ResponseHeaderDescriptor describes a response header.
type ResponseHeaderDescriptor struct {
	Name     string
	GoName   string
	Required bool
	Schema   *SchemaDescriptor
}

// SecurityRequirement describes a security requirement for an operation.
type SecurityRequirement struct {
	Name   string   // Security scheme name
	Scopes []string // Required scopes (for OAuth2)
}

// Helper functions for computing descriptor fields

// ComputeStyleFunc returns the style function name for a parameter.
func ComputeStyleFunc(style string, explode bool) string {
	base := "Style" + ToCamelCase(style)
	if explode {
		return base + "ExplodeParam"
	}
	return base + "Param"
}

// ComputeBindFunc returns the bind function name for a parameter.
func ComputeBindFunc(style string, explode bool) string {
	base := "Bind" + ToCamelCase(style)
	if explode {
		return base + "ExplodeParam"
	}
	return base + "Param"
}

// ComputeBodyNameTag returns the name tag for a content type.
func ComputeBodyNameTag(contentType string) string {
	switch {
	case contentType == "application/json":
		return "JSON"
	case IsMediaTypeJSON(contentType):
		return MediaTypeToCamelCase(contentType)
	case strings.HasPrefix(contentType, "multipart/"):
		return "Multipart"
	case contentType == "application/x-www-form-urlencoded":
		return "Formdata"
	case contentType == "text/plain":
		return "Text"
	case strings.HasPrefix(contentType, "application/xml") || strings.HasSuffix(contentType, "+xml"):
		return "XML"
	default:
		return ""
	}
}

// IsMediaTypeJSON returns true if the content type is a JSON media type.
func IsMediaTypeJSON(contentType string) bool {
	if contentType == "application/json" {
		return true
	}
	if strings.HasSuffix(contentType, "+json") {
		return true
	}
	if strings.Contains(contentType, "json") {
		return true
	}
	return false
}

// MediaTypeToCamelCase converts a media type to a CamelCase identifier.
func MediaTypeToCamelCase(mediaType string) string {
	// application/vnd.api+json -> ApplicationVndApiJson
	mediaType = strings.ReplaceAll(mediaType, "/", " ")
	mediaType = strings.ReplaceAll(mediaType, "+", " ")
	mediaType = strings.ReplaceAll(mediaType, ".", " ")
	mediaType = strings.ReplaceAll(mediaType, "-", " ")
	return ToCamelCase(mediaType)
}
