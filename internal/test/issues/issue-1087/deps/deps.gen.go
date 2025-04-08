// Package deps provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.0.0-00010101000000-000000000000 DO NOT EDIT.
package deps

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"

	"github.com/getkin/kin-openapi/openapi3"
)

// BaseError defines model for BaseError.
type BaseError struct {
	// Code The underlying http status code
	Code int32 `json:"code"`

	// Domain The domain where the error is originating from as defined by the service
	Domain string `json:"domain"`

	// Message A simple message in english describing the error and can be returned to the consumer
	Message string `json:"message"`

	// Metadata Any additional details to be conveyed as determined by the service. If present, will return map of key value pairs
	Metadata *map[string]string `json:"metadata,omitempty"`
}

// Error defines model for Error.
type Error = BaseError

// N401 defines model for 401.
type N401 = Error

// N403 defines model for 403.
type N403 = Error

// N410 defines model for 410.
type N410 = Error

// DefaultError defines model for DefaultError.
type DefaultError = Error

// Base64 encoded, compressed with deflate, json marshaled Swagger object
const swaggerSpec = "" +
	"tJXfj+M0EMf/lZHhMWrT6yKhvBUO0CIBpwOeTquVG0+aOZyxz550t6z6vyM7adMfq1vupH1qE894vvOZ" +
	"H3lSteu8Y2SJqnpSAaN3HDE/3JSL9FM7FmRJf7X3lmot5Hj+MTpO7/BRd97iYGlQVTflolDGdZpYVWrV" +
	"S6sK1WGMeoOqUre81ZYM6F5aZBmvU4UKqPOVB4v71bnFvlCxbrHTKdS3ARtVqW/mk/75cBrnP4Xggtrv" +
	"CyX4KHNvs5KnE++jZvXrgwBFwEdPAY0qlOx8eh8lEG/UPt1iMNaBfBZRqb85KXeB/kUzg1s2SR9GkFYL" +
	"+OC2ZNBAHdAk7drGdD87gZxUyuKmXH4d1+XnuK7qGmOEt8iUEzniHA7ux4NXoXgZ+0WIP7uwJmOQrwgG" +
	"/NRjFKg1J2hrhAl3hrcovwreojyBdwbuF8d4yis/vwqmMdKLdN5j7YIBdmAdbzCA3mqyem2zrrfY6N7K" +
	"EPhlFF+WxpWWMRpgMoDDfoDGBdC8G15HIAZpEVbvbjOK8dYU9Acd8SjVB+cxCA37ZajM00XAv1qEng0G" +
	"uyPeQCviIYqWPkJ2KCag35VloRoXOi2qUsSyfDPhJRbcYEjEDnV/LtRwBg8tBsw5DIlSBBdoQ6wlqWiC" +
	"60BHMNgQo4H1LttGDFuqzzSp6wKfNNulghVESn4wWiSQyBtLsYXBcp3CT7o0mzQaaS4CSh+SGHHZoHYc" +
	"+w7DmZrVBk9GyaYplVYzLL5/XqdooyV3izaGkkpt351V7crpIiPeweQKBkWTjUnjOkvc4g7NgFIwdM/Q" +
	"nMFtAz5gRJYCHsjaMVXotAfXwD+4S8u0R/CaQjzN99hiu991l2SePqaipvWSN/3+mL5bf8Ract8eT6sP" +
	"amy2sXemGt5dORbq2ODa2j8aVX34/LBNM7G/Ky6G4rCGrjtlOMlDANFjTQ3Vh9qP6E7bo49Da1D+DjUD" +
	"YnzUdfrgxR5n8GfremuyLdOnHuGBpCUGDcekp0Z6n6Pf/zhQuVxh/w/dccleM0xXEDcudxhJDvmbM2hT" +
	"ebcY4kDhzayclYm488jak6rUclbOlqpQXkubCKb1gyG55Dr0wapKzdX+bv9fAAAA//8="

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	compressed, err := base64.StdEncoding.DecodeString(swaggerSpec)
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr := flate.NewReader(bytes.NewReader(compressed))
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(zr); err != nil {
		return nil, fmt.Errorf("read flate: %w", err)
	}
	if err := zr.Close(); err != nil {
		return nil, fmt.Errorf("close flate reader: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
