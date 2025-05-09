package util

import (
	"bytes"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/speakeasy-api/openapi-overlay/pkg/loader"
	"gopkg.in/yaml.v3"
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

// Deprecated: In kin-openapi v0.126.0 (https://github.com/getkin/kin-openapi/tree/v0.126.0?tab=readme-ov-file#v01260) the Circular Reference Counter functionality was removed, instead resolving all references with backtracking, to avoid needing to provide a limit to reference counts.
//
// This is now identital in method as `LoadSwagger`.
func LoadSwaggerWithCircularReferenceCount(filePath string, _ int) (swagger *openapi3.T, err error) {
	return LoadSwagger(filePath)
}

type LoadSwaggerWithOverlayOpts struct {
	Path   string
	Strict bool
}

func LoadSwaggerWithOverlay(filePath string, opts LoadSwaggerWithOverlayOpts) (swagger *openapi3.T, err error) {
	spec, err := LoadSwagger(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI specification: %w", err)
	}

	if opts.Path == "" {
		return spec, nil
	}

	// parse out the yaml.Node, which is required by the overlay library
	data, err := yaml.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal spec from %#v as YAML: %w", filePath, err)
	}

	var node yaml.Node
	err = yaml.NewDecoder(bytes.NewReader(data)).Decode(&node)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spec from %#v: %w", filePath, err)
	}

	overlay, err := loader.LoadOverlay(opts.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to load Overlay from %#v: %v", opts.Path, err)
	}

	err = overlay.Validate()
	if err != nil {
		return nil, fmt.Errorf("the Overlay in %#v was not valid: %v", opts.Path, err)
	}

	if opts.Strict {
		err, vs := overlay.ApplyToStrict(&node)
		if err != nil {
			return nil, fmt.Errorf("failed to apply Overlay %#v to specification %#v: %v\nAdditionally, the following validation errors were found:\n- %s", opts.Path, filePath, err, strings.Join(vs, "\n- "))
		}
	} else {
		err = overlay.ApplyTo(&node)
		if err != nil {
			return nil, fmt.Errorf("failed to apply Overlay %#v to specification %#v: %v", opts.Path, filePath, err)
		}
	}

	b, err := yaml.Marshal(&node)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize Overlay'd specification %#v: %v", opts.Path, err)
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	swagger, err = loader.LoadFromDataWithPath(b, &url.URL{
		Path: filepath.ToSlash(filePath),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to serialize Overlay'd specification %#v: %v", opts.Path, err)
	}

	return swagger, nil
}
