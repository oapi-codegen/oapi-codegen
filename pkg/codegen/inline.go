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
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
)

// GenerateInlinedSpec generates a flate-compressed, base64 encoded JSON representation of the
// swagger definition, which we embed inside the generated code.
func GenerateInlinedSpec(t *template.Template, importMapping importMap, swagger *openapi3.T) (string, error) {
	// ensure that any external file references are embedded into the embedded spec
	swagger.InternalizeRefs(context.Background(), nil)
	// Marshal to json
	encoded, err := swagger.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("error marshaling swagger: %w", err)
	}

	str, err := compressSpec(encoded)
	if err != nil {
		return "", err
	}

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

// compressSpec flate-compresses data and base64-encodes the result. The
// compression settings here must match the flate.NewReader call in
// inline.tmpl's decodeSpec, otherwise generated code will fail to
// decompress its own embedded spec.
func compressSpec(data []byte) (string, error) {
	var buf bytes.Buffer
	zw, err := flate.NewWriter(&buf, flate.BestCompression)
	if err != nil {
		return "", fmt.Errorf("new flate writer: %w", err)
	}

	if _, err := zw.Write(data); err != nil {
		return "", fmt.Errorf("write flate: %w", err)
	}

	if err := zw.Close(); err != nil {
		return "", fmt.Errorf("close flate writer: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
