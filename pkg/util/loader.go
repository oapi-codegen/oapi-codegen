package util

import (
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
)

// Deprecated: LoadSwagger is deprecated as that name isn't ours to use. Call
// LoadOpenAPI instead.
func LoadSwagger(filePath string) (*openapi3.T, error) {
	return LoadOpenAPI(filePath)
}

// LoadOpenAPI loads a local or remote OpenAPI spec
func LoadOpenAPI(filePath string) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	u, err := url.Parse(filePath)
	if err == nil && u.Scheme != "" && u.Host != "" {
		return loader.LoadFromURI(u)
	} else {
		return loader.LoadFromFile(filePath)
	}
}

// Deprecated: In kin-openapi v0.126.0 (https://github.com/getkin/kin-openapi/tree/v0.126.0?tab=readme-ov-file#v01260) the Circular Reference Counter functionality was removed, instead resolving all references with backtracking, to avoid needing to provide a limit to reference counts.
//
// This is now identital in method as `LoadOpenAPI`.
func LoadSwaggerWithCircularReferenceCount(filePath string, _ int) (swagger *openapi3.T, err error) {
	return LoadOpenAPI(filePath)
}
