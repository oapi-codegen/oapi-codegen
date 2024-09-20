// Copyright 2019 DeepMap, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package codegen

import (
	"bufio"
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
)

type ParameterDefinition struct {
	ParamName string // The original json parameter name, eg param_name
	In        string // Where the parameter is defined - path, header, cookie, query
	Required  bool   // Is this a required parameter?
	Spec      *openapi3.Parameter
	Schema    Schema
}

// TypeDef is here as an adapter after a large refactoring so that I don't
// have to update all the templates. It returns the type definition for a parameter,
// without the leading '*' for optional ones.
func (pd ParameterDefinition) TypeDef() string {
	typeDecl := pd.Schema.TypeDecl()
	return typeDecl
}

// JsonTag generates the JSON annotation to map GoType to json type name. If Parameter
// Foo is marshaled to json as "foo", this will create the annotation
// 'json:"foo"'
func (pd *ParameterDefinition) JsonTag() string {
	if pd.Required {
		return fmt.Sprintf("`json:\"%s\"`", pd.ParamName)
	} else {
		return fmt.Sprintf("`json:\"%s,omitempty\"`", pd.ParamName)
	}
}

func (pd *ParameterDefinition) IsJson() bool {
	p := pd.Spec
	if len(p.Content) == 1 {
		for k := range p.Content {
			if util.IsMediaTypeJson(k) {
				return true
			}
		}
	}
	return false
}

func (pd *ParameterDefinition) IsPassThrough() bool {
	p := pd.Spec
	if len(p.Content) > 1 {
		return true
	}
	if len(p.Content) == 1 {
		return !pd.IsJson()
	}
	return false
}

func (pd *ParameterDefinition) IsStyled() bool {
	p := pd.Spec
	return p.Schema != nil
}

func (pd *ParameterDefinition) Style() string {
	style := pd.Spec.Style
	if style == "" {
		in := pd.Spec.In
		switch in {
		case "path", "header":
			return "simple"
		case "query", "cookie":
			return "form"
		default:
			panic("unknown parameter format")
		}
	}
	return style
}

func (pd *ParameterDefinition) Explode() bool {
	if pd.Spec.Explode == nil {
		in := pd.Spec.In
		switch in {
		case "path", "header":
			return false
		case "query", "cookie":
			return true
		default:
			panic("unknown parameter format")
		}
	}
	return *pd.Spec.Explode
}

func (pd ParameterDefinition) GoVariableName() string {
	name := LowercaseFirstCharacters(pd.GoName())
	if IsGoKeyword(name) {
		name = "p" + UppercaseFirstCharacter(name)
	}
	if unicode.IsNumber([]rune(name)[0]) {
		name = "n" + name
	}
	return name
}

func (pd ParameterDefinition) GoName() string {
	goName := pd.ParamName
	if extension, ok := pd.Spec.Extensions[extGoName]; ok {
		if extGoFieldName, err := extParseGoFieldName(extension); err == nil {
			goName = extGoFieldName
		}
	}
	return SchemaNameToTypeName(goName)
}

func (pd ParameterDefinition) IndirectOptional() bool {
	return !pd.Required && !pd.Schema.SkipOptionalPointer
}

type ParameterDefinitions []ParameterDefinition

func (p ParameterDefinitions) FindByName(name string) *ParameterDefinition {
	for _, param := range p {
		if param.ParamName == name {
			return &param
		}
	}
	return nil
}

