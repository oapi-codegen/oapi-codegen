package codegen

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateEmbeddedSpec(t *testing.T) {
	specData := []byte(`{"openapi":"3.0.0","info":{"title":"Test","version":"1.0"}}`)

	code, err := generateEmbeddedSpec(specData)
	require.NoError(t, err)

	// Should contain the chunked base64 variable
	assert.Contains(t, code, "var openAPISpecJSON = []string{")

	// Should contain the decode function
	assert.Contains(t, code, "func decodeOpenAPISpec() ([]byte, error)")

	// Should contain the cached decode function
	assert.Contains(t, code, "func decodeOpenAPISpecCached() func() ([]byte, error)")

	// Should contain the public API
	assert.Contains(t, code, "func GetOpenAPISpecJSON() ([]byte, error)")

	// Should contain the cached var
	assert.Contains(t, code, "var openAPISpec = decodeOpenAPISpecCached()")
}

func TestGenerateEmbeddedSpecRoundTrip(t *testing.T) {
	specData := []byte(`{"openapi":"3.0.0","info":{"title":"Test API","version":"1.0"},"paths":{}}`)

	code, err := generateEmbeddedSpec(specData)
	require.NoError(t, err)

	// Extract the base64 chunks from the generated code
	var chunks []string
	for _, line := range strings.Split(code, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, `"`) && strings.HasSuffix(line, `",`) {
			// Remove quotes and trailing comma
			chunk := line[1 : len(line)-2]
			chunks = append(chunks, chunk)
		}
	}
	require.NotEmpty(t, chunks, "should have extracted base64 chunks")

	// Decode base64
	joined := strings.Join(chunks, "")
	raw, err := base64.StdEncoding.DecodeString(joined)
	require.NoError(t, err)

	// Decompress gzip
	r, err := gzip.NewReader(bytes.NewReader(raw))
	require.NoError(t, err)
	defer func() { _ = r.Close() }()

	var out bytes.Buffer
	_, err = out.ReadFrom(r)
	require.NoError(t, err)

	// Should match original spec
	assert.Equal(t, specData, out.Bytes())
}

func TestGenerateEmbeddedSpecInGenerate(t *testing.T) {
	spec := `openapi: "3.0.0"
info:
  title: Test API
  version: "1.0"
paths: {}
components:
  schemas:
    Pet:
      type: object
      properties:
        name:
          type: string
`

	specBytes := []byte(spec)

	doc, err := libopenapi.NewDocument(specBytes)
	require.NoError(t, err)

	cfg := Configuration{
		PackageName: "testpkg",
	}

	code, err := Generate(doc, specBytes, cfg)
	require.NoError(t, err)

	// Should contain the model type
	assert.Contains(t, code, "type Pet struct")

	// Should contain the embedded spec
	assert.Contains(t, code, "GetOpenAPISpecJSON")
	assert.Contains(t, code, "openAPISpecJSON")
}

func TestGenerateWithNilSpecData(t *testing.T) {
	spec := `openapi: "3.0.0"
info:
  title: Test API
  version: "1.0"
paths: {}
components:
  schemas:
    Pet:
      type: object
      properties:
        name:
          type: string
`

	doc, err := libopenapi.NewDocument([]byte(spec))
	require.NoError(t, err)

	cfg := Configuration{
		PackageName: "testpkg",
	}

	code, err := Generate(doc, nil, cfg)
	require.NoError(t, err)

	// Should contain the model type
	assert.Contains(t, code, "type Pet struct")

	// Should NOT contain the embedded spec
	assert.NotContains(t, code, "GetOpenAPISpecJSON")
	assert.NotContains(t, code, "openAPISpecJSON")
}
