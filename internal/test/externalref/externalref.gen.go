// Package externalref provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen/v2 version v2.0.0-00010101000000-000000000000 DO NOT EDIT.
package externalref

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"

	externalRef0 "github.com/deepmap/oapi-codegen/v2/internal/test/externalref/packageA"
	externalRef1 "github.com/deepmap/oapi-codegen/v2/internal/test/externalref/packageB"
	externalRef2 "github.com/deepmap/oapi-codegen/v2/internal/test/externalref/petstore"
	"github.com/getkin/kin-openapi/openapi3"
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

	"H4sIAAAAAAAC/5xUwW7bMAz9lYDbUWgybNjBt7W7L4ftVBQDYzOOBlvSKKZrUOjfB0pN7MRbk+YSKPJ7",
	"5HvUk56h9n3wjpxEqJ4h1hvqMS/vUKj1vNN1YB+IxVL+Yhv9pSfsQ0dQfTCw9tyjQAXWyedPYEB2gcpf",
	"aokhGXDY0xENvvo2DtAobF0LKR12/OoX1QIGnvpOmaUC1HtdCr3zTtA64qnKQv+Jun7PtIYK3s0Ht/MX",
	"q/NvGfdFNb5QVpdRbkeU+hzlgEsGAsk5+JIEkhrcq5vY28/zZHxXmBja3F7cRjnLYuMYX49S81r3Q7qS",
	"meRpcWWgGt+2lqaRMhA2XvwP7kp8hfo49XSasz1nHElkxt2A/MMYAjVQCW9JYVFQtrl2Q7FmG8R6p7VI",
	"ZuXbzLqZbGgWxbNKJbftoboHfETb4arTvUCuKYqi7xp4+IchwfbYy2uz/o6Fc4mHZIDp99aybt2XWYzn",
	"93DueoacXAPa9D8vxxsO962Pg2BBja8lNo3Vc8BuORKjdk+rZfvWrX1ubSWnCgw8EsdykPmCBXIYLFTw",
	"8WZxs9DxoGzUX0p/AwAA//++YR/sUAUAAA==",
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

	pathPrefix := path.Dir(pathToFile)

	for rawPath, rawFunc := range externalRef0.PathToRawSpec(path.Join(pathPrefix, "./packageA/spec.yaml")) {
		if _, ok := res[rawPath]; ok {
			// it is not possible to compare functions in golang, so always overwrite the old value
		}
		res[rawPath] = rawFunc
	}
	for rawPath, rawFunc := range externalRef1.PathToRawSpec(path.Join(pathPrefix, "./packageB/spec.yaml")) {
		if _, ok := res[rawPath]; ok {
			// it is not possible to compare functions in golang, so always overwrite the old value
		}
		res[rawPath] = rawFunc
	}
	for rawPath, rawFunc := range externalRef2.PathToRawSpec(path.Join(pathPrefix, "https://petstore3.swagger.io/api/v3/openapi.json")) {
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
