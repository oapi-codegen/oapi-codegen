package codegen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// GatherOperations traverses an OpenAPI document and collects all operations.
func GatherOperations(doc libopenapi.Document, paramTracker *ParamUsageTracker) ([]*OperationDescriptor, error) {
	model, err := doc.BuildV3Model()
	if err != nil {
		return nil, fmt.Errorf("building v3 model: %w", err)
	}
	if model == nil {
		return nil, fmt.Errorf("failed to build v3 model")
	}

	g := &operationGatherer{
		paramTracker: paramTracker,
	}

	return g.gatherFromDocument(&model.Model)
}

type operationGatherer struct {
	paramTracker *ParamUsageTracker
}

func (g *operationGatherer) gatherFromDocument(doc *v3.Document) ([]*OperationDescriptor, error) {
	var operations []*OperationDescriptor

	if doc.Paths == nil || doc.Paths.PathItems == nil {
		return operations, nil
	}

	// Collect paths in sorted order for deterministic output
	var paths []string
	for pair := doc.Paths.PathItems.First(); pair != nil; pair = pair.Next() {
		paths = append(paths, pair.Key())
	}
	sort.Strings(paths)

	for _, pathStr := range paths {
		pathItem := doc.Paths.PathItems.GetOrZero(pathStr)
		if pathItem == nil {
			continue
		}

		// Gather path-level parameters (shared by all operations on this path)
		globalParams, err := g.gatherParameters(pathItem.Parameters)
		if err != nil {
			return nil, fmt.Errorf("error gathering path-level parameters for %s: %w", pathStr, err)
		}

		// Process each operation on this path
		ops := pathItem.GetOperations()
		if ops == nil {
			continue
		}

		// Collect methods in sorted order
		var methods []string
		for pair := ops.First(); pair != nil; pair = pair.Next() {
			methods = append(methods, pair.Key())
		}
		sort.Strings(methods)

		for _, method := range methods {
			op := ops.GetOrZero(method)
			if op == nil {
				continue
			}

			opDesc, err := g.gatherOperation(method, pathStr, op, globalParams)
			if err != nil {
				return nil, fmt.Errorf("error gathering operation %s %s: %w", method, pathStr, err)
			}
			operations = append(operations, opDesc)
		}
	}

	return operations, nil
}

func (g *operationGatherer) gatherOperation(method, path string, op *v3.Operation, globalParams []*ParameterDescriptor) (*OperationDescriptor, error) {
	// Determine operation ID
	operationID := op.OperationId
	if operationID == "" {
		operationID = generateOperationID(method, path)
	}
	goOperationID := ToGoIdentifier(operationID)

	// Gather operation-level parameters
	localParams, err := g.gatherParameters(op.Parameters)
	if err != nil {
		return nil, fmt.Errorf("error gathering parameters: %w", err)
	}

	// Combine global and local parameters (local overrides global)
	allParams := combineParameters(globalParams, localParams)

	// Sort path params to match order in path
	pathParams := filterParamsByLocation(allParams, "path")
	pathParams, err = sortPathParamsByPath(path, pathParams)
	if err != nil {
		return nil, fmt.Errorf("error sorting path params: %w", err)
	}

	// Gather request bodies
	bodies, err := g.gatherRequestBodies(operationID, op.RequestBody)
	if err != nil {
		return nil, fmt.Errorf("error gathering request bodies: %w", err)
	}

	// Gather responses
	responses, err := g.gatherResponses(operationID, op.Responses)
	if err != nil {
		return nil, fmt.Errorf("error gathering responses: %w", err)
	}

	// Gather security requirements
	security := g.gatherSecurity(op.Security)

	queryParams := filterParamsByLocation(allParams, "query")
	headerParams := filterParamsByLocation(allParams, "header")
	cookieParams := filterParamsByLocation(allParams, "cookie")

	hasParams := len(queryParams)+len(headerParams)+len(cookieParams) > 0

	desc := &OperationDescriptor{
		OperationID:   operationID,
		GoOperationID: goOperationID,
		Method:        strings.ToUpper(method),
		Path:          path,
		Summary:       op.Summary,
		Description:   op.Description,

		PathParams:   pathParams,
		QueryParams:  queryParams,
		HeaderParams: headerParams,
		CookieParams: cookieParams,

		Bodies:    bodies,
		Responses: responses,
		Security:  security,

		HasBody:        len(bodies) > 0,
		HasParams:      hasParams,
		ParamsTypeName: goOperationID + "Params",

		Spec: op,
	}

	return desc, nil
}