// DescribeParameters walks the given parameters dictionary, and generates the above
// descriptors into a flat list. This makes it a lot easier to traverse the
// data in the template engine.
func DescribeParameters(params openapi3.Parameters, path []string) ([]ParameterDefinition, error) {
	outParams := make([]ParameterDefinition, 0)
	for _, paramOrRef := range params {
		param := paramOrRef.Value

		goType, err := paramToGoType(param, append(path, param.Name))
		if err != nil {
			return nil, fmt.Errorf("error generating type for param (%s): %s",
				param.Name, err)
		}

		pd := ParameterDefinition{
			ParamName: param.Name,
			In:        param.In,
			Required:  param.Required,
			Spec:      param,
			Schema:    goType,
		}

		// If this is a reference to a predefined type, simply use the reference
		// name as the type. $ref: "#/components/schemas/custom_type" becomes
		// "CustomType".
		if IsGoTypeReference(paramOrRef.Ref) {
			goType, err := RefPathToGoType(paramOrRef.Ref)
			if err != nil {
				return nil, fmt.Errorf("error dereferencing (%s) for param (%s): %s",
					paramOrRef.Ref, param.Name, err)
			}
			pd.Schema.GoType = goType
		}
		outParams = append(outParams, pd)
	}
	return outParams, nil
}

type SecurityDefinition struct {
	ProviderName string
	Scopes       []string
}

