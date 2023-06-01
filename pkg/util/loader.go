package util

import (
	"errors"
	"net/url"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func ParseOpenAPI(filepath string) (*libopenapi.DocumentModel[v3.Document], error) {
	b, err := os.ReadFile(filepath)
	if err != nil {

		return nil, err
	}
	document, err := libopenapi.NewDocument(b)
	if err != nil {
		return nil, err
	}

	d, errs := document.BuildV3Model()
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return d, nil
}

func LoadSwagger(filePath string) (swagger *libopenapi.DocumentModel[v3.Document], err error) {

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	u, err := url.Parse(filePath)
	if err == nil && u.Scheme != "" && u.Host != "" {
		return loader.LoadFromURI(u)
	} else {
		return loader.LoadFromFile(filePath)
	}
}
