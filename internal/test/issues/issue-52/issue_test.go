package issue_52

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"

	"github.com/tidepool-org/oapi-codegen/pkg/codegen"
)

const spec = `
openapi: 3.0.2
info:
  version: '0.0.1'
  title: example
  desscription: |
    Make sure that recursive types are handled properly
paths:
  /example:
    get:
      operationId: exampleGet
      responses:
        '200':
          description: "OK"
          content:
            'application/json':
              schema:
                $ref: '#/components/schemas/Document'
components:
  schemas:
    Document:
      type: object
      properties:
        fields:
          type: object
          additionalProperties:
            $ref: '#/components/schemas/Value'
    Value:
      type: object
      properties:
        stringValue:
          type: string
        arrayValue:
          $ref: '#/components/schemas/ArrayValue'
    ArrayValue:
      type: array
      items:
        $ref: '#/components/schemas/Value'
`

func TestIssue(t *testing.T) {
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(spec))
	require.NoError(t, err)

	opts := codegen.Options{
		GenerateClient:     true,
		GenerateEchoServer: true,
		GenerateTypes:      true,
		EmbedSpec:          true,
	}

	_, err = codegen.Generate(swagger, "issue_52", opts)
	require.NoError(t, err)
}
