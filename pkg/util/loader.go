package util

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/speakeasy-api/openapi/overlay/loader"
	"gopkg.in/yaml.v3"
)

// LoadSwagger loads the OpenAPI document at filePath, which is a local path or
// a URL, and resolves its external references. References to other local
// files are followed; references to remote URLs are refused, except those on
// the host the document itself was loaded from when filePath is a URL. Use
// LoadSwaggerWithOverlay with AllowRemoteRefsFrom to follow remote references.
func LoadSwagger(filePath string) (swagger *openapi3.T, err error) {
	return loadSwagger(filePath, nil, http.DefaultClient)
}

func loadSwagger(filePath string, allowRemoteRefsFrom []string, client *http.Client) (*openapi3.T, error) {
	u, err := url.Parse(filePath)
	if err == nil && u.Scheme != "" && u.Host != "" {
		// The document's own host is the caller's choice; references relative
		// to it resolve there as well.
		loader := newLoader(append([]string{u.Host}, allowRemoteRefsFrom...), client)
		return loader.LoadFromURI(u)
	}
	return newLoader(allowRemoteRefsFrom, client).LoadFromFile(filePath)
}

// newLoader returns a kin-openapi loader that follows references to local
// files and to remote hosts listed in allowRemoteRefsFrom.
func newLoader(allowRemoteRefsFrom []string, client *http.Client) *openapi3.Loader {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	// Record each element's source location so route registration can be
	// emitted in the order paths are declared in the spec (issue #1887).
	loader.IncludeOrigin = true
	// A custom ReadFromURIFunc owns the access policy: kin-openapi's default
	// reader fetches any URL and reads any local path a $ref names, which lets
	// an untrusted document make the generator issue requests to arbitrary
	// hosts. Remote references go through the allow list first; local files
	// keep working so that import-mapping across files is unaffected.
	loader.ReadFromURIFunc = openapi3.URIMapCache(openapi3.ReadFromURIs(
		remoteRefReader(allowRemoteRefsFrom, client),
		openapi3.ReadFromFile,
	))
	return loader
}

// remoteRefReader returns a reader for references that name a host. It refuses
// hosts that are not allowed and fetches the rest over HTTP. Locations without
// a host are left to the next reader.
func remoteRefReader(allowed []string, client *http.Client) openapi3.ReadFromURIFunc {
	readHTTP := openapi3.ReadFromHTTP(client)
	return func(loader *openapi3.Loader, location *url.URL) ([]byte, error) {
		if location.Scheme == "" || location.Host == "" {
			return nil, openapi3.ErrURINotSupported
		}
		if !RemoteRefHostAllowed(allowed, location) {
			return nil, fmt.Errorf("refusing to resolve remote reference %q: host %q is not listed in allow-remote-refs-from", location.String(), location.Host)
		}
		return readHTTP(loader, location)
	}
}

// RemoteRefHostAllowed reports whether a $ref to location may be fetched given
// the allow-remote-refs-from list. An entry of "*" allows every host. Any
// other entry is a host name, optionally with a port, compared
// case-insensitively: an entry without a port matches that host on any port,
// an entry with a port matches only that host and port.
func RemoteRefHostAllowed(allowed []string, location *url.URL) bool {
	for _, entry := range allowed {
		entry = strings.TrimSpace(entry)
		if entry == "*" {
			return true
		}
		if strings.EqualFold(entry, location.Host) || strings.EqualFold(entry, location.Hostname()) {
			return true
		}
	}
	return false
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
	// AllowRemoteRefsFrom lists the hosts that external $refs may be fetched
	// from, see RemoteRefHostAllowed. Empty refuses every remote reference
	// (other than to the host the document itself is loaded from).
	AllowRemoteRefsFrom []string
}

func LoadSwaggerWithOverlay(filePath string, opts LoadSwaggerWithOverlayOpts) (swagger *openapi3.T, err error) {
	spec, err := loadSwagger(filePath, opts.AllowRemoteRefsFrom, http.DefaultClient)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI specification: %w", err)
	}

	if opts.Path == "" {
		return spec, nil
	}

	// parse out the yaml.Node, which is required by the overlay library
	buf := &bytes.Buffer{}
	enc := yaml.NewEncoder(buf)
	// set to 2 to work around https://github.com/yaml/go-yaml/issues/76
	enc.SetIndent(2)
	err = enc.Encode(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal spec from %#v as YAML: %w", filePath, err)
	}

	var node yaml.Node
	err = yaml.NewDecoder(buf).Decode(&node)
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
		vs, err := overlay.ApplyToStrict(&node)
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

	loader := newLoader(opts.AllowRemoteRefsFrom, http.DefaultClient)

	swagger, err = loader.LoadFromDataWithPath(b, &url.URL{
		Path: filepath.ToSlash(filePath),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to serialize Overlay'd specification %#v: %v", opts.Path, err)
	}

	return swagger, nil
}
