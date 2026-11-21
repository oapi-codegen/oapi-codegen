package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

func TestSkipUnimplemented(t *testing.T) {
	const spec = `openapi: 3.0.0
info:
  title: widgets
  version: 1.0.0
paths:
  /widgets/{id}:
    get:
      operationId: getWidget
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
        - name: limit
          in: query
          required: true
          schema:
            type: integer
      responses:
        '200':
          description: OK
`
	for _, tt := range []struct {
		name   string
		option string
		skip   bool
	}{
		{name: "default"},
		{name: "skip", option: "output-options:\n  skip-unimplemented: true\n", skip: true},
		{name: "explicit false", option: "output-options:\n  skip-unimplemented: false\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Configuration
			decoder := yaml.NewDecoder(strings.NewReader(`package: api
generate:
  chi-server: true
  models: true
` + tt.option))
			decoder.KnownFields(true)
			require.NoError(t, decoder.Decode(&cfg))
			require.NoError(t, cfg.Validate())

			code := generateFromSpec(t, spec, cfg)
			if tt.skip {
				assert.NotContains(t, code, "Unimplemented")
			} else {
				assert.Contains(t, code, "type Unimplemented struct{}")
				assert.Contains(t, code, "func (_ Unimplemented) GetWidget(")
				assert.Contains(t, code, "w.WriteHeader(http.StatusNotImplemented)")
			}
			assert.Contains(t, code, "type ServerInterface interface {")
			assert.Contains(t, code, "GetWidget(w http.ResponseWriter, r *http.Request, id string, params GetWidgetParams)")
			assert.Contains(t, code, "func (siw *ServerInterfaceWrapper) GetWidget(")
			assert.Contains(t, code, "func Handler(si ServerInterface) http.Handler")
		})
	}
}
