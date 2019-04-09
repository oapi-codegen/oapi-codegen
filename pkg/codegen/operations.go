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
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
)

type ParameterDefinition struct {
	ParamName string // The original json parameter name, eg param_name
	GoName    string // The Go friendly type name of above, ParamName
	TypeDef   string // The Go type definition of the parameter, "int" or "CustomType"
	Reference string // Swagger reference if present
	In        string // Where the parameter is defined - path, header, cookie, query
	Required  bool   // Is this a required parameter?
	Spec      *openapi3.Parameter
}

// Generate the JSON annotation to map GoType to json type name. If Parameter
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
		_, found := p.Content["application/json"]
		return found
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

// This function walks the given parameters dictionary, and generates the above
// descriptors into a flat list. This makes it a lot easier to traverse the
// data in the template engine.
func DescribeParameters(params openapi3.Parameters) ([]ParameterDefinition, error) {
	outParams := make([]ParameterDefinition, 0)
	for _, paramOrRef := range params {
		param := paramOrRef.Value
		// If this is a reference to a predefined type, simply use the reference
		// name as the type. $ref: "#/components/schemas/custom_type" becomes
		// "CustomType".
		if paramOrRef.Ref != "" {
			// We have a reference to a predefined parameter
			goType, err := RefPathToGoType(paramOrRef.Ref)
			if err != nil {
				return nil, fmt.Errorf("error dereferencing (%s) for param (%s): %s",
					paramOrRef.Ref, param.Name, err)
			}
			pd := ParameterDefinition{
				ParamName: param.Name,
				GoName:    ToCamelCase(param.Name),
				TypeDef:   goType,
				Reference: paramOrRef.Ref,
				In:        param.In,
				Required:  param.Required,
				Spec:      param,
			}
			outParams = append(outParams, pd)
		} else {
			// Inline parameter definition. We'll generate the full Go type
			// definition.
			goType, err := paramToGoType(param)
			if err != nil {
				return nil, fmt.Errorf("error generating type for param (%s): %s",
					param.Name, err)
			}
			pd := ParameterDefinition{
				ParamName: param.Name,
				GoName:    ToCamelCase(param.Name),
				TypeDef:   goType,
				Reference: paramOrRef.Ref,
				In:        param.In,
				Required:  param.Required,
				Spec:      param,
			}
			outParams = append(outParams, pd)
		}
	}
	return outParams, nil
}

// This structure describes an Operation
type OperationDefinition struct {
	PathParams   []ParameterDefinition  // Parameters in the path, eg, /path/:param
	HeaderParams []ParameterDefinition  // Parameters in HTTP headers
	QueryParams  []ParameterDefinition  // Parameters in the query, /path?param
	CookieParams []ParameterDefinition  // Parameters in cookies
	OperationId  string                 // The operation_id description from Swagger, used to generate function names
	Body         *RequestBodyDefinition // The body of the request if it takes one
	Summary      string                 // Summary string from Swagger, used to generate a comment
	Method       string                 // GET, POST, DELETE, etc.
	Path         string                 // The Swagger path for the operation, like /resource/{id}
	Spec         *openapi3.Operation
}

// Returns the list of all parameters except Path parameters. Path parameters
// are handled differently from the rest, since they're mandatory.
func (o *OperationDefinition) Params() []ParameterDefinition {
	result := append(o.QueryParams, o.HeaderParams...)
	result = append(result, o.CookieParams...)
	return result
}

// Returns all parameters
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

// Called by template engine to determine whether to generate a body definition
func (o *OperationDefinition) HasBody() bool {
	return o.Body != nil
}

// Called by the template engine to get the body definition
func (o *OperationDefinition) GetBodyDefinition() RequestBodyDefinition {
	return *(o.Body)
}

// This describes a request body
type RequestBodyDefinition struct {
	TypeDef    string // The go type definition for the body
	Required   bool   // Is this body required, or optional?
	CustomType bool   // Is the type pre-defined, or defined inline?
}

// This function returns the subset of the specified parameters which are of the
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

