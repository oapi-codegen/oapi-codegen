package codegen

import (
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
)

// SchemaPath represents the location of a schema in the OpenAPI document.
// Used for deriving type names and disambiguating collisions.
// Example: ["components", "schemas", "Pet", "properties", "address"]
type SchemaPath []string

// String returns the path as a JSON pointer-style string.
func (p SchemaPath) String() string {
	return "#/" + strings.Join(p, "/")
}

// Append returns a new SchemaPath with the given elements appended.
// This creates a fresh slice to avoid aliasing issues with append.
func (p SchemaPath) Append(elements ...string) SchemaPath {
	result := make(SchemaPath, len(p)+len(elements))
	copy(result, p)
	copy(result[len(p):], elements)
	return result
}

// ContainsProperties returns true if this path contains "properties" anywhere.
// This indicates it's an inline property schema rather than a component schema.
func (p SchemaPath) ContainsProperties() bool {
	for _, element := range p {
		if element == "properties" {
			return true
		}
	}
	return false
}

// SchemaDescriptor represents a schema found during the first pass through the spec.
type SchemaDescriptor struct {
	// Path is where this schema appears in the document
	Path SchemaPath

	// Ref is the $ref string if this is a reference (e.g., "#/components/schemas/Pet")
	// Empty if this is an inline schema definition
	Ref string

	// Schema is the underlying schema from libopenapi
	// nil for unresolved external references
	Schema *base.Schema

	// Parent points to the containing schema (nil for top-level schemas)
	Parent *SchemaDescriptor

	// StableName is the deterministic Go type name derived from the full path.
	// This name is stable across spec changes and should be used for type definitions.
	// Example: #/components/schemas/Cat -> CatSchemaComponent
	StableName string

	// ShortName is a friendly alias that may change due to deduplication.
	// Generated as a type alias pointing to StableName.
	ShortName string

	// OperationID is the operationId from the path operation, if this schema
	// comes from a path's request body or response. Used for friendlier naming.
	OperationID string

	// ContentType is the media type (e.g., "application/json") if this schema
	// comes from a request body or response content. Used for naming.
	ContentType string

	// Extensions holds parsed x- extension values for this schema.
	// These control code generation behavior (type overrides, field names, etc.)
	Extensions *Extensions

	// Recursive structure:
	Properties      map[string]*SchemaDescriptor
	Items           *SchemaDescriptor
	AllOf           []*SchemaDescriptor
	AnyOf           []*SchemaDescriptor
	OneOf           []*SchemaDescriptor
	AdditionalProps *SchemaDescriptor
}

// IsReference returns true if this schema is a $ref to another schema
func (d *SchemaDescriptor) IsReference() bool {
	return d.Ref != ""
}

// IsExternalReference returns true if this is a reference to an external file.
// External refs have the format: file.yaml#/path/to/schema
func (d *SchemaDescriptor) IsExternalReference() bool {
	if d.Ref == "" {
		return false
	}
	// External refs contain # but don't start with it
	return !strings.HasPrefix(d.Ref, "#") && strings.Contains(d.Ref, "#")
}

// ParseExternalRef splits an external reference into its file path and internal path.
// For "common/api.yaml#/components/schemas/Pet", returns ("common/api.yaml", "#/components/schemas/Pet").
// Returns empty strings if not an external ref.
func (d *SchemaDescriptor) ParseExternalRef() (filePath, internalPath string) {
	if !d.IsExternalReference() {
		return "", ""
	}
	parts := strings.SplitN(d.Ref, "#", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], "#" + parts[1]
}

// IsComponentSchema returns true if this schema is defined in #/components/schemas
func (d *SchemaDescriptor) IsComponentSchema() bool {
	return len(d.Path) >= 2 && d.Path[0] == "components" && d.Path[1] == "schemas"
}

// IsTopLevelComponentSchema returns true if this schema is a direct child of #/components/schemas
// (i.e., #/components/schemas/Foo, not #/components/schemas/Foo/properties/bar).
func (d *SchemaDescriptor) IsTopLevelComponentSchema() bool {
	return len(d.Path) == 3 && d.Path[0] == "components" && d.Path[1] == "schemas"
}
