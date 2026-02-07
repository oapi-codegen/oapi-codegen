package codegen

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"strings"
)

// generateEmbeddedSpec produces Go code that embeds the raw OpenAPI spec as
// gzip+base64 encoded data, with a public GetOpenAPISpecJSON() function to
// retrieve the decompressed JSON bytes.
func generateEmbeddedSpec(specData []byte) (string, error) {
	// Gzip compress
	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return "", fmt.Errorf("creating gzip writer: %w", err)
	}
	if _, err := gz.Write(specData); err != nil {
		return "", fmt.Errorf("gzip writing: %w", err)
	}
	if err := gz.Close(); err != nil {
		return "", fmt.Errorf("gzip close: %w", err)
	}

	// Base64 encode
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	// Split into 80-char chunks
	var chunks []string
	for len(encoded) > 0 {
		end := 80
		if end > len(encoded) {
			end = len(encoded)
		}
		chunks = append(chunks, encoded[:end])
		encoded = encoded[end:]
	}

	// Build Go code
	var b strings.Builder

	b.WriteString("// Base64-encoded, gzip-compressed OpenAPI spec.\n")
	b.WriteString("var openAPISpecJSON = []string{\n")
	for _, chunk := range chunks {
		fmt.Fprintf(&b, "\t%q,\n", chunk)
	}
	b.WriteString("}\n\n")

	b.WriteString("// decodeOpenAPISpec decodes and decompresses the embedded spec.\n")
	b.WriteString("func decodeOpenAPISpec() ([]byte, error) {\n")
	b.WriteString("\tjoined := strings.Join(openAPISpecJSON, \"\")\n")
	b.WriteString("\traw, err := base64.StdEncoding.DecodeString(joined)\n")
	b.WriteString("\tif err != nil {\n")
	b.WriteString("\t\treturn nil, fmt.Errorf(\"decoding base64: %w\", err)\n")
	b.WriteString("\t}\n")
	b.WriteString("\tr, err := gzip.NewReader(bytes.NewReader(raw))\n")
	b.WriteString("\tif err != nil {\n")
	b.WriteString("\t\treturn nil, fmt.Errorf(\"creating gzip reader: %w\", err)\n")
	b.WriteString("\t}\n")
	b.WriteString("\tdefer r.Close()\n")
	b.WriteString("\tvar out bytes.Buffer\n")
	b.WriteString("\tif _, err := out.ReadFrom(r); err != nil {\n")
	b.WriteString("\t\treturn nil, fmt.Errorf(\"decompressing: %w\", err)\n")
	b.WriteString("\t}\n")
	b.WriteString("\treturn out.Bytes(), nil\n")
	b.WriteString("}\n\n")

	b.WriteString("// decodeOpenAPISpecCached returns a closure that caches the decoded spec.\n")
	b.WriteString("func decodeOpenAPISpecCached() func() ([]byte, error) {\n")
	b.WriteString("\tvar cached []byte\n")
	b.WriteString("\tvar cachedErr error\n")
	b.WriteString("\tvar once sync.Once\n")
	b.WriteString("\treturn func() ([]byte, error) {\n")
	b.WriteString("\t\tonce.Do(func() {\n")
	b.WriteString("\t\t\tcached, cachedErr = decodeOpenAPISpec()\n")
	b.WriteString("\t\t})\n")
	b.WriteString("\t\treturn cached, cachedErr\n")
	b.WriteString("\t}\n")
	b.WriteString("}\n\n")

	b.WriteString("var openAPISpec = decodeOpenAPISpecCached()\n\n")

	b.WriteString("// GetOpenAPISpecJSON returns the raw OpenAPI spec as JSON bytes.\n")
	b.WriteString("func GetOpenAPISpecJSON() ([]byte, error) {\n")
	b.WriteString("\treturn openAPISpec()\n")
	b.WriteString("}\n")

	return b.String(), nil
}
