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
	"text/template"
)

// This generates a gzipped, base64 encoded JSON representation of the
// swagger definition, which we embed inside the generated code.
func GenerateInlinedSpec(t *template.Template, importMapping importMap, opts Options) (string, error) {
	// Marshal to json

	var err error
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	err = t.ExecuteTemplate(w, "inline.tmpl", struct {
		FileBase      string
		ImportMapping importMap
	}{FileBase: opts.SpecFileName, ImportMapping: importMapping})
	if err != nil {
		return "", fmt.Errorf("error generating inlined spec: %s", err)
	}
	err = w.Flush()
	if err != nil {
		return "", fmt.Errorf("error flushing output buffer for inlined spec: %s", err)
	}
	return buf.String(), nil
}
