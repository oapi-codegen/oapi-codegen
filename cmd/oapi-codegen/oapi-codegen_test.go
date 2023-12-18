package main

import (
	"testing"

	"github.com/deepmap/oapi-codegen/v2/pkg/codegen/openapi"
)

func TestLoader(t *testing.T) {

	paths := []string{
		"../../examples/petstore-expanded/petstore-expanded.yaml",
		"https://petstore3.swagger.io/api/v3/openapi.json",
	}

	for _, v := range paths {

		swagger, err := openapi.LoadOpenAPI(v)
		if err != nil {
			t.Error(err)
		}
		if swagger == nil || swagger.Model.Info == nil || swagger.Model.Info.Version == "" {
			t.Error("missing data")
		}
	}
}
