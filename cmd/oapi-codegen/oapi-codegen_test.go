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

func TestWithNewline(t *testing.T) {
	testmap := [][2]string {
		{ "", "" },
		{ "\n", "\n" },
		{ "\n\n", "\n\n" },
		{ "%s", "%s\n" },
		{ "%s\n", "%s\n" },
		{ "%s %d", "%s %d\n" },
		{ "%s %d\n", "%s %d\n" },
	}

	for i := 0; i < len(testmap); i += 1 {
		if withNewline(testmap[i][0]) != testmap[i][1] {
			t.Errorf("withNewline(%q) != %q", testmap[i][0], testmap[i][1])
		}
	}
}