// This function generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GeneratePathHandlers(t *template.Template, swagger *openapi3.Swagger) (string, error) {
	var operations []OperationDefinition

	for _, pathName := range SortedPathsKeys(swagger.Paths) {
		pathItem := swagger.Paths[pathName]
		// These are parameters defined for all methods on a given path. They
		// are shared by all methods.
		globalParams, err := DescribeParameters(pathItem.Parameters)
		if err != nil {
			return "", fmt.Errorf("error describing global parameters for %s: %s",
				pathName, err)
		}

		// Each path can have a number of operations, POST, GET, OPTIONS, etc.
		pathOps := pathItem.Operations()
		for _, opName := range SortedOperationsKeys(pathOps) {
			op := pathOps[opName]

			// These are parameters defined for the specific path method that
			// we're iterating over.
			localParams, err := DescribeParameters(op.Parameters)
			if err != nil {
				return "", fmt.Errorf("error describing global parameters for %s/%s: %s",
					opName, pathName, err)
			}
			// All the parameters required by a handler are the union of the
			// global parameters and the local parameters.
			allParams := append(globalParams, localParams...)

			// We don't know how to extract parameters from cookies yet, so we'll
			// return an error until we do. This looks like it should be possible
			// to fish out from cookie data via Echo.
			if len(FilterParameterDefinitionByType(allParams, "cookie")) != 0 {
				return "", fmt.Errorf("cookie parameters are not yet supported")
			}

			opDef := OperationDefinition{
				PathParams:   FilterParameterDefinitionByType(allParams, "path"),
				HeaderParams: FilterParameterDefinitionByType(allParams, "header"),
				QueryParams:  FilterParameterDefinitionByType(allParams, "query"),
				CookieParams: FilterParameterDefinitionByType(allParams, "cookie"),
				OperationId:  ToCamelCase(op.OperationID),
				Summary:      op.Summary,
				Method:       opName,
				Path:         pathName,
				Spec:         op,
			}

			// Does request have a body payload?
			if op.RequestBody != nil {
				bodyOrRef := op.RequestBody
				body := bodyOrRef.Value
				if bodyOrRef.Ref != "" {
					// If it's a reference to an existing type, our job is easy,
					// just use that.
					bodyType, err := RefPathToGoType(bodyOrRef.Ref)
					if err != nil {
						return "", fmt.Errorf("error dereferencing type %s for request body: %s",
							bodyOrRef.Ref, err)
					}
					opDef.Body = &RequestBodyDefinition{
						TypeDef:    bodyType,
						Required:   body.Required,
						CustomType: false,
					}
				} else {
					// We only generate the body type inline for application/json
					// content. Users can marshal other body types themselves.
					content, found := body.Content["application/json"]
					if found {
						bodyType, err := schemaToGoType(content.Schema, true)
						if err != nil {
							return "", fmt.Errorf("error generating request body type for operation %s: %s",
								op.OperationID, err)
						}
						opDef.Body = &RequestBodyDefinition{
							TypeDef:    bodyType,
							Required:   body.Required,
							CustomType: content.Schema.Ref == "",
						}
					}
				}
			}

			operations = append(operations, opDef)
		}

	}
	si, err := GenerateServerInterface(t, operations)
	if err != nil {
		return "", fmt.Errorf("Error generating server types and interface: %s", err)
	}

	wrappers, err := GenerateWrappers(t, operations)
	if err != nil {
		return "", fmt.Errorf("Error generating handler wrappers: %s", err)
	}

	register, err := GenerateRegistration(t, operations)
	if err != nil {
		return "", fmt.Errorf("Error generating handler registration: %s", err)
	}
	return strings.Join([]string{si, wrappers, register}, "\n"), nil
}

// Uses the template engine to generate the server interface
func GenerateServerInterface(t *template.Template, ops []OperationDefinition) (string, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	err := t.ExecuteTemplate(w, "server-interface.tmpl", ops)

	if err != nil {
		return "", fmt.Errorf("error generating server interface: %s", err)
	}
	err = w.Flush()
	if err != nil {
		return "", fmt.Errorf("error flushing output buffer for server interface: %s", err)
	}
	return buf.String(), nil
}

// Uses the template engine to generate all the wrappers which wrap our simple
// interface functions and perform marshallin/unmarshalling from HTTP
// request objects.
func GenerateWrappers(t *template.Template, ops []OperationDefinition) (string, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	err := t.ExecuteTemplate(w, "wrappers.tmpl", ops)

	if err != nil {
		return "", fmt.Errorf("error generating server interface: %s", err)
	}
	err = w.Flush()
	if err != nil {
		return "", fmt.Errorf("error flushing output buffer for server interface: %s", err)
	}
	return buf.String(), nil
}

// Uses the template engine to generate the function which registers our wrappers
// as Echo path handlers.
func GenerateRegistration(t *template.Template, ops []OperationDefinition) (string, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	err := t.ExecuteTemplate(w, "register.tmpl", ops)

	if err != nil {
		return "", fmt.Errorf("error generating route registration: %s", err)
	}
	err = w.Flush()
	if err != nil {
		return "", fmt.Errorf("error flushing output buffer for route registration: %s", err)
	}
	return buf.String(), nil
}
