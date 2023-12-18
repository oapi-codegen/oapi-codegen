package util

import (
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
)

func LoadSwagger(filePath string) (swagger *openapi3.T, err error) {

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	u, err := url.Parse(filePath)
	if err == nil && u.Scheme != "" && u.Host != "" {
		return loader.LoadFromURI(u)
	} else {
		return loader.LoadFromFile(filePath)
	}
}

func LoadSwaggerWithCircularReferenceCount(filePath string, circularReferenceCount int) (swagger *openapi3.T, err error) {
	// get a copy of the existing count
	existingCircularReferenceCount := openapi3.CircularReferenceCounter
	if circularReferenceCount > 0 {
		openapi3.CircularReferenceCounter = circularReferenceCount
	}

	swagger, err = LoadSwagger(filePath)

	if circularReferenceCount > 0 {
		// and make sure to reset it
		openapi3.CircularReferenceCounter = existingCircularReferenceCount
	}

	return swagger, err
}
