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
	"fmt"
	"strings"
	"text/template"
)

// This function takes an array of Parameter definition, and generates a valid
// Go parameter declaration from them, eg:
// ", foo int, bar string, baz float32". The preceding comma is there to save
// a lot of work in the template engine.
func genParamArgs(params []ParameterDefinition) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, len(params))
	for i, p := range params {
		paramName := LowercaseFirstCharacter(ToCamelCase(p.ParamName))
		parts[i] = fmt.Sprintf("%s %s", paramName, p.TypeDef)
	}
	return ", " + strings.Join(parts, ", ")
}

// This function is much like the one above, except it only produces the
// types of the parameters for a type declaration. It would produce this
// from the same input as above:
// ", int, string, float32".
func genParamTypes(params []ParameterDefinition) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, len(params))
	for i, p := range params {
		parts[i] = fmt.Sprintf(p.TypeDef)
	}
	return ", " + strings.Join(parts, ", ")
}

// This is another variation of the function above which generates only the
// parameter names:
// ", foo, bar, baz"
func genParamNames(params []ParameterDefinition) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, len(params))
	for i, p := range params {
		paramName := LowercaseFirstCharacter(ToCamelCase(p.ParamName))
		parts[i] = paramName
	}
	return ", " + strings.Join(parts, ", ")
}

func genParamFmtString(path string) string {
	return ReplacePathParamsWithStr(path)
}

// This function map is passed to the template engine, and we can call each
// function here by keyName from the template code.
var TemplateFunctions = template.FuncMap{
	"genParamArgs":        genParamArgs,
	"genParamTypes":       genParamTypes,
	"genParamNames":       genParamNames,
	"genParamFmtString":   genParamFmtString,
	"swaggerUriToEchoUri": SwaggerUriToEchoUri,
	"lcFirst":             LowercaseFirstCharacter,
	"camelCase":           ToCamelCase,
}