func DescribeSecurityDefinition(securityRequirements openapi3.SecurityRequirements) []SecurityDefinition {
	outDefs := make([]SecurityDefinition, 0)

	for _, sr := range securityRequirements {
		for _, k := range SortedMapKeys(sr) {
			v := sr[k]
			outDefs = append(outDefs, SecurityDefinition{ProviderName: k, Scopes: v})
		}
	}

	return outDefs
}

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
						return nil, fmt.Errorf("Unable to determine Go type for %s.%s: %w", o.OperationId, contentTypeName, err)
					}

					var typeName string
					switch {

					// HAL+JSON:
					case StringInArray(contentTypeName, contentTypesHalJSON):
						typeName = fmt.Sprintf("HALJSON%s", nameNormalizer(responseName))
					case "application/json" == contentTypeName:
						// if it's the standard application/json
						typeName = fmt.Sprintf("JSON%s", nameNormalizer(responseName))
					// Vendored JSON
					case StringInArray(contentTypeName, contentTypesJSON) || util.IsMediaTypeJson(contentTypeName):
						baseTypeName := fmt.Sprintf("%s%s", nameNormalizer(contentTypeName), nameNormalizer(responseName))

						typeName = strings.ReplaceAll(baseTypeName, "Json", "JSON")
					// YAML:
					case StringInArray(contentTypeName, contentTypesYAML):
						typeName = fmt.Sprintf("YAML%s", nameNormalizer(responseName))
					// XML:
					case StringInArray(contentTypeName, contentTypesXML):
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

// FilterParameterDefinitionByType returns the subset of the specified parameters which are of the
// specified type.
func FilterParameterDefinitionByType(params []ParameterDefinition, in string) []ParameterDefinition {
	var out []ParameterDefinition
	for _, p := range params {
		if p.In == in {
			out = append(out, p)
		}
	}
	return out
}

// OperationDefinitions returns all operations for a swagger definition.
func OperationDefinitions(swagger *openapi3.T, initialismOverrides bool) ([]OperationDefinition, error) {
	var operations []OperationDefinition

	var toCamelCaseFunc func(string) string
	if initialismOverrides {
		toCamelCaseFunc = ToCamelCaseWithInitialism
	} else {
		toCamelCaseFunc = ToCamelCase
	}

	if swagger == nil || swagger.Paths == nil {
		return operations, nil
	}

	for _, requestPath := range SortedMapKeys(swagger.Paths.Map()) {
		pathItem := swagger.Paths.Value(requestPath)
		// These are parameters defined for all methods on a given path. They
		// are shared by all methods.
		globalParams, err := DescribeParameters(pathItem.Parameters, nil)
		if err != nil {
			return nil, fmt.Errorf("error describing global parameters for %s: %s",
				requestPath, err)
		}

		// Each path can have a number of operations, POST, GET, OPTIONS, etc.
		pathOps := pathItem.Operations()
		for _, opName := range SortedMapKeys(pathOps) {
			op := pathOps[opName]
			if pathItem.Servers != nil {
				op.Servers = &pathItem.Servers
			}
			// We rely on OperationID to generate function names, it's required
			if op.OperationID == "" {
				op.OperationID, err = generateDefaultOperationID(opName, requestPath, toCamelCaseFunc)
				if err != nil {
					return nil, fmt.Errorf("error generating default OperationID for %s/%s: %s",
						opName, requestPath, err)
				}
			} else {
				op.OperationID = nameNormalizer(op.OperationID)
			}
			op.OperationID = typeNamePrefix(op.OperationID) + op.OperationID

			// These are parameters defined for the specific path method that
			// we're iterating over.
			localParams, err := DescribeParameters(op.Parameters, []string{op.OperationID + "Params"})
			if err != nil {
				return nil, fmt.Errorf("error describing global parameters for %s/%s: %s",
					opName, requestPath, err)
			}
			// All the parameters required by a handler are the union of the
			// global parameters and the local parameters.
			allParams, err := CombineOperationParameters(globalParams, localParams)
			if err != nil {
				return nil, err
			}

			ensureExternalRefsInParameterDefinitions(&allParams, pathItem.Ref)

			// Order the path parameters to match the order as specified in
			// the path, not in the swagger spec, and validate that the parameter
			// names match, as downstream code depends on that.
			pathParams := FilterParameterDefinitionByType(allParams, "path")
			pathParams, err = SortParamsByPath(requestPath, pathParams)
			if err != nil {
				return nil, err
			}

			bodyDefinitions, typeDefinitions, err := GenerateBodyDefinitions(op.OperationID, op.RequestBody)
			if err != nil {
				return nil, fmt.Errorf("error generating body definitions: %w", err)
			}

			ensureExternalRefsInRequestBodyDefinitions(&bodyDefinitions, pathItem.Ref)

			responseDefinitions, err := GenerateResponseDefinitions(op.OperationID, op.Responses.Map())
			if err != nil {
				return nil, fmt.Errorf("error generating response definitions: %w", err)
			}

			ensureExternalRefsInResponseDefinitions(&responseDefinitions, pathItem.Ref)

			opDef := OperationDefinition{
				PathParams:   pathParams,
				HeaderParams: FilterParameterDefinitionByType(allParams, "header"),
				QueryParams:  FilterParameterDefinitionByType(allParams, "query"),
				CookieParams: FilterParameterDefinitionByType(allParams, "cookie"),
				OperationId:  nameNormalizer(op.OperationID),
				// Replace newlines in summary.
				Summary:         op.Summary,
				Method:          opName,
				Path:            requestPath,
				Spec:            op,
				Bodies:          bodyDefinitions,
				Responses:       responseDefinitions,
				TypeDefinitions: typeDefinitions,
			}

			// check for overrides of SecurityDefinitions.
			// See: "Step 2. Applying security:" from the spec:
			// https://swagger.io/docs/specification/authentication/
			if op.Security != nil {
				opDef.SecurityDefinitions = DescribeSecurityDefinition(*op.Security)
			} else {
				// use global securityDefinitions
				// globalSecurityDefinitions contains the top-level securityDefinitions.
				// They are the default securityPermissions which are injected into each
				// path, except for the case where a path explicitly overrides them.
				opDef.SecurityDefinitions = DescribeSecurityDefinition(swagger.Security)

			}

			if op.RequestBody != nil {
				opDef.BodyRequired = op.RequestBody.Value.Required
			}

			// Generate all the type definitions needed for this operation
			opDef.TypeDefinitions = append(opDef.TypeDefinitions, GenerateTypeDefsForOperation(opDef)...)

			operations = append(operations, opDef)
		}
	}
	return operations, nil
}

func generateDefaultOperationID(opName string, requestPath string, toCamelCaseFunc func(string) string) (string, error) {
	var operationId = strings.ToLower(opName)

	if opName == "" {
		return "", fmt.Errorf("operation name cannot be an empty string")
	}

	if requestPath == "" {
		return "", fmt.Errorf("request path cannot be an empty string")
	}

	for _, part := range strings.Split(requestPath, "/") {
		if part != "" {
			operationId = operationId + "-" + part
		}
	}

	return nameNormalizer(operationId), nil
}

// GenerateBodyDefinitions turns the Swagger body definitions into a list of our body
// definitions which will be used for code generation.
func GenerateBodyDefinitions(operationID string, bodyOrRef *openapi3.RequestBodyRef) ([]RequestBodyDefinition, []TypeDefinition, error) {
	if bodyOrRef == nil {
		return nil, nil, nil
	}
	body := bodyOrRef.Value

	var bodyDefinitions []RequestBodyDefinition
	var typeDefinitions []TypeDefinition

	for _, contentType := range SortedMapKeys(body.Content) {
		content := body.Content[contentType]
		var tag string
		var defaultBody bool

		switch {
		case contentType == "application/json":
			tag = "JSON"
			defaultBody = true
		case util.IsMediaTypeJson(contentType):
			tag = mediaTypeToCamelCase(contentType)
		case strings.HasPrefix(contentType, "multipart/"):
			tag = "Multipart"
		case contentType == "application/x-www-form-urlencoded":
			tag = "Formdata"
		case contentType == "text/plain":
			tag = "Text"
		default:
			bd := RequestBodyDefinition{
				Required:    body.Required,
				ContentType: contentType,
			}
			bodyDefinitions = append(bodyDefinitions, bd)
			continue
		}

		bodyTypeName := operationID + tag + "Body"
		bodySchema, err := GenerateGoSchema(content.Schema, []string{bodyTypeName})
		if err != nil {
			return nil, nil, fmt.Errorf("error generating request body definition: %w", err)
		}

		// If the body is a pre-defined type
		if content.Schema != nil && IsGoTypeReference(content.Schema.Ref) {
			// Convert the reference path to Go type
			refType, err := RefPathToGoType(content.Schema.Ref)
			if err != nil {
				return nil, nil, fmt.Errorf("error turning reference (%s) into a Go type: %w", content.Schema.Ref, err)
			}
			bodySchema.RefType = refType
		}

		// If the request has a body, but it's not a user defined
		// type under #/components, we'll define a type for it, so
		// that we have an easy to use type for marshaling.
		if bodySchema.RefType == "" {
			if contentType == "application/x-www-form-urlencoded" {
				// Apply the appropriate structure tag if the request
				// schema was defined under the operations' section.
				for i := range bodySchema.Properties {
					bodySchema.Properties[i].NeedsFormTag = true
				}

				// Regenerate the Golang struct adding the new form tag.
				bodySchema.GoType = GenStructFromSchema(bodySchema)
			}

			td := TypeDefinition{
				TypeName: bodyTypeName,
				Schema:   bodySchema,
			}
			typeDefinitions = append(typeDefinitions, td)
			// The body schema now is a reference to a type
			bodySchema.RefType = bodyTypeName
		}

		bd := RequestBodyDefinition{
			Required:    body.Required,
			Schema:      bodySchema,
			NameTag:     tag,
			ContentType: contentType,
			Default:     defaultBody,
		}

		if len(content.Encoding) != 0 {
			bd.Encoding = make(map[string]RequestBodyEncoding)
			for k, v := range content.Encoding {
				encoding := RequestBodyEncoding{ContentType: v.ContentType, Style: v.Style, Explode: v.Explode}
				bd.Encoding[k] = encoding
			}
		}

		bodyDefinitions = append(bodyDefinitions, bd)
	}
	sort.Slice(bodyDefinitions, func(i, j int) bool {
		return bodyDefinitions[i].ContentType < bodyDefinitions[j].ContentType
	})
	return bodyDefinitions, typeDefinitions, nil
}

func GenerateResponseDefinitions(operationID string, responses map[string]*openapi3.ResponseRef) ([]ResponseDefinition, error) {
	var responseDefinitions []ResponseDefinition
	// do not let multiple status codes ref to same response, it will break the type switch
	refSet := make(map[string]struct{})

	for _, statusCode := range SortedMapKeys(responses) {
		responseOrRef := responses[statusCode]
		if responseOrRef == nil {
			continue
		}
		response := responseOrRef.Value

		var responseContentDefinitions []ResponseContentDefinition

		for _, contentType := range SortedMapKeys(response.Content) {
			content := response.Content[contentType]
			var tag string
			switch {
			case contentType == "application/json":
				tag = "JSON"
			case util.IsMediaTypeJson(contentType):
				tag = mediaTypeToCamelCase(contentType)
			case contentType == "application/x-www-form-urlencoded":
				tag = "Formdata"
			case strings.HasPrefix(contentType, "multipart/"):
				tag = "Multipart"
			case contentType == "text/plain":
				tag = "Text"
			default:
				rcd := ResponseContentDefinition{
					ContentType: contentType,
				}
				responseContentDefinitions = append(responseContentDefinitions, rcd)
				continue
			}

			responseTypeName := operationID + statusCode + tag + "Response"
			contentSchema, err := GenerateGoSchema(content.Schema, []string{responseTypeName})
			if err != nil {
				return nil, fmt.Errorf("error generating request body definition: %w", err)
			}

			rcd := ResponseContentDefinition{
				ContentType: contentType,
				NameTag:     tag,
				Schema:      contentSchema,
			}

			responseContentDefinitions = append(responseContentDefinitions, rcd)
		}

		var responseHeaderDefinitions []ResponseHeaderDefinition
		for _, headerName := range SortedMapKeys(response.Headers) {
			header := response.Headers[headerName]
			contentSchema, err := GenerateGoSchema(header.Value.Schema, []string{})
			if err != nil {
				return nil, fmt.Errorf("error generating response header definition: %w", err)
			}
			headerDefinition := ResponseHeaderDefinition{Name: headerName, GoName: SchemaNameToTypeName(headerName), Schema: contentSchema}
			responseHeaderDefinitions = append(responseHeaderDefinitions, headerDefinition)
		}

		rd := ResponseDefinition{
			StatusCode: statusCode,
			Contents:   responseContentDefinitions,
			Headers:    responseHeaderDefinitions,
		}
		if response.Description != nil {
			rd.Description = *response.Description
		}
		if IsGoTypeReference(responseOrRef.Ref) {
			// Convert the reference path to Go type
			refType, err := RefPathToGoType(responseOrRef.Ref)
			if err != nil {
				return nil, fmt.Errorf("error turning reference (%s) into a Go type: %w", responseOrRef.Ref, err)
			}
			// Check if this ref is already used by another response definition. If not use the ref
			// If we let multiple response definitions alias to same response it will break the type switch
			// so only the first response will use the ref, other will generate new structs
			if _, ok := refSet[refType]; !ok {
				rd.Ref = refType
				refSet[refType] = struct{}{}
			}
		}
		responseDefinitions = append(responseDefinitions, rd)
	}

	return responseDefinitions, nil
}

func GenerateTypeDefsForOperation(op OperationDefinition) []TypeDefinition {
	var typeDefs []TypeDefinition
	// Start with the params object itself
	if len(op.Params()) != 0 {
		typeDefs = append(typeDefs, GenerateParamsTypes(op)...)
	}

	// Now, go through all the additional types we need to declare.
	for _, param := range op.AllParams() {
		typeDefs = append(typeDefs, param.Schema.AdditionalTypes...)
	}

	for _, body := range op.Bodies {
		typeDefs = append(typeDefs, body.Schema.AdditionalTypes...)
	}
	return typeDefs
}

// GenerateParamsTypes defines the schema for a parameters definition object
// which encapsulates all the query, header and cookie parameters for an operation.
func GenerateParamsTypes(op OperationDefinition) []TypeDefinition {
	var typeDefs []TypeDefinition

	objectParams := op.QueryParams
	objectParams = append(objectParams, op.HeaderParams...)
	objectParams = append(objectParams, op.CookieParams...)

	typeName := op.OperationId + "Params"

	s := Schema{}
	for _, param := range objectParams {
		pSchema := param.Schema
		param.Style()
		if pSchema.HasAdditionalProperties {
			propRefName := strings.Join([]string{typeName, param.GoName()}, "_")
			pSchema.RefType = propRefName
			typeDefs = append(typeDefs, TypeDefinition{
				TypeName: propRefName,
				Schema:   param.Schema,
			})
		}
		prop := Property{
			Description:   param.Spec.Description,
			JsonFieldName: param.ParamName,
			Required:      param.Required,
			Schema:        pSchema,
			NeedsFormTag:  param.Style() == "form",
			Extensions:    param.Spec.Extensions,
		}
		s.Properties = append(s.Properties, prop)
	}

	s.Description = op.Spec.Description
	s.GoType = GenStructFromSchema(s)

	td := TypeDefinition{
		TypeName: typeName,
		Schema:   s,
	}
	return append(typeDefs, td)
}

// GenerateTypesForOperations generates code for all types produced within operations
func GenerateTypesForOperations(t *template.Template, ops []OperationDefinition) (string, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	addTypes, err := GenerateTemplates([]string{"param-types.tmpl", "request-bodies.tmpl"}, t, ops)
	if err != nil {
		return "", fmt.Errorf("error generating type boilerplate for operations: %w", err)
	}
	if _, err := w.WriteString(addTypes); err != nil {
		return "", fmt.Errorf("error writing boilerplate to buffer: %w", err)
	}

	// Generate boiler plate for all additional types.
	var td []TypeDefinition
	for _, op := range ops {
		td = append(td, op.TypeDefinitions...)
	}

	addProps, err := GenerateAdditionalPropertyBoilerplate(t, td)
	if err != nil {
		return "", fmt.Errorf("error generating additional properties boilerplate for operations: %w", err)
	}

	if _, err := w.WriteString("\n"); err != nil {
		return "", fmt.Errorf("error generating additional properties boilerplate for operations: %w", err)
	}

	if _, err := w.WriteString(addProps); err != nil {
		return "", fmt.Errorf("error generating additional properties boilerplate for operations: %w", err)
	}

	if err = w.Flush(); err != nil {
		return "", fmt.Errorf("error flushing output buffer for server interface: %w", err)
	}

	return buf.String(), nil
}

// GenerateIrisServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateIrisServer(t *template.Template, operations []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"iris/iris-interface.tmpl", "iris/iris-middleware.tmpl", "iris/iris-handler.tmpl"}, t, operations)
}

// GenerateChiServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateChiServer(t *template.Template, operations []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"chi/chi-interface.tmpl", "chi/chi-middleware.tmpl", "chi/chi-handler.tmpl"}, t, operations)
}

// GenerateFiberServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateFiberServer(t *template.Template, operations []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"fiber/fiber-interface.tmpl", "fiber/fiber-middleware.tmpl", "fiber/fiber-handler.tmpl"}, t, operations)
}

// GenerateEchoServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateEchoServer(t *template.Template, operations []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"echo/echo-interface.tmpl", "echo/echo-wrappers.tmpl", "echo/echo-register.tmpl"}, t, operations)
}

// GenerateGinServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateGinServer(t *template.Template, operations []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"gin/gin-interface.tmpl", "gin/gin-wrappers.tmpl", "gin/gin-register.tmpl"}, t, operations)
}

// GenerateGorillaServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateGorillaServer(t *template.Template, operations []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"gorilla/gorilla-interface.tmpl", "gorilla/gorilla-middleware.tmpl", "gorilla/gorilla-register.tmpl"}, t, operations)
}

// GenerateStdHTTPServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateStdHTTPServer(t *template.Template, operations []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"stdhttp/std-http-interface.tmpl", "stdhttp/std-http-middleware.tmpl", "stdhttp/std-http-handler.tmpl"}, t, operations)
}

