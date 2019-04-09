package codegen

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
)

// This generates a gzipped, base64 encoded JSON representation of the
// swagger definition, which we embed inside the generated code.
func GenerateInlinedSpec(t *template.Template, swagger *openapi3.Swagger) (string, error)  {
	// Marshal to json
	encoded, err := swagger.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("error marshaling swagger: %s", err)
	}

	// gzip
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return "", fmt.Errorf("error creating gzip compressor: %s", err)
	}
	_, err = zw.Write(encoded)
	if err != nil {
		return "", fmt.Errorf("error gzipping swagger file: %s", err)
	}
	err = zw.Close()
	if err != nil {
		return "", fmt.Errorf("error gzipping swagger file: %s", err)
	}
	str := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Generate inline code.
	buf.Reset()
	w := bufio.NewWriter(&buf)
	err = t.ExecuteTemplate(w,"inline.tmpl", str)
	if err != nil {
		return "", fmt.Errorf("error generating inlined spec: %s", err)
	}
	err = w.Flush()
	if err != nil {
		return "", fmt.Errorf("error flushing output buffer for inlined spec: %s", err)
	}
	return buf.String(), nil
}