func (g *operationGatherer) gatherParameters(params []*v3.Parameter) ([]*ParameterDescriptor, error) {
	var result []*ParameterDescriptor

	for _, param := range params {
		if param == nil {
			continue
		}

		desc, err := g.gatherParameter(param)
		if err != nil {
			return nil, fmt.Errorf("error gathering parameter %s: %w", param.Name, err)
		}
		result = append(result, desc)
	}

	return result, nil
}

func (g *operationGatherer) gatherParameter(param *v3.Parameter) (*ParameterDescriptor, error) {
	// Determine style and explode (with defaults based on location)
	style := param.Style
	if style == "" {
		style = DefaultParamStyle(param.In)
	}

	explode := DefaultParamExplode(param.In)
	if param.Explode != nil {
		explode = *param.Explode
	}

	// Record param usage for function generation
	if g.paramTracker != nil {
		g.paramTracker.RecordParam(style, explode)
	}

	// Determine encoding mode
	isStyled := param.Schema != nil
	isJSON := false
	isPassThrough := false

	if param.Content != nil && param.Content.Len() > 0 {
		// Parameter uses content encoding
		isStyled = false
		for pair := param.Content.First(); pair != nil; pair = pair.Next() {
			contentType := pair.Key()
			if IsMediaTypeJSON(contentType) {
				isJSON = true
				break
			}
		}
		if !isJSON {
			isPassThrough = true
		}
	}

	// Get type declaration from schema
	typeDecl := "string" // Default
	var schemaDesc *SchemaDescriptor
	if param.Schema != nil {
		schema := param.Schema.Schema()
		if schema != nil {
			schemaDesc = &SchemaDescriptor{
				Schema: schema,
			}
			typeDecl = schemaToGoType(schema)
		}
	}

	goName := ToCamelCase(param.Name)

	// Handle *bool for Required
	required := false
	if param.Required != nil {
		required = *param.Required
	}

	desc := &ParameterDescriptor{
		Name:     param.Name,
		GoName:   goName,
		Location: param.In,
		Required: required,

		Style:   style,
		Explode: explode,

		Schema:   schemaDesc,
		TypeDecl: typeDecl,

		StyleFunc: ComputeStyleFunc(style, explode),
		BindFunc:  ComputeBindFunc(style, explode),

		IsStyled:      isStyled,
		IsPassThrough: isPassThrough,
		IsJSON:        isJSON,

		Spec: param,
	}

	return desc, nil
}

func (g *operationGatherer) gatherRequestBodies(operationID string, bodyRef *v3.RequestBody) ([]*RequestBodyDescriptor, error) {
	if bodyRef == nil {
		return nil, nil
	}

	var bodies []*RequestBodyDescriptor

	if bodyRef.Content == nil {
		return bodies, nil
	}

	// Collect content types in sorted order
	var contentTypes []string
	for pair := bodyRef.Content.First(); pair != nil; pair = pair.Next() {
		contentTypes = append(contentTypes, pair.Key())
	}
	sort.Strings(contentTypes)

	// Determine which is the default (application/json if present)
	hasApplicationJSON := false
	for _, ct := range contentTypes {
		if ct == "application/json" {
			hasApplicationJSON = true
			break
		}
	}

	for _, contentType := range contentTypes {
		mediaType := bodyRef.Content.GetOrZero(contentType)
		if mediaType == nil {
			continue
		}

		nameTag := ComputeBodyNameTag(contentType)
		isDefault := contentType == "application/json" || (!hasApplicationJSON && contentType == contentTypes[0])

		var schemaDesc *SchemaDescriptor
		if mediaType.Schema != nil {
			if schema := mediaType.Schema.Schema(); schema != nil {
				schemaDesc = &SchemaDescriptor{
					Schema: schema,
				}
			}
		}

		funcSuffix := ""
		if !isDefault && nameTag != "" {
			funcSuffix = "With" + nameTag + "Body"
		}

		goTypeName := operationID + nameTag + "RequestBody"
		if nameTag == "" {
			goTypeName = operationID + "RequestBody"
		}

		// Handle *bool for Required
		bodyRequired := false
		if bodyRef.Required != nil {
			bodyRequired = *bodyRef.Required
		}

		desc := &RequestBodyDescriptor{
			ContentType: contentType,
			Required:    bodyRequired,
			Schema:      schemaDesc,

			NameTag:    nameTag,
			GoTypeName: goTypeName,
			FuncSuffix: funcSuffix,
			IsDefault:  isDefault,
			IsJSON:     IsMediaTypeJSON(contentType),
		}

		// Gather encoding options for form data
		if mediaType.Encoding != nil && mediaType.Encoding.Len() > 0 {
			desc.Encoding = make(map[string]RequestBodyEncoding)
			for pair := mediaType.Encoding.First(); pair != nil; pair = pair.Next() {
				enc := pair.Value()
				desc.Encoding[pair.Key()] = RequestBodyEncoding{
					ContentType: enc.ContentType,
					Style:       enc.Style,
					Explode:     enc.Explode,
				}
			}
		}

		bodies = append(bodies, desc)
	}

	return bodies, nil
}

