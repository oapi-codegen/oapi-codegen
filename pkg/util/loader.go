package util

import (
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
)

// LoadSwagger loads a swagger specification from filePath.
// filePath can be either a URI or local file.
func LoadSwagger(filePath string) (swagger *openapi3.T, err error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	u, err := url.Parse(filePath)
	if err == nil && u.Scheme != "" && u.Host != "" {
		return loader.LoadFromURI(u)
	}

	return loader.LoadFromFile(filePath)
}
