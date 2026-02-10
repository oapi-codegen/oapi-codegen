package codegen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
)

// SchemaPath represents the document location of a schema, e.g.
// ["components", "schemas", "Pet", "properties", "name"].
type SchemaPath []string

// String returns the path joined with "/".
func (sp SchemaPath) String() string {
	return strings.Join(sp, "/")
}

// SchemaContext identifies where in the OpenAPI document a schema was found.
type SchemaContext int

const (
	ContextComponentSchema SchemaContext = iota
	ContextComponentParameter
	ContextComponentRequestBody
	ContextComponentResponse
	ContextComponentHeader
	ContextOperationParameter
	ContextOperationRequestBody
	ContextOperationResponse
	ContextClientResponseWrapper
)

// String returns a human-readable name for the context.
func (sc SchemaContext) String() string {
	switch sc {
	case ContextComponentSchema:
		return "Schema"
	case ContextComponentParameter:
		return "Parameter"
	case ContextComponentRequestBody:
		return "RequestBody"
	case ContextComponentResponse:
		return "Response"
	case ContextComponentHeader:
		return "Header"
	case ContextOperationParameter:
		return "OperationParameter"
	case ContextOperationRequestBody:
		return "OperationRequestBody"
	case ContextOperationResponse:
		return "OperationResponse"
	case ContextClientResponseWrapper:
		return "ClientResponseWrapper"
	default:
		return "Unknown"
	}
}

// Suffix returns the suffix to use for collision resolution.
func (sc SchemaContext) Suffix() string {
	switch sc {
	case ContextComponentSchema:
		return "Schema"
	case ContextComponentParameter, ContextOperationParameter:
		return "Parameter"
	case ContextComponentRequestBody, ContextOperationRequestBody:
		return "RequestBody"
	case ContextComponentResponse, ContextOperationResponse:
		return "Response"
	case ContextComponentHeader:
		return "Header"
	case ContextClientResponseWrapper:
		return "Response"
	default:
		return ""
	}
}

// GatheredSchema represents a schema discovered during the gather pass,
// along with its document location and context metadata.
type GatheredSchema struct {
	Path        SchemaPath
	Context     SchemaContext
	Ref         string           // $ref string if this is a reference
	Schema      *openapi3.Schema // The resolved schema value
	OperationID string           // Enclosing operation's ID, if any
	ContentType string           // Media type, if from request/response body
	StatusCode  string           // HTTP status code, if from a response
	ParamIndex  int              // Parameter index within an operation
	ComponentName string         // The component name (e.g., "Bar" for components/schemas/Bar)
}

// IsComponentSchema returns true if this schema came from components/schemas.
func (gs *GatheredSchema) IsComponentSchema() bool {
	return gs.Context == ContextComponentSchema
}

// GatherSchemas walks the entire OpenAPI spec and collects all schemas that
// will need Go type names. This is the first pass of the multi-pass resolution.
func GatherSchemas(spec *openapi3.T, opts Configuration) []*GatheredSchema {
	var schemas []*GatheredSchema

	if spec.Components != nil {
		schemas = append(schemas, gatherComponentSchemas(spec.Components)...)
		schemas = append(schemas, gatherComponentParameters(spec.Components)...)
		schemas = append(schemas, gatherComponentResponses(spec.Components)...)
		schemas = append(schemas, gatherComponentRequestBodies(spec.Components)...)
		schemas = append(schemas, gatherComponentHeaders(spec.Components)...)
	}

	// Gather client response wrapper types for operations that will generate
	// client code. These synthetic entries exist so wrapper types like
	// `CreateChatCompletionResponse` participate in collision detection.
	if opts.Generate.Client {
		schemas = append(schemas, gatherClientResponseWrappers(spec)...)
	}

	return schemas
}

func gatherComponentSchemas(components *openapi3.Components) []*GatheredSchema {
	var result []*GatheredSchema
	for _, name := range SortedSchemaKeys(components.Schemas) {
		schemaRef := components.Schemas[name]
		if schemaRef == nil || schemaRef.Value == nil {
			continue
		}
		result = append(result, &GatheredSchema{
			Path:          SchemaPath{"components", "schemas", name},
			Context:       ContextComponentSchema,
			Ref:           schemaRef.Ref,
			Schema:        schemaRef.Value,
			ComponentName: name,
		})
	}
	return result
}

