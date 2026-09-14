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
	"errors"
	"fmt"
	"maps"
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
		siblingFields := schemaRefKeywordSiblingFields(schemaRef)
		if schemaRef.Value == nil || len(siblingFields) == 0 {
			return false, nil
		}

		// Clearing this reference would make encoding/json follow Value. Keep
		// the reference as a serialization boundary if the same SchemaRef is
		// reachable from its resolved value.
		if schemaRefValueIsCyclic(schemaRef) {
			preserveSchemaRefKeywordSiblings(schemaRef, siblingFields)
			return false, nil
		}

		if len(schemaRef.Extensions) != 0 {
			extensions := maps.Clone(schemaRef.Value.Extensions)
			if extensions == nil {
				extensions = make(map[string]any, len(schemaRef.Extensions))
			}
			maps.Copy(extensions, schemaRef.Extensions)
			schemaRef.Value.Extensions = extensions
		}
		schemaRef.Ref = ""
		return false, nil
	})
}

func schemaRefKeywordSiblingFields(ref *openapi3.SchemaRef) []string {
	// SchemaRef keeps the original sibling field names internally and exposes
	// them through ExtraSiblingFieldsError. Clear Value on a shallow copy so
	// validation only inspects this reference and cannot report an error from
	// the resolved schema graph.
	probe := *ref
	probe.Value = nil
	var siblingErr *openapi3.ExtraSiblingFieldsError
	if !errors.As(probe.Validate(context.Background()), &siblingErr) {
		return nil
	}
	return siblingErr.Fields
}

func preserveSchemaRefKeywordSiblings(ref *openapi3.SchemaRef, fields []string) {
	// SchemaRef.MarshalJSON emits Ref and Extensions only. Put just the original
	// sibling fields into a cloned extension map so the reference can remain a
	// serialization boundary without losing its OpenAPI 3.1 sibling keywords.
	marshaled, _ := ref.Value.MarshalYAML()
	valueFields := marshaled.(map[string]any)

	extensions := maps.Clone(ref.Extensions)
	if extensions == nil {
		extensions = make(map[string]any, len(fields))
	}
	for _, field := range fields {
		if value, ok := valueFields[field]; ok {
			extensions[field] = value
		}
	}
	ref.Extensions = extensions
}

func schemaRefValueIsCyclic(root *openapi3.SchemaRef) bool {
	visiting := make(map[*openapi3.Schema]bool)
	visited := make(map[*openapi3.Schema]bool)

	var visit func(*openapi3.Schema) bool
	visit = func(schema *openapi3.Schema) bool {
		if schema == nil {
			return false
		}
		if visiting[schema] {
			return true
		}
		if visited[schema] {
			return false
		}

		visiting[schema] = true
		for _, child := range schemaChildRefs(schema) {
			if child == nil || child.Value == nil {
				continue
			}
			// A non-empty ref is normally a serialization boundary. The root is
			// the exception because the caller is considering clearing its Ref.
			if child.Ref != "" && child != root {
				continue
			}
			if visit(child.Value) {
				return true
			}
		}
		visiting[schema] = false
		visited[schema] = true
		return false
	}

	return root != nil && visit(root.Value)
}
