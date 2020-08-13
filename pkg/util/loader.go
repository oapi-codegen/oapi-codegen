package util

import (
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
)

func LoadSwagger(filePath string) (swagger *openapi3.Swagger, err error) {

	loader := openapi3.NewSwaggerLoader()
	loader.IsExternalRefsAllowed = true

	u, err := url.Parse(filePath)
	if err == nil && u.Scheme != "" && u.Host != "" {
		return loader.LoadSwaggerFromURI(u)
	} else {
		return loader.LoadSwaggerFromFile(filePath)
	}
}
