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
// a URL, and resolves its external references, including references to remote
// URLs. Use LoadSwaggerWithOverlay with AllowRemoteRefsFrom to restrict which
// remote documents may be fetched.
func LoadSwagger(filePath string) (swagger *openapi3.T, err error) {
	return loadSwagger(filePath, nil, http.DefaultClient)
}

func loadSwagger(filePath string, allowRemoteRefsFrom []string, client *http.Client) (*openapi3.T, error) {
	if u := specURL(filePath); u != nil {
		return newLoader(filePath, allowRemoteRefsFrom, client).LoadFromURI(u)
	}
	return newLoader(filePath, allowRemoteRefsFrom, client).LoadFromFile(filePath)
}

// specURL returns filePath parsed as a URL when it is one with a scheme and a
// host, and nil when it is a local path.
func specURL(filePath string) *url.URL {
	u, err := url.Parse(filePath)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil
	}
	return u
}

// newLoader returns a kin-openapi loader for the document at filePath that
// follows references to local files and to the remote URLs listed in
// allowRemoteRefsFrom. When filePath is itself a URL, the document is fetched
// through the same reader, so that one URL is allowed as well: the caller
// named it.
func newLoader(filePath string, allowRemoteRefsFrom []string, client *http.Client) *openapi3.Loader {
	if allowRemoteRefsFrom == nil {
		// On the 2.8 release line an unset list keeps the behaviour of earlier
		// releases, which followed every remote reference. An explicit empty
		// list refuses them all. This lives here, not in loadSwagger, so that
		// the loader built after an overlay is applied gets the same default.
		allowRemoteRefsFrom = []string{"*"}
	}
	if u := specURL(filePath); u != nil {
		// Listed like any other entry: the document URL without a fragment,
		// which is what the reader compares against.
		allowRemoteRefsFrom = append([]string{remoteRefKey(u)}, allowRemoteRefsFrom...)
	}
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
// URLs that are not allowed and fetches the rest over HTTP. Locations without
// a host are left to the next reader.
func remoteRefReader(allowed []string, client *http.Client) openapi3.ReadFromURIFunc {
	readHTTP := openapi3.ReadFromHTTP(client)
	return func(loader *openapi3.Loader, location *url.URL) ([]byte, error) {
		if location.Scheme == "" || location.Host == "" {
			return nil, openapi3.ErrURINotSupported
		}
		if !RemoteRefAllowed(allowed, location) {
			return nil, fmt.Errorf("refusing to resolve remote reference %q: it is not listed in allow-remote-refs-from", remoteRefKey(location))
		}
		return readHTTP(loader, location)
	}
}

// RemoteRefAllowed reports whether a $ref to location may be fetched given the
// allow-remote-refs-from list. An entry of "*" allows every URL. Any other
// entry must equal the URL of the referenced document exactly, that is the
// $ref with its fragment removed. A relative $ref inside a remote document
// resolves to an absolute URL before it gets here, and that absolute URL is
// what has to be listed. Nothing is normalised on purpose.
func RemoteRefAllowed(allowed []string, location *url.URL) bool {
	key := remoteRefKey(location)
	for _, entry := range allowed {
		entry = strings.TrimSpace(entry)
		if entry == "*" || entry == key {
			return true
		}
	}
	return false
}

// remoteRefKey is the URL of the document a $ref points at: the location
// without its fragment.
func remoteRefKey(location *url.URL) string {
	doc := *location
	doc.Fragment = ""
	doc.RawFragment = ""
	return doc.String()
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
	// AllowRemoteRefsFrom lists the remote document URLs that external $refs
	// may be fetched from, see RemoteRefAllowed. nil allows every URL, as
	// earlier 2.8 releases did; an empty, non-nil list refuses every remote
	// reference.
	AllowRemoteRefsFrom []string
}

func LoadSwaggerWithOverlay(filePath string, opts LoadSwaggerWithOverlayOpts) (swagger *openapi3.T, err error) {
	return loadSwaggerWithOverlay(filePath, opts, http.DefaultClient)
}

func loadSwaggerWithOverlay(filePath string, opts LoadSwaggerWithOverlayOpts, client *http.Client) (swagger *openapi3.T, err error) {
	spec, err := loadSwagger(filePath, opts.AllowRemoteRefsFrom, client)
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

	// The same policy as the first load, including the document's own URL,
	// and the same location, so that relative references in the overlaid
	// document resolve where they did in the original one.
	loader := newLoader(filePath, opts.AllowRemoteRefsFrom, client)
	location := specURL(filePath)
	if location == nil {
		location = &url.URL{Path: filepath.ToSlash(filePath)}
	}

	swagger, err = loader.LoadFromDataWithPath(b, location)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize Overlay'd specification %#v: %v", opts.Path, err)
	}

	return swagger, nil
}
