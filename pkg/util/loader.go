package util

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

func LoadSwagger(filePath string) (*openapi3.Swagger, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var swagger *openapi3.Swagger
	ext := filepath.Ext(filePath)
	ext = strings.ToLower(ext)
	switch ext {
	case ".yaml", ".yml":
		swagger, err = openapi3.NewSwaggerLoader().LoadSwaggerFromYAMLData(data)
	case ".json":
		swagger, err = openapi3.NewSwaggerLoader().LoadSwaggerFromData(data)
	default:
		return nil, fmt.Errorf("%s is not a supported extension, use .yaml, .yml or .json", ext)
	}
	if err != nil {
		return nil, err
	}
	return swagger, nil
}