func (g *operationGatherer) gatherResponses(operationID string, responses *v3.Responses) ([]*ResponseDescriptor, error) {
	if responses == nil {
		return nil, nil
	}

	var result []*ResponseDescriptor

	// Gather default response
	if responses.Default != nil {
		desc, err := g.gatherResponse(operationID, "default", responses.Default)
		if err != nil {
			return nil, err
		}
		if desc != nil {
			result = append(result, desc)
		}
	}

	// Gather status code responses
	if responses.Codes != nil {
		var codes []string
		for pair := responses.Codes.First(); pair != nil; pair = pair.Next() {
			codes = append(codes, pair.Key())
		}
		sort.Strings(codes)

		for _, code := range codes {
			respRef := responses.Codes.GetOrZero(code)
			if respRef == nil {
				continue
			}

			desc, err := g.gatherResponse(operationID, code, respRef)
			if err != nil {
				return nil, err
			}
			if desc != nil {
				result = append(result, desc)
			}
		}
	}

	return result, nil
}

func (g *operationGatherer) gatherResponse(operationID, statusCode string, resp *v3.Response) (*ResponseDescriptor, error) {
	if resp == nil {
		return nil, nil
	}

	var contents []*ResponseContentDescriptor
	if resp.Content != nil {
		var contentTypes []string
		for pair := resp.Content.First(); pair != nil; pair = pair.Next() {
			contentTypes = append(contentTypes, pair.Key())
		}
		sort.Strings(contentTypes)

		for _, contentType := range contentTypes {
			mediaType := resp.Content.GetOrZero(contentType)
			if mediaType == nil {
				continue
			}

			var schemaDesc *SchemaDescriptor
			if mediaType.Schema != nil {
				schemaDesc = schemaProxyToDescriptor(mediaType.Schema)
			}

			nameTag := ComputeBodyNameTag(contentType)

			contents = append(contents, &ResponseContentDescriptor{
				ContentType: contentType,
				Schema:      schemaDesc,
				NameTag:     nameTag,
				IsJSON:      IsMediaTypeJSON(contentType),
			})
		}
	}

	var headers []*ResponseHeaderDescriptor
	if resp.Headers != nil {
		var headerNames []string
		for pair := resp.Headers.First(); pair != nil; pair = pair.Next() {
			headerNames = append(headerNames, pair.Key())
		}
		sort.Strings(headerNames)

		for _, name := range headerNames {
			header := resp.Headers.GetOrZero(name)
			if header == nil {
				continue
			}

			var schemaDesc *SchemaDescriptor
			if header.Schema != nil {
				schemaDesc = schemaProxyToDescriptor(header.Schema)
			}

			headers = append(headers, &ResponseHeaderDescriptor{
				Name:     name,
				GoName:   ToCamelCase(name),
				Required: header.Required,
				Schema:   schemaDesc,
			})
		}
	}

	description := ""
	if resp.Description != "" {
		description = resp.Description
	}

	return &ResponseDescriptor{
		StatusCode:  statusCode,
		Description: description,
		Contents:    contents,
		Headers:     headers,
	}, nil
}