func gatherComponentParameters(components *openapi3.Components) []*GatheredSchema {
	var result []*GatheredSchema
	for _, name := range SortedMapKeys(components.Parameters) {
		paramRef := components.Parameters[name]
		if paramRef == nil || paramRef.Value == nil {
			continue
		}
		param := paramRef.Value
		if param.Schema != nil && param.Schema.Value != nil {
			result = append(result, &GatheredSchema{
				Path:          SchemaPath{"components", "parameters", name},
				Context:       ContextComponentParameter,
				Ref:           paramRef.Ref,
				Schema:        param.Schema.Value,
				ComponentName: name,
			})
		}
	}
	return result
}

func gatherComponentResponses(components *openapi3.Components) []*GatheredSchema {
	var result []*GatheredSchema
	for _, name := range SortedMapKeys(components.Responses) {
		responseRef := components.Responses[name]
		if responseRef == nil || responseRef.Value == nil {
			continue
		}
		response := responseRef.Value
		for _, mediaType := range SortedMapKeys(response.Content) {
			if !util.IsMediaTypeJson(mediaType) {
				continue
			}
			mt := response.Content[mediaType]
			if mt.Schema != nil && mt.Schema.Value != nil {
				result = append(result, &GatheredSchema{
					Path:          SchemaPath{"components", "responses", name, "content", mediaType},
					Context:       ContextComponentResponse,
					Ref:           responseRef.Ref,
					Schema:        mt.Schema.Value,
					ContentType:   mediaType,
					ComponentName: name,
				})
			}
		}
	}
	return result
}

func gatherComponentRequestBodies(components *openapi3.Components) []*GatheredSchema {
	var result []*GatheredSchema
	for _, name := range SortedMapKeys(components.RequestBodies) {
		bodyRef := components.RequestBodies[name]
		if bodyRef == nil || bodyRef.Value == nil {
			continue
		}
		body := bodyRef.Value
		for _, mediaType := range SortedMapKeys(body.Content) {
			if !util.IsMediaTypeJson(mediaType) {
				continue
			}
			mt := body.Content[mediaType]
			if mt.Schema != nil && mt.Schema.Value != nil {
				result = append(result, &GatheredSchema{
					Path:          SchemaPath{"components", "requestBodies", name, "content", mediaType},
					Context:       ContextComponentRequestBody,
					Ref:           bodyRef.Ref,
					Schema:        mt.Schema.Value,
					ContentType:   mediaType,
					ComponentName: name,
				})
			}
		}
	}
	return result
}

func gatherComponentHeaders(components *openapi3.Components) []*GatheredSchema {
	var result []*GatheredSchema
	for _, name := range SortedMapKeys(components.Headers) {
		headerRef := components.Headers[name]
		if headerRef == nil || headerRef.Value == nil {
			continue
		}
		header := headerRef.Value
		if header.Schema != nil && header.Schema.Value != nil {
			result = append(result, &GatheredSchema{
				Path:          SchemaPath{"components", "headers", name},
				Context:       ContextComponentHeader,
				Ref:           headerRef.Ref,
				Schema:        header.Schema.Value,
				ComponentName: name,
			})
		}
	}
	return result
}

// gatherClientResponseWrappers creates synthetic schema entries for each
// operation that would generate a client response wrapper type like
// `<OperationId>Response`. These don't correspond to a real schema in the
// spec but they need names that don't collide with real types.
func gatherClientResponseWrappers(spec *openapi3.T) []*GatheredSchema {
	var result []*GatheredSchema

	if spec.Paths == nil {
		return result
	}

	// Collect all operations sorted for determinism
	type opEntry struct {
		path   string
		method string
		op     *openapi3.Operation
	}
	var ops []opEntry

	pathKeys := SortedMapKeys(spec.Paths.Map())
	for _, path := range pathKeys {
		pathItem := spec.Paths.Find(path)
		if pathItem == nil {
			continue
		}
		for method, op := range pathItem.Operations() {
			if op != nil && op.OperationID != "" {
				ops = append(ops, opEntry{path: path, method: method, op: op})
			}
		}
	}

	// Sort by operationID for determinism
	sort.Slice(ops, func(i, j int) bool {
		return ops[i].op.OperationID < ops[j].op.OperationID
	})

	for _, entry := range ops {
		result = append(result, &GatheredSchema{
			Path:        SchemaPath{"paths", entry.path, entry.method, "x-client-response-wrapper"},
			Context:     ContextClientResponseWrapper,
			OperationID: entry.op.OperationID,
		})
	}

	return result
}

// gatherOperationID returns a normalized operation ID for naming purposes.
func gatherOperationID(op *openapi3.Operation) string {
	if op == nil || op.OperationID == "" {
		return ""
	}
	return op.OperationID
}

// FormatPath returns a human-readable representation of the path for debugging.
func (gs *GatheredSchema) FormatPath() string {
	return fmt.Sprintf("#/%s", strings.Join(gs.Path, "/"))
}
