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
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen/schema"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
)

// DescribeParameters walks the given parameters dictionary, and generates the above
// descriptors into a flat list. This makes it a lot easier to traverse the
// data in the template engine.
func DescribeParameters(params openapi3.Parameters, path []string) ([]schema.ParameterDefinition, error) {
	outParams := make([]schema.ParameterDefinition, 0)
	for _, paramOrRef := range params {
		param := paramOrRef.Value

		goType, err := schema.ParamToGoType(param, append(path, param.Name))
		if err != nil {
			return nil, fmt.Errorf("error generating type for param (%s): %s",
				param.Name, err)
		}

		pd := schema.ParameterDefinition{
			ParamName: param.Name,
			In:        param.In,
			Required:  param.Required,
			Spec:      param,
			Schema:    goType,
		}

		// If this is a reference to a predefined type, simply use the reference
		// name as the type. $ref: "#/components/schemas/custom_type" becomes
		// "CustomType".
		if schema.IsGoTypeReference(paramOrRef.Ref) {
			goType, err := schema.RefPathToGoType(paramOrRef.Ref)
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

func DescribeSecurityDefinition(securityRequirements openapi3.SecurityRequirements) []schema.SecurityDefinition {
	outDefs := make([]schema.SecurityDefinition, 0)

	for _, sr := range securityRequirements {
		for _, k := range schema.SortedMapKeys(sr) {
			v := sr[k]
			outDefs = append(outDefs, schema.SecurityDefinition{ProviderName: k, Scopes: v})
		}
	}

	return outDefs
}

// FilterParameterDefinitionByType returns the subset of the specified parameters which are of the
// specified type.
func FilterParameterDefinitionByType(params []schema.ParameterDefinition, in string) []schema.ParameterDefinition {
	var out []schema.ParameterDefinition
	for _, p := range params {
		if p.In == in {
			out = append(out, p)
		}
	}
	return out
}

// OperationDefinitions returns all operations for a swagger definition.
func OperationDefinitions(swagger *openapi3.T, initialismOverrides bool) ([]schema.OperationDefinition, error) {
	var operations []schema.OperationDefinition

	var toCamelCaseFunc func(string) string
	if initialismOverrides {
		toCamelCaseFunc = schema.ToCamelCaseWithInitialism
	} else {
		toCamelCaseFunc = schema.ToCamelCase
	}

	if swagger == nil || swagger.Paths == nil {
		return operations, nil
	}

	for _, requestPath := range schema.SortedMapKeys(swagger.Paths.Map()) {
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
		for _, opName := range schema.SortedMapKeys(pathOps) {
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
				op.OperationID = schema.DefaultNameNormalizer(op.OperationID)
			}
			op.OperationID = schema.TypeNamePrefix(op.OperationID) + op.OperationID

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

			schema.EnsureExternalRefsInParameterDefinitions(&allParams, pathItem.Ref)

			// Order the path parameters to match the order as specified in
			// the path, not in the swagger spec, and validate that the parameter
			// names match, as downstream code depends on that.
			pathParams := FilterParameterDefinitionByType(allParams, "path")
			pathParams, err = schema.SortParamsByPath(requestPath, pathParams)
			if err != nil {
				return nil, err
			}

			bodyDefinitions, typeDefinitions, err := GenerateBodyDefinitions(op.OperationID, op.RequestBody)
			if err != nil {
				return nil, fmt.Errorf("error generating body definitions: %w", err)
			}

			schema.EnsureExternalRefsInRequestBodyDefinitions(&bodyDefinitions, pathItem.Ref)

			responseDefinitions, err := GenerateResponseDefinitions(op.OperationID, op.Responses.Map())
			if err != nil {
				return nil, fmt.Errorf("error generating response definitions: %w", err)
			}

			schema.EnsureExternalRefsInResponseDefinitions(&responseDefinitions, pathItem.Ref)

			opDef := schema.OperationDefinition{
				PathParams:   pathParams,
				HeaderParams: FilterParameterDefinitionByType(allParams, "header"),
				QueryParams:  FilterParameterDefinitionByType(allParams, "query"),
				CookieParams: FilterParameterDefinitionByType(allParams, "cookie"),
				OperationId:  schema.DefaultNameNormalizer(op.OperationID),
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

	return schema.DefaultNameNormalizer(operationId), nil
}

// GenerateBodyDefinitions turns the Swagger body definitions into a list of our body
// definitions which will be used for code generation.
func GenerateBodyDefinitions(operationID string, bodyOrRef *openapi3.RequestBodyRef) ([]schema.RequestBodyDefinition, []schema.TypeDefinition, error) {
	if bodyOrRef == nil {
		return nil, nil, nil
	}
	body := bodyOrRef.Value

	var bodyDefinitions []schema.RequestBodyDefinition
	var typeDefinitions []schema.TypeDefinition

	for _, contentType := range schema.SortedMapKeys(body.Content) {
		content := body.Content[contentType]
		var tag string
		var defaultBody bool

		switch {
		case contentType == "application/json":
			tag = "JSON"
			defaultBody = true
		case util.IsMediaTypeJson(contentType):
			tag = schema.MediaTypeToCamelCase(contentType)
		case strings.HasPrefix(contentType, "multipart/"):
			tag = "Multipart"
		case contentType == "application/x-www-form-urlencoded":
			tag = "Formdata"
		case contentType == "text/plain":
			tag = "Text"
		default:
			bd := schema.RequestBodyDefinition{
				Required:    body.Required,
				ContentType: contentType,
			}
			bodyDefinitions = append(bodyDefinitions, bd)
			continue
		}

		bodyTypeName := operationID + tag + "Body"
		bodySchema, err := schema.GenerateGoSchema(content.Schema, []string{bodyTypeName})
		if err != nil {
			return nil, nil, fmt.Errorf("error generating request body definition: %w", err)
		}

		// If the body is a pre-defined type
		if content.Schema != nil && schema.IsGoTypeReference(content.Schema.Ref) {
			// Convert the reference path to Go type
			refType, err := schema.RefPathToGoType(content.Schema.Ref)
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
				bodySchema.GoType = schema.GenStructFromSchema(bodySchema)
			}

			td := schema.TypeDefinition{
				TypeName: bodyTypeName,
				Schema:   bodySchema,
			}
			typeDefinitions = append(typeDefinitions, td)
			// The body schema now is a reference to a type
			bodySchema.RefType = bodyTypeName
		}

		bd := schema.RequestBodyDefinition{
			Required:    body.Required,
			Schema:      bodySchema,
			NameTag:     tag,
			ContentType: contentType,
			Default:     defaultBody,
		}

		if len(content.Encoding) != 0 {
			bd.Encoding = make(map[string]schema.RequestBodyEncoding)
			for k, v := range content.Encoding {
				encoding := schema.RequestBodyEncoding{ContentType: v.ContentType, Style: v.Style, Explode: v.Explode}
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

func GenerateResponseDefinitions(operationID string, responses map[string]*openapi3.ResponseRef) ([]schema.ResponseDefinition, error) {
	var responseDefinitions []schema.ResponseDefinition
	// do not let multiple status codes ref to same response, it will break the type switch
	refSet := make(map[string]struct{})

	for _, statusCode := range schema.SortedMapKeys(responses) {
		responseOrRef := responses[statusCode]
		if responseOrRef == nil {
			continue
		}
		response := responseOrRef.Value

		var responseContentDefinitions []schema.ResponseContentDefinition

		for _, contentType := range schema.SortedMapKeys(response.Content) {
			content := response.Content[contentType]
			var tag string
			switch {
			case contentType == "application/json":
				tag = "JSON"
			case util.IsMediaTypeJson(contentType):
				tag = schema.MediaTypeToCamelCase(contentType)
			case contentType == "application/x-www-form-urlencoded":
				tag = "Formdata"
			case strings.HasPrefix(contentType, "multipart/"):
				tag = "Multipart"
			case contentType == "text/plain":
				tag = "Text"
			default:
				rcd := schema.ResponseContentDefinition{
					ContentType: contentType,
				}
				responseContentDefinitions = append(responseContentDefinitions, rcd)
				continue
			}

			responseTypeName := operationID + statusCode + tag + "Response"
			contentSchema, err := schema.GenerateGoSchema(content.Schema, []string{responseTypeName})
			if err != nil {
				return nil, fmt.Errorf("error generating request body definition: %w", err)
			}

			rcd := schema.ResponseContentDefinition{
				ContentType: contentType,
				NameTag:     tag,
				Schema:      contentSchema,
			}

			responseContentDefinitions = append(responseContentDefinitions, rcd)
		}

		var responseHeaderDefinitions []schema.ResponseHeaderDefinition
		for _, headerName := range schema.SortedMapKeys(response.Headers) {
			header := response.Headers[headerName]
			contentSchema, err := schema.GenerateGoSchema(header.Value.Schema, []string{})
			if err != nil {
				return nil, fmt.Errorf("error generating response header definition: %w", err)
			}
			headerDefinition := schema.ResponseHeaderDefinition{Name: headerName, GoName: schema.SchemaNameToTypeName(headerName), Schema: contentSchema}
			responseHeaderDefinitions = append(responseHeaderDefinitions, headerDefinition)
		}

		rd := schema.ResponseDefinition{
			StatusCode: statusCode,
			Contents:   responseContentDefinitions,
			Headers:    responseHeaderDefinitions,
		}
		if response.Description != nil {
			rd.Description = *response.Description
		}
		if schema.IsGoTypeReference(responseOrRef.Ref) {
			// Convert the reference path to Go type
			refType, err := schema.RefPathToGoType(responseOrRef.Ref)
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

func GenerateTypeDefsForOperation(op schema.OperationDefinition) []schema.TypeDefinition {
	var typeDefs []schema.TypeDefinition
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
func GenerateParamsTypes(op schema.OperationDefinition) []schema.TypeDefinition {
	var typeDefs []schema.TypeDefinition

	objectParams := op.QueryParams
	objectParams = append(objectParams, op.HeaderParams...)
	objectParams = append(objectParams, op.CookieParams...)

	typeName := op.OperationId + "Params"

	s := schema.Schema{}
	for _, param := range objectParams {
		pSchema := param.Schema
		param.Style()
		if pSchema.HasAdditionalProperties {
			propRefName := strings.Join([]string{typeName, param.GoName()}, "_")
			pSchema.RefType = propRefName
			typeDefs = append(typeDefs, schema.TypeDefinition{
				TypeName: propRefName,
				Schema:   param.Schema,
			})
		}
		prop := schema.Property{
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
	s.GoType = schema.GenStructFromSchema(s)

	td := schema.TypeDefinition{
		TypeName: typeName,
		Schema:   s,
	}
	return append(typeDefs, td)
}

// GenerateTypesForOperations generates code for all types produced within operations
func GenerateTypesForOperations(t *template.Template, ops []schema.OperationDefinition) (string, error) {
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
	var td []schema.TypeDefinition
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

// CombineOperationParameters combines the Parameters defined at a global level (Parameters defined for all methods on a given path) with the Parameters defined at a local level (Parameters defined for a specific path), preferring the locally defined parameter over the global one
func CombineOperationParameters(globalParams []schema.ParameterDefinition, localParams []schema.ParameterDefinition) ([]schema.ParameterDefinition, error) {
	allParams := make([]schema.ParameterDefinition, 0, len(globalParams)+len(localParams))
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
