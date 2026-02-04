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

// IsComponentSchema returns true if this schema is defined in #/components/schemas
func (d *SchemaDescriptor) IsComponentSchema() bool {
	return len(d.Path) >= 2 && d.Path[0] == "components" && d.Path[1] == "schemas"
}