func (g *operationGatherer) gatherSecurity(security []*base.SecurityRequirement) []SecurityRequirement {
	if security == nil {
		return nil
	}

	var result []SecurityRequirement
	for _, req := range security {
		if req == nil || req.Requirements == nil {
			continue
		}
		for pair := req.Requirements.First(); pair != nil; pair = pair.Next() {
			result = append(result, SecurityRequirement{
				Name:   pair.Key(),
				Scopes: pair.Value(),
			})
		}
	}
	return result
}

// Helper functions

func generateOperationID(method, path string) string {
	// Generate operation ID from method and path
	// GET /users/{id} -> GetUsersId
	id := strings.ToLower(method)
	for _, part := range strings.Split(path, "/") {
		if part == "" {
			continue
		}
		// Remove path parameter braces
		part = strings.TrimPrefix(part, "{")
		part = strings.TrimSuffix(part, "}")
		id += "-" + part
	}
	return ToCamelCase(id)
}

func combineParameters(global, local []*ParameterDescriptor) []*ParameterDescriptor {
	// Local parameters override global parameters with the same name and location
	seen := make(map[string]bool)
	var result []*ParameterDescriptor

	for _, p := range local {
		key := p.Location + ":" + p.Name
		seen[key] = true
		result = append(result, p)
	}

	for _, p := range global {
		key := p.Location + ":" + p.Name
		if !seen[key] {
			result = append(result, p)
		}
	}

	return result
}

func filterParamsByLocation(params []*ParameterDescriptor, location string) []*ParameterDescriptor {
	var result []*ParameterDescriptor
	for _, p := range params {
		if p.Location == location {
			result = append(result, p)
		}
	}
	return result
}

func sortPathParamsByPath(path string, params []*ParameterDescriptor) ([]*ParameterDescriptor, error) {
	// Extract parameter names from path in order
	var pathParamNames []string
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			name := strings.TrimPrefix(part, "{")
			name = strings.TrimSuffix(name, "}")
			pathParamNames = append(pathParamNames, name)
		}
	}

	// Build a map of params by name
	paramMap := make(map[string]*ParameterDescriptor)
	for _, p := range params {
		paramMap[p.Name] = p
	}

	// Sort params according to path order
	var result []*ParameterDescriptor
	for _, name := range pathParamNames {
		if p, ok := paramMap[name]; ok {
			result = append(result, p)
		}
	}

	return result, nil
}

// schemaProxyToDescriptor converts a schema proxy to a basic descriptor.
// This is a simplified version - for full type resolution, use the schema gatherer.
func schemaProxyToDescriptor(proxy *base.SchemaProxy) *SchemaDescriptor {
	if proxy == nil {
		return nil
	}

	schema := proxy.Schema()
	if schema == nil {
		return nil
	}

	desc := &SchemaDescriptor{
		Schema: schema,
	}

	// Capture reference if this is a reference schema
	if proxy.IsReference() {
		desc.Ref = proxy.GetReference()
	}

	return desc
}

// schemaToGoType converts a schema to a Go type string.
// This is a simplified version for parameter types.
func schemaToGoType(schema *base.Schema) string {
	if schema == nil {
		return "interface{}"
	}

	// Check for array
	if schema.Items != nil && schema.Items.A != nil {
		itemType := "interface{}"
		if itemSchema := schema.Items.A.Schema(); itemSchema != nil {
			itemType = schemaToGoType(itemSchema)
		}
		return "[]" + itemType
	}

	// Check explicit type
	for _, t := range schema.Type {
		switch t {
		case "string":
			if schema.Format == "date-time" {
				return "time.Time"
			}
			if schema.Format == "date" {
				return "Date"
			}
			if schema.Format == "uuid" {
				return "uuid.UUID"
			}
			return "string"
		case "integer":
			if schema.Format == "int64" {
				return "int64"
			}
			if schema.Format == "int32" {
				return "int32"
			}
			return "int"
		case "number":
			if schema.Format == "float" {
				return "float32"
			}
			return "float64"
		case "boolean":
			return "bool"
		case "array":
			// Handled above
			return "[]interface{}"
		case "object":
			return "map[string]interface{}"
		}
	}

	return "interface{}"
}
