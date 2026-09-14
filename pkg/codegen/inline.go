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
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"fmt"
	"maps"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
)

// GenerateInlinedSpec generates a gzipped, base64 encoded JSON representation of the
// swagger definition, which we embed inside the generated code.
func GenerateInlinedSpec(t *template.Template, importMapping importMap, swagger *openapi3.T) (string, error) {
	encoded, err := marshalInlinedSpec(swagger)
	if err != nil {
		return "", err
	}

	// flate
	var buf bytes.Buffer
	zw, err := flate.NewWriter(&buf, flate.BestCompression)
	if err != nil {
		return "", fmt.Errorf("new flate writer: %w", err)
	}

	if _, err := zw.Write(encoded); err != nil {
		return "", fmt.Errorf("write flate: %w", err)
	}

	if err := zw.Close(); err != nil {
		return "", fmt.Errorf("close flate writer: %w", err)
	}

	str := base64.StdEncoding.EncodeToString(buf.Bytes())

	var parts []string
	const width = 80

	// Chop up the string into an array of strings.
	for len(str) > width {
		part := str[0:width]
		parts = append(parts, part)
		str = str[width:]
	}
	if len(str) > 0 {
		parts = append(parts, str)
	}

	return GenerateTemplates(
		[]string{"inline.tmpl"},
		t,
		struct {
			SpecParts     []string
			ImportMapping importMap
		}{
			SpecParts:     parts,
			ImportMapping: importMapping,
		})
}

func marshalInlinedSpec(swagger *openapi3.T) ([]byte, error) {
	// ensure that any external file references are embedded into the embedded spec
	swagger.InternalizeRefs(context.Background(), nil)
	inlineSchemaRefSiblings(swagger)

	// Marshal to json
	encoded, err := swagger.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error marshaling swagger: %w", err)
	}
	return encoded, nil
}

// inlineSchemaRefSiblings preserves the resolved value of OpenAPI 3.1 schema
// references which originally had JSON Schema keywords next to $ref.
// kin-openapi applies those keywords to Value while loading, but marshaling a
// SchemaRef with a non-empty Ref emits only $ref and extensions.
func inlineSchemaRefSiblings(swagger *openapi3.T) {
	if !swagger.IsOpenAPI31OrLater() {
		return
	}

	_ = walkSwagger(swagger, func(ref RefWrapper) (bool, error) {
		schemaRef, ok := ref.SourceRef.(*openapi3.SchemaRef)
		if !ok || schemaRef.Ref == "" {
			return true, nil
		}

		// Do not follow resolved references: their definitions are visited through
		// components, and following them here can recurse forever for cyclic schemas.
		if schemaRef.Value == nil || !hasSchemaRefKeywordSibling(schemaRef) {
			return false, nil
		}

		if len(schemaRef.Extensions) != 0 {
			if schemaRef.Value.Extensions == nil {
				schemaRef.Value.Extensions = make(map[string]any, len(schemaRef.Extensions))
			}
			maps.Copy(schemaRef.Value.Extensions, schemaRef.Extensions)
		}
		schemaRef.Ref = ""
		return false, nil
	})
}

func hasSchemaRefKeywordSibling(ref *openapi3.SchemaRef) bool {
	if ref.Origin == nil {
		return false
	}
	for _, field := range ref.Origin.Fields {
		if field.Name != "$ref" && !strings.HasPrefix(field.Name, "x-") {
			return true
		}
	}
	return false
}
