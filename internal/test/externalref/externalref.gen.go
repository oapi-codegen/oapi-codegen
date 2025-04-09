// Package externalref provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.0.0-00010101000000-000000000000 DO NOT EDIT.
package externalref

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	externalRef0 "github.com/oapi-codegen/oapi-codegen/v2/internal/test/externalref/packageA"
	externalRef1 "github.com/oapi-codegen/oapi-codegen/v2/internal/test/externalref/packageB"
	externalRef2 "github.com/oapi-codegen/oapi-codegen/v2/internal/test/externalref/petstore"
)

// Container defines model for Container.
type Container struct {
	ObjectA *externalRef0.ObjectA   `json:"object_a,omitempty"`
	ObjectB *externalRef1.ObjectB   `json:"object_b,omitempty"`
	ObjectC *map[string]interface{} `json:"object_c,omitempty"`
	Pet     *externalRef2.Pet       `json:"pet,omitempty"`
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/6RUTW/bMAz9KwG3o9Bk6LCDb213Xw7bqSgMRmYcbbakSUzWoNB/Hyg1jROnSLJdDJof",
	"D+890n4B7XrvLFmOUL1A1CvqMYcPzjIaS0FefHCeAhvKJbf4SZprlPhjoCVU8GG6B5q+okw96l/Y0l0d",
	"Pen6W566g6R2AIsLAe6HAPcDAH0O4K0vKfDE59rRm3pzWztPVsI5MaSUFBzlH5CpdWE7dsY08qRn7H1H",
	"UH1SsHShR4YKjOUvn0EBbz2VV2opCDGLPR2MwVfXxn1r5GBsC0LkNVNkgYLnvpPJggB6x+sE53lRf0hX",
	"D4Rc4cub/qRGimf/KLlxbWtoLFqBXzl2P0JXDGbqc3DYduzEbmZoGoaA233nn4DeUwMVhzVJW2TkdcZu",
	"KOpgPBtnBYt4UmoTYye8oklkF4Qq2XUP1SPgBk2Hi05ynmxTGEXXNfB0QhBje6jlCuu/Y4G4RFJSEOj3",
	"2gRJPRZrhnY+nbsnn+9/dErC4Z3Lv2L11x43Y+kafvrYNEa2hN18QEbUH6PJHZ38G42EvMPvv39aaU/h",
	"qHQphZQxjF26XDScPxxQsKEQy61mnmVNUMHtzexmJitHXglwSn8DAAD//xIv2MTwBQAA",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var RawSpec = decodeSpecCached()

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
		res[pathToFile] = RawSpec
	}

	for rawPath, rawFunc := range externalRef0.PathToRawSpec(path.Join(path.Dir(pathToFile), "./packageA/spec.yaml")) {
		if _, ok := res[rawPath]; ok {
			// it is not possible to compare functions in golang, so always overwrite the old value
		}
		res[rawPath] = rawFunc
	}
	for rawPath, rawFunc := range externalRef1.PathToRawSpec(path.Join(path.Dir(pathToFile), "./packageB/spec.yaml")) {
		if _, ok := res[rawPath]; ok {
			// it is not possible to compare functions in golang, so always overwrite the old value
		}
		res[rawPath] = rawFunc
	}
	for rawPath, rawFunc := range externalRef2.PathToRawSpec(path.Join(path.Dir(pathToFile), "https://petstore3.swagger.io/api/v3/openapi.json")) {
		if _, ok := res[rawPath]; ok {
			// it is not possible to compare functions in golang, so always overwrite the old value
		}
		res[rawPath] = rawFunc
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
	specData, err = RawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
