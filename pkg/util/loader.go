package util

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v2"
)

// LoadSwagger loads an OpenAPI spec, and hooks into the kin loader to parse
// version information from the spec.
func LoadSwagger(filePath string) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	// We're using a shim to intercept the loads being done by the kin-openapi
	// loader. We're going to peek inside and try to find version information
	// about the file.
	var ls loaderShim
	loader.ReadFromURIFunc = ls.InterceptLoad

	var openapi *openapi3.T

	u, err := url.Parse(filePath)
	if err == nil && u.Scheme != "" && u.Host != "" {
		// The shim needs a URL to compare with, since it can be called multiple
		// times durin a load when resolving multiple refs.
		ls.srcURL = u
		openapi, err = loader.LoadFromURI(u)
	} else {
		// In the case where the URL failed to parse, we'll construct one explicitly
		// in the same way that Kin does. LoadFromFile simply calls LoadFromURI
		// internally.
		ls.srcURL = &url.URL{Path: filePath}
		openapi, err = loader.LoadFromFile(filePath)
	}

	if err != nil {
		return openapi, err
	}

	// Now, our shim will contain version information. We can return errors here
	// which won't affect loading, but will tell the higher level some information
	// about versions.

	version := ls.versions.OpenAPI
	if version == "" {
		version = ls.versions.Swagger
	}

	// Try to extract the major, minor. Openapi will have patch level, swagger won't. If
	// it doesn't match x.y pattern, we bail out without further checking.
	versionParts := strings.Split(version, ".")
	if len(versionParts) < 2 {
		return openapi, nil
	}
	major := versionParts[0]
	minor := versionParts[1]
	if major == "2" {
		// TODO: we can actually use openapi2conv to convert swagger2 to OpenAPI 3
		return openapi, ErrSwagger2NotSupported
	}
	if major != "3" {
		return openapi, fmt.Errorf("OpenAPI/Swagger %v is not supported", version)
	}
	// Now, we know we've got major 3.
	if minor != "0" {
		return openapi, ErrOpenAPI31NotSupported
	}

	return openapi, nil
}

type loaderShim struct {
	srcURL   *url.URL
	versions openAPIorSwaggerVersion
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
		// We've found our file of interest. Parse it and figure out a version. We'll parse as
		// YAML since this handles JSON too.
		err = yaml.Unmarshal(buf, &l.versions)
		// If we failed to unmarshal we don't react, maintaining previous behavior of
		// trying to process the file.
		if err != nil {
			return buf, nil
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
	return LoadSwagger(filePath)
}
