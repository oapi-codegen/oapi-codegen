package main

import (
	"os"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/deepmap/oapi-codegen/pkg/codegen"
	"github.com/deepmap/oapi-codegen/pkg/util"
)

func TestLoader(t *testing.T) {

	paths := []string{
		"../../examples/petstore-expanded/petstore-expanded.yaml",
		"https://petstore3.swagger.io/api/v3/openapi.json",
	}

	for _, v := range paths {

		swagger, err := util.LoadSwagger(v)
		if err != nil {
			t.Error(err)
		}
		if swagger == nil || swagger.Info == nil || swagger.Info.Version == "" {
			t.Error("missing data")
		}
	}
}

func TestImportMapping_bug_1093(t *testing.T) {
	for _, scenario := range []struct {
		name    string
		config  string
		swagger string
	}{
		{
			name:    "parent",
			config:  "../../examples/import-mapping/parent.cfg.yaml",
			swagger: "../../examples/import-mapping/parent.api.yaml",
		}, {
			name:    "child",
			config:  "../../examples/import-mapping/child.cfg.yaml",
			swagger: "../../examples/import-mapping/child.api.yaml",
		},
	} {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			swagger, err := util.LoadSwagger(scenario.swagger)
			if err != nil {
				t.Error(err)
			}
			opts, err := readConfig(scenario.config)
			if err != nil {
				t.Error(err)
			}
			_, err = codegen.Generate(swagger, opts.Configuration)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func readConfig(configFile string) (opts configuration, err error) {
	var buf []byte
	buf, err = os.ReadFile(configFile)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(buf, &opts)
	if err != nil {
		return
	}

	opts.Configuration = opts.UpdateDefaults()

	// Now, ensure that the config options are valid.
	err = opts.Validate()
	return
}
