package util

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

func LoadSwagger(filePath string, allowRefs bool, insecure bool, clearRefs bool) (*openapi3.Swagger, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var swagger *openapi3.Swagger
	ext := filepath.Ext(filePath)
	ext = strings.ToLower(ext)
	switch ext {
	case ".yaml", ".yml":
		sl := openapi3.NewSwaggerLoader(openapi3.WithClearResolvedRefs(clearRefs))
		sl.IsExternalRefsAllowed = allowRefs
		if insecure {
			sl.LoadSwaggerFromURIFunc = insecureReadUrl
		}
		swagger, err = sl.LoadSwaggerFromFile(filePath)
	case ".json":
		swagger = &openapi3.Swagger{}
		err = json.Unmarshal(data, swagger)
	default:
		return nil, fmt.Errorf("%s is not a supported extension, use .yaml, .yml or .json", ext)
	}
	if err != nil {
		return nil, err
	}
	return swagger, nil
}

func LoadSwaggerFromURL(url *url.URL, allowRefs, insecure bool) (*openapi3.Swagger, error) {
	var err error
	var swagger *openapi3.Swagger

	ext := filepath.Ext(url.Path)
	ext = strings.ToLower(ext)
	switch ext {
	case ".yaml", ".yml":
		sl := openapi3.NewSwaggerLoader()
		sl.IsExternalRefsAllowed = allowRefs
		if insecure {
			sl.LoadSwaggerFromURIFunc = insecureReadUrl
		}
		swagger, err = sl.LoadSwaggerFromURI(url)

	// enable support for loading remote JSON files later
	//case ".json":
	//	swagger = &openapi3.Swagger{}
	//	err = json.Unmarshal(data, swagger)
	default:
		return nil, fmt.Errorf("%s is not a supported extension, use .yaml, .yml when specifying remote schemas", ext)
	}
	if err != nil {
		return nil, err
	}
	return swagger, nil
}

func insecureReadUrl(sl *openapi3.SwaggerLoader, location *url.URL) (*openapi3.Swagger, error) {
	if location.Scheme != "" && location.Host != "" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		cl := &http.Client{Transport: tr}
		resp, err := cl.Get(location.String())
		if err != nil {
			return nil, err
		}
		data, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return nil, err
		}
		return sl.LoadSwaggerFromDataWithPath(data, location)
	}
	if location.Scheme != "" || location.Host != "" || location.RawQuery != "" {
		return nil, fmt.Errorf("Unsupported URI: '%s'", location.String())
	}
	data, err := ioutil.ReadFile(location.Path)
	if err != nil {
		return nil, err
	}
	return sl.LoadSwaggerFromDataWithPath(data, location)
}
