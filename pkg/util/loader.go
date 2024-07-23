package util

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v2"
)

// Deprecated: LoadSwagger loads an OpenAPI 3.0 definition from a file or a
// URL. Use LoadOpenAPI instead.
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

// LoadOpenAPI loads an OpenAPI spec, and hooks into the kin loader to parse
// version information from the spec.
func LoadOpenAPI(filePath string) (openapi *openapi3.T, err error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	// We're using a shim to intercept the loads being done by the kin-openapi
	// loader. We're going to peek inside and try to find version information
	// about the file.
	var ls loaderShim
	loader.ReadFromURIFunc = ls.InterceptLoad

	u, err := url.Parse(filePath)
	if err == nil && u.Scheme != "" && u.Host != "" {
		// The shim needs a URL to compare with, since it can be called multiple
		// times durin a load when resolving multiple refs.
		ls.srcURL = u
		return loader.LoadFromURI(u)
	} else {
		// In the case where the URL failed to parse, we'll construct one explicitly
		// in the same way that Kin does. LoadFromFile simply calls LoadFromURI
		// internally.
		ls.srcURL = &url.URL{Path: filePath}
		return loader.LoadFromFile(filePath)
	}
}

type loaderShim struct {
	srcURL *url.URL
}

// openAPIorSwaggerVersion is used to parse either OpenAPI or Swagger version
// from a JSON or YAML file.
type openAPIorSwaggerVersion struct {
	Swagger string `json:"swagger" yaml:"swagger"`
	OpenAPI string `json:"openapi" yaml:"openapi"`
}

var ErrSwagger2NotSupported = errors.New("swagger version 2.0 is not supported")
var ErrOpenAPI31NotSupported = errors.New("OpenAPI version 3.1 is not yet supported")

func (l *loaderShim) InterceptLoad(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
	buf, err := openapi3.DefaultReadFromURI(loader, url)
	if err != nil {
		return buf, err
	}

	if l.srcURL.Scheme == url.Scheme && l.srcURL.Host == url.Host && l.srcURL.Path == url.Path {
		var versionInfo openAPIorSwaggerVersion
		// We've found our file of interest. Parse it and figure out a version. We'll parse as
		// YAML since this handles JSON too.
		err = yaml.Unmarshal(buf, &versionInfo)
		// If we failed to unmarshal we don't react, maintaining previous behavior of
		// trying to process the file.
		if err != nil {
			return buf, nil
		}

		version := versionInfo.OpenAPI
		if version == "" {
			version = versionInfo.Swagger
		}

		// Try to extract the major, minor. Openapi will have patch level, swagger won't
		versionParts := strings.Split(version, ".")
		if len(versionParts) < 2 {
			return buf, nil
		}
		major := versionParts[0]
		minor := versionParts[1]
		if major == "2" {
			// TODO: we can actually use openapi2conv to convert swagger2 to OpenAPI 3
			return nil, ErrSwagger2NotSupported
		}
		if major != "3" {
			return nil, fmt.Errorf("OpenAPI/Swagger %v is not supported", version)
		}
		// Now, we know we've got major 3.
		if minor != "0" {
			return nil, ErrOpenAPI31NotSupported
		}
	}

	return buf, nil
}

// Deprecated: In kin-openapi v0.126.0 (https://github.com/getkin/kin-openapi/tree/v0.126.0?tab=readme-ov-file#v01260) the
// Circular Reference Counter functionality was removed, instead resolving all references with backtracking, to avoid
// needing to provide a limit to reference counts.
//
// This is now identical in method as `LoadSwagger`.
func LoadSwaggerWithCircularReferenceCount(filePath string, _ int) (swagger *openapi3.T, err error) {
	return LoadOpenAPI(filePath)
}
