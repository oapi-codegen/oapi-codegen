package util

import (
	"fmt"
	"io/ioutil"
	"net/url"
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
	dir := filepath.Dir(filePath)
	switch ext {
	// The YAML handler can parse both YAML and JSON
	case ".yaml", ".yml", ".json":
		loader := openapi3.NewSwaggerLoader()
		loader.IsExternalRefsAllowed = true
		swagger, err = loader.LoadSwaggerFromDataWithPath(data, &url.URL{Path: dir + "/"})
	default:
		return nil, fmt.Errorf("%s is not a supported extension, use .yaml, .yml or .json", ext)
	}
	if err != nil {
		return nil, err
	}
	return swagger, nil
}