func GenerateStrictServer(t *template.Template, operations []OperationDefinition, opts Configuration) (string, error) {

	var templates []string

	if opts.Generate.ChiServer || opts.Generate.GorillaServer || opts.Generate.StdHTTPServer {
		templates = append(templates, "strict/strict-interface.tmpl", "strict/strict-http.tmpl")
	}
	if opts.Generate.EchoServer {
		templates = append(templates, "strict/strict-interface.tmpl", "strict/strict-echo.tmpl")
	}
	if opts.Generate.GinServer {
		templates = append(templates, "strict/strict-interface.tmpl", "strict/strict-gin.tmpl")
	}
	if opts.Generate.FiberServer {
		templates = append(templates, "strict/strict-fiber-interface.tmpl", "strict/strict-fiber.tmpl")
	}
	if opts.Generate.IrisServer {
		templates = append(templates, "strict/strict-iris-interface.tmpl", "strict/strict-iris.tmpl")
	}

	return GenerateTemplates(templates, t, operations)
}

func GenerateStrictResponses(t *template.Template, responses []ResponseDefinition) (string, error) {
	return GenerateTemplates([]string{"strict/strict-responses.tmpl"}, t, responses)
}

// GenerateClient uses the template engine to generate the function which registers our wrappers
// as Echo path handlers.
func GenerateClient(t *template.Template, ops []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"client.tmpl"}, t, ops)
}

