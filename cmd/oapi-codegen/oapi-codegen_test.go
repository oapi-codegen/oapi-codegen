package main

import (
	"testing"

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
