package util

import (
	"github.com/getkin/kin-openapi/openapi3"
)

func LoadSwagger(filePath string) (swagger *openapi3.Swagger, err error) {
	loader := openapi3.NewSwaggerLoader()
	loader.IsExternalRefsAllowed = false
	swagger, err = loader.LoadSwaggerFromFile(filePath)
	return
}