// GenerateClientWithResponses generates a client which extends the basic client which does response
// unmarshaling.
func GenerateClientWithResponses(t *template.Template, ops []OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"client-with-responses.tmpl"}, t, ops)
}

// GenerateTemplates used to generate templates
func GenerateTemplates(templates []string, t *template.Template, ops interface{}) (string, error) {
	var generatedTemplates []string
	for _, tmpl := range templates {
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)

		if err := t.ExecuteTemplate(w, tmpl, ops); err != nil {
			return "", fmt.Errorf("error generating %s: %s", tmpl, err)
		}
		if err := w.Flush(); err != nil {
			return "", fmt.Errorf("error flushing output buffer for %s: %s", tmpl, err)
		}
		generatedTemplates = append(generatedTemplates, buf.String())
	}

	return strings.Join(generatedTemplates, "\n"), nil
}

// CombineOperationParameters combines the Parameters defined at a global level (Parameters defined for all methods on a given path) with the Parameters defined at a local level (Parameters defined for a specific path), preferring the locally defined parameter over the global one
func CombineOperationParameters(globalParams []ParameterDefinition, localParams []ParameterDefinition) ([]ParameterDefinition, error) {
	allParams := make([]ParameterDefinition, 0, len(globalParams)+len(localParams))
	dupCheck := make(map[string]map[string]string)
	for _, p := range localParams {
		if dupCheck[p.In] == nil {
			dupCheck[p.In] = make(map[string]string)
		}
		if _, exist := dupCheck[p.In][p.ParamName]; !exist {
			dupCheck[p.In][p.ParamName] = "local"
			allParams = append(allParams, p)
		} else {
			return nil, fmt.Errorf("duplicate local parameter %s/%s", p.In, p.ParamName)
		}
	}
	for _, p := range globalParams {
		if dupCheck[p.In] == nil {
			dupCheck[p.In] = make(map[string]string)
		}
		if t, exist := dupCheck[p.In][p.ParamName]; !exist {
			dupCheck[p.In][p.ParamName] = "global"
			allParams = append(allParams, p)
		} else if t == "global" {
			return nil, fmt.Errorf("duplicate global parameter %s/%s", p.In, p.ParamName)
		}
	}

	return allParams, nil
}
