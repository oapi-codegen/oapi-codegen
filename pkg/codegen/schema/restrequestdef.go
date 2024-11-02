package schema

import (
	"fmt"
	"strings"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
)

// RequestBodyDefinition describes a request body
type RequestBodyDefinition struct {
	// Is this body required, or optional?
	Required bool

	// This is the schema describing this body
	Schema Schema

	// When we generate type names, we need a Tag for it, such as JSON, in
	// which case we will produce "JSONBody".
	NameTag string

	// This is the content type corresponding to the body, eg, application/json
	ContentType string

	// Whether this is the default body type. For an operation named OpFoo, we
	// will not add suffixes like OpFooJSONBody for this one.
	Default bool

	// Contains encoding options for formdata
	Encoding map[string]RequestBodyEncoding
}

// TypeDef returns the Go type definition for a request body
func (r RequestBodyDefinition) TypeDef(opID string) *TypeDefinition {
	return &TypeDefinition{
		TypeName: fmt.Sprintf("%s%sRequestBody", opID, r.NameTag),
		Schema:   r.Schema,
	}
}

// CustomType returns whether the body is a custom inline type, or pre-defined. This is
// poorly named, but it's here for compatibility reasons post-refactoring
// TODO: clean up the templates code, it can be simpler.
func (r RequestBodyDefinition) CustomType() bool {
	return r.Schema.RefType == ""
}

// When we're generating multiple functions which relate to request bodies,
// this generates the suffix. Such as Operation DoFoo would be suffixed with
// DoFooWithXMLBody.
func (r RequestBodyDefinition) Suffix() string {
	// The default response is never suffixed.
	if r.Default {
		return ""
	}
	return "With" + r.NameTag + "Body"
}

// IsSupportedByClient returns true if we support this content type for client. Otherwise only generic method will ge generated
func (r RequestBodyDefinition) IsSupportedByClient() bool {
	return r.IsJSON() || r.NameTag == "Formdata" || r.NameTag == "Text"
}

// IsJSON returns whether this is a JSON media type, for instance:
// - application/json
// - application/vnd.api+json
// - application/*+json
func (r RequestBodyDefinition) IsJSON() bool {
	return util.IsMediaTypeJson(r.ContentType)
}

// IsSupported returns true if we support this content type for server. Otherwise io.Reader will be generated
func (r RequestBodyDefinition) IsSupported() bool {
	return r.NameTag != ""
}

// IsFixedContentType returns true if content type has fixed content type, i.e. contains no "*" symbol
func (r RequestBodyDefinition) IsFixedContentType() bool {
	return !strings.Contains(r.ContentType, "*")
}

type RequestBodyEncoding struct {
	ContentType string
	Style       string
	Explode     *bool
}
