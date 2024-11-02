package schema

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen/constants"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
)

// OperationDefinition describes an Operation
type OperationDefinition struct {
	OperationId string // The operation_id description from Swagger, used to generate function names

	PathParams          []ParameterDefinition // Parameters in the path, eg, /path/:param
	HeaderParams        []ParameterDefinition // Parameters in HTTP headers
	QueryParams         []ParameterDefinition // Parameters in the query, /path?param
	CookieParams        []ParameterDefinition // Parameters in cookies
	TypeDefinitions     []TypeDefinition      // These are all the types we need to define for this operation
	SecurityDefinitions []SecurityDefinition  // These are the security providers
	BodyRequired        bool
	Bodies              []RequestBodyDefinition // The list of bodies for which to generate handlers.
	Responses           []ResponseDefinition    // The list of responses that can be accepted by handlers.
	Summary             string                  // Summary string from Swagger, used to generate a comment
	Method              string                  // GET, POST, DELETE, etc.
	Path                string                  // The Swagger path for the operation, like /resource/{id}
	Spec                *openapi3.Operation
}

// Params returns the list of all parameters except Path parameters. Path parameters
// are handled differently from the rest, since they're mandatory.
func (o *OperationDefinition) Params() []ParameterDefinition {
	result := append(o.QueryParams, o.HeaderParams...)
	result = append(result, o.CookieParams...)
	return result
}

// AllParams returns all parameters
func (o *OperationDefinition) AllParams() []ParameterDefinition {
	result := append(o.QueryParams, o.HeaderParams...)
	result = append(result, o.CookieParams...)
	result = append(result, o.PathParams...)
	return result
}

// If we have parameters other than path parameters, they're bundled into an
// object. Returns true if we have any of those. This is used from the template
// engine.
func (o *OperationDefinition) RequiresParamObject() bool {
	return len(o.Params()) > 0
}

// HasBody is called by the template engine to determine whether to generate body
// marshaling code on the client. This is true for all body types, whether
// we generate types for them.
func (o *OperationDefinition) HasBody() bool {
	return o.Spec.RequestBody != nil
}

// SummaryAsComment returns the Operations summary as a multi line comment
func (o *OperationDefinition) SummaryAsComment() string {
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

// GetResponseTypeDefinitions produces a list of type definitions for a given Operation for the response
// types which we know how to parse. These will be turned into fields on a
// response object for automatic deserialization of responses in the generated
// Client code. See "client-with-responses.tmpl".
func (o *OperationDefinition) GetResponseTypeDefinitions() ([]ResponseTypeDefinition, error) {
	var tds []ResponseTypeDefinition

	if o.Spec == nil || o.Spec.Responses == nil {
		return tds, nil
	}

	sortedResponsesKeys := SortedMapKeys(o.Spec.Responses.Map())
	for _, responseName := range sortedResponsesKeys {
		responseRef := o.Spec.Responses.Value(responseName)

		// We can only generate a type if we have a value:
		if responseRef.Value != nil {
			jsonCount := 0
			for mediaType := range responseRef.Value.Content {
				if util.IsMediaTypeJson(mediaType) {
					jsonCount++
				}
			}

			sortedContentKeys := SortedMapKeys(responseRef.Value.Content)
			for _, contentTypeName := range sortedContentKeys {
				contentType := responseRef.Value.Content[contentTypeName]
				// We can only generate a type if we have a schema:
				if contentType.Schema != nil {
					responseSchema, err := GenerateGoSchema(contentType.Schema, []string{o.OperationId, responseName})
					if err != nil {
						return nil, fmt.Errorf("unable to determine go type for %s.%s: %w", o.OperationId, contentTypeName, err)
					}

					var typeName string
					switch {

					// HAL+JSON:
					case StringInArray(contentTypeName, constants.ContentTypesHalJSON):
						typeName = fmt.Sprintf("HALJSON%s", nameNormalizer(responseName))
					case contentTypeName == "application/json":
						// if it's the standard application/json
						typeName = fmt.Sprintf("JSON%s", nameNormalizer(responseName))
					// Vendored JSON
					case StringInArray(contentTypeName, constants.ContentTypesJSON) || util.IsMediaTypeJson(contentTypeName):
						baseTypeName := fmt.Sprintf("%s%s", nameNormalizer(contentTypeName), nameNormalizer(responseName))

						typeName = strings.ReplaceAll(baseTypeName, "Json", "JSON")
					// YAML:
					case StringInArray(contentTypeName, constants.ContentTypesYAML):
						typeName = fmt.Sprintf("YAML%s", nameNormalizer(responseName))
					// XML:
					case StringInArray(contentTypeName, constants.ContentTypesXML):
						typeName = fmt.Sprintf("XML%s", nameNormalizer(responseName))
					default:
						continue
					}

					td := ResponseTypeDefinition{
						TypeDefinition: TypeDefinition{
							TypeName: typeName,
							Schema:   responseSchema,
						},
						ResponseName:              responseName,
						ContentTypeName:           contentTypeName,
						AdditionalTypeDefinitions: responseSchema.GetAdditionalTypeDefs(),
					}
					if IsGoTypeReference(responseRef.Ref) {
						refType, err := RefPathToGoType(responseRef.Ref)
						if err != nil {
							return nil, fmt.Errorf("error dereferencing response Ref: %w", err)
						}
						if jsonCount > 1 && util.IsMediaTypeJson(contentTypeName) {
							refType += mediaTypeToCamelCase(contentTypeName)
						}
						td.Schema.RefType = refType
					}
					tds = append(tds, td)
				}
			}
		}
	}
	return tds, nil
}

func (o OperationDefinition) HasMaskedRequestContentTypes() bool {
	for _, body := range o.Bodies {
		if !body.IsFixedContentType() {
			return true
		}
	}
	return false
}
