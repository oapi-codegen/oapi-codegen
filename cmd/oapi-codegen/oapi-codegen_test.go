package main

import (
	"testing"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
)

func TestLoader(t *testing.T) {

	paths := []string{
		"../../examples/petstore-expanded/petstore-expanded.yaml",
		"https://raw.githubusercontent.com/oapi-codegen/oapi-codegen/v2.4.1/examples/petstore-expanded/petstore-expanded.yaml",
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
