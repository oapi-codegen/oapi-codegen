package schema

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
)

type ResponseDefinition struct {
	StatusCode  string
	Description string
	Contents    []ResponseContentDefinition
	Headers     []ResponseHeaderDefinition
	Ref         string
}

func (r ResponseDefinition) HasFixedStatusCode() bool {
	_, err := strconv.Atoi(r.StatusCode)
	return err == nil
}

func (r ResponseDefinition) GoName() string {
	return SchemaNameToTypeName(r.StatusCode)
}

func (r ResponseDefinition) IsRef() bool {
	return r.Ref != ""
}

func (r ResponseDefinition) IsExternalRef() bool {
	if !r.IsRef() {
		return false
	}
	return strings.Contains(r.Ref, ".")
}

type ResponseContentDefinition struct {
	// This is the schema describing this content
	Schema Schema

	// This is the content type corresponding to the body, eg, application/json
	ContentType string

	// When we generate type names, we need a Tag for it, such as JSON, in
	// which case we will produce "Response200JSONContent".
	NameTag string
}

// TypeDef returns the Go type definition for a request body
func (r ResponseContentDefinition) TypeDef(opID string, statusCode int) *TypeDefinition {
	return &TypeDefinition{
		TypeName: fmt.Sprintf("%s%v%sResponse", opID, statusCode, r.NameTagOrContentType()),
		Schema:   r.Schema,
	}
}

func (r ResponseContentDefinition) IsSupported() bool {
	return r.NameTag != ""
}

// HasFixedContentType returns true if content type has fixed content type, i.e. contains no "*" symbol
func (r ResponseContentDefinition) HasFixedContentType() bool {
	return !strings.Contains(r.ContentType, "*")
}

func (r ResponseContentDefinition) NameTagOrContentType() string {
	if r.NameTag != "" {
		return r.NameTag
	}
	return SchemaNameToTypeName(r.ContentType)
}

// IsJSON returns whether this is a JSON media type, for instance:
// - application/json
// - application/vnd.api+json
// - application/*+json
func (r ResponseContentDefinition) IsJSON() bool {
	return util.IsMediaTypeJson(r.ContentType)
}

type ResponseHeaderDefinition struct {
	Name   string
	GoName string
	Schema Schema
}
