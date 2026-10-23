package util

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const remoteDoc = `openapi: "3.0.0"
info:
  title: remote
  version: "1.0.0"
paths: {}
components:
  schemas:
    Remote:
      type: string
`

// fakeTransport serves remoteDoc for every request, or the document in docs
// for a URL listed there, and records the URLs it was asked for, so the tests
// need no listening socket.
type fakeTransport struct {
	docs      map[string]string
	requested []string
}

func (f *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	f.requested = append(f.requested, req.URL.String())
	doc, ok := f.docs[req.URL.String()]
	if !ok {
		doc = remoteDoc
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(doc)),
		Header:     http.Header{"Content-Type": []string{"application/yaml"}},
		Request:    req,
	}, nil
}

func writeSpec(t *testing.T, dir, name, ref string) string {
	t.Helper()
	spec := `openapi: "3.0.0"
info:
  title: local
  version: "1.0.0"
paths: {}
components:
  schemas:
    Local:
      type: object
      properties:
        remote:
          $ref: "` + ref + `"
`
	p := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(p, []byte(spec), 0o600))
	return p
}

func TestLoadSwagger_RemoteRefs(t *testing.T) {
	dir := t.TempDir()
	const doc = "http://schemas.example.com:8080/remote.yaml"
	spec := writeSpec(t, dir, "spec.yaml", doc+"#/components/schemas/Remote")

	cases := []struct {
		name    string
		allowed []string
		refused bool
	}{
		{name: "refused by default", allowed: nil, refused: true},
		{name: "refused with empty list", allowed: []string{}, refused: true},
		{name: "allowed by exact document URL", allowed: []string{doc}},
		{name: "allowed with surrounding whitespace", allowed: []string{" " + doc + " "}},
		{name: "allowed by wildcard", allowed: []string{"*"}},
		{name: "allowed among others", allowed: []string{"https://other.example.com/x.yaml", doc}},
		{name: "host alone is not enough", allowed: []string{"schemas.example.com"}, refused: true},
		{name: "host and port alone is not enough", allowed: []string{"http://schemas.example.com:8080"}, refused: true},
		{name: "other path refused", allowed: []string{"http://schemas.example.com:8080/other.yaml"}, refused: true},
		{name: "other port refused", allowed: []string{"http://schemas.example.com:9090/remote.yaml"}, refused: true},
		{name: "other scheme refused", allowed: []string{"https://schemas.example.com:8080/remote.yaml"}, refused: true},
		{name: "prefix refused", allowed: []string{"http://schemas.example.com:8080/"}, refused: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			transport := &fakeTransport{}
			swagger, err := loadSwagger(spec, tc.allowed, &http.Client{Transport: transport})
			if tc.refused {
				require.ErrorContains(t, err, `refusing to resolve remote reference "`+doc+`"`)
				assert.Empty(t, transport.requested, "no request must leave the process for a refused reference")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, []string{doc}, transport.requested)
			remote := swagger.Components.Schemas["Local"].Value.Properties["remote"]
			require.NotNil(t, remote.Value)
			assert.Equal(t, "string", remote.Value.Type.Slice()[0])
		})
	}
}

func TestLoadSwagger_RemoteRefsRefusedThroughPublicAPI(t *testing.T) {
	// LoadSwagger and LoadSwaggerWithOverlay use http.DefaultClient; the
	// refusal happens before any request is built, so this needs no network.
	dir := t.TempDir()
	spec := writeSpec(t, dir, "spec.yaml", "http://169.254.169.254/latest/meta-data/#/x")

	_, err := LoadSwagger(spec)
	require.ErrorContains(t, err, `refusing to resolve remote reference "http://169.254.169.254/latest/meta-data/"`)
	assert.ErrorContains(t, err, "is not listed in allow-remote-refs-from")

	_, err = LoadSwaggerWithOverlay(spec, LoadSwaggerWithOverlayOpts{})
	require.ErrorContains(t, err, "refusing to resolve remote reference")
}

func TestLoadSwagger_SpecFromURL(t *testing.T) {
	const top = "https://specs.example.com/v1/api.yaml"
	const sibling = "https://specs.example.com/v1/common.yaml"
	topDoc := func(ref string) string {
		return `openapi: "3.0.0"
info:
  title: top
  version: "1.0.0"
paths: {}
components:
  schemas:
    Top:
      type: object
      properties:
        remote:
          $ref: "` + ref + `"
`
	}

	t.Run("the document itself loads without being listed", func(t *testing.T) {
		transport := &fakeTransport{docs: map[string]string{top: remoteDoc}}
		_, err := loadSwagger(top, nil, &http.Client{Transport: transport})
		require.NoError(t, err)
		assert.Equal(t, []string{top}, transport.requested)
	})

	t.Run("a fragment on the document URL does not stop it loading", func(t *testing.T) {
		transport := &fakeTransport{docs: map[string]string{top: remoteDoc}}
		_, err := loadSwagger(top+"#overview", nil, &http.Client{Transport: transport})
		require.NoError(t, err)
		assert.Equal(t, []string{top}, transport.requested)
	})

	t.Run("a relative reference to the same host is refused until its absolute URL is listed", func(t *testing.T) {
		transport := &fakeTransport{docs: map[string]string{top: topDoc("common.yaml#/components/schemas/Remote")}}
		_, err := loadSwagger(top, nil, &http.Client{Transport: transport})
		require.ErrorContains(t, err, `refusing to resolve remote reference "`+sibling+`"`)
		assert.Equal(t, []string{top}, transport.requested)

		transport = &fakeTransport{docs: map[string]string{top: topDoc("common.yaml#/components/schemas/Remote")}}
		swagger, err := loadSwagger(top, []string{sibling}, &http.Client{Transport: transport})
		require.NoError(t, err)
		assert.Equal(t, []string{top, sibling}, transport.requested)
		remote := swagger.Components.Schemas["Top"].Value.Properties["remote"]
		require.NotNil(t, remote.Value)
		assert.Equal(t, "string", remote.Value.Type.Slice()[0])
	})
}

func TestLoadSwaggerWithOverlay_RemoteRefs(t *testing.T) {
	// The document is loaded twice when an overlay is applied; both loads
	// must apply the same policy, including the document's own URL.
	const top = "https://specs.example.com/v1/api.yaml"
	const sibling = "https://specs.example.com/v1/common.yaml"
	topDoc := `openapi: "3.0.0"
info:
  title: top
  version: "1.0.0"
paths: {}
components:
  schemas:
    Top:
      type: object
      properties:
        remote:
          $ref: "common.yaml#/components/schemas/Remote"
`
	overlay := filepath.Join(t.TempDir(), "overlay.yaml")
	require.NoError(t, os.WriteFile(overlay, []byte(`overlay: 1.0.0
info:
  title: add a description
  version: 1.0.0
actions:
  - target: $.info
    update:
      description: overlaid
`), 0o600))

	transport := &fakeTransport{docs: map[string]string{top: topDoc}}
	_, err := loadSwaggerWithOverlay(top, LoadSwaggerWithOverlayOpts{Path: overlay, Strict: true}, &http.Client{Transport: transport})
	require.ErrorContains(t, err, `refusing to resolve remote reference "`+sibling+`"`)

	transport = &fakeTransport{docs: map[string]string{top: topDoc}}
	swagger, err := loadSwaggerWithOverlay(top, LoadSwaggerWithOverlayOpts{Path: overlay, Strict: true, AllowRemoteRefsFrom: []string{sibling}}, &http.Client{Transport: transport})
	require.NoError(t, err)
	assert.Equal(t, "overlaid", swagger.Info.Description)
	assert.Equal(t, []string{top, sibling, sibling}, transport.requested, "the overlaid document is loaded again and resolves the sibling again")
}

func TestLoadSwagger_LocalFileRefsStillFollowed(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "remote.yaml"), []byte(remoteDoc), 0o600))
	spec := writeSpec(t, dir, "spec.yaml", "./remote.yaml#/components/schemas/Remote")

	transport := &fakeTransport{}
	swagger, err := loadSwagger(spec, nil, &http.Client{Transport: transport})
	require.NoError(t, err)
	assert.Empty(t, transport.requested)
	remote := swagger.Components.Schemas["Local"].Value.Properties["remote"]
	require.NotNil(t, remote.Value)
	assert.Equal(t, "string", remote.Value.Type.Slice()[0])
}

func TestRemoteRefAllowed(t *testing.T) {
	loc := func(s string) *url.URL {
		u, err := url.Parse(s)
		require.NoError(t, err)
		return u
	}
	assert.False(t, RemoteRefAllowed(nil, loc("http://a.example/x.yaml")))
	assert.True(t, RemoteRefAllowed([]string{"*"}, loc("http://a.example/x.yaml")))
	assert.True(t, RemoteRefAllowed([]string{"http://a.example/x.yaml"}, loc("http://a.example/x.yaml")))
	assert.True(t, RemoteRefAllowed([]string{"http://a.example/x.yaml"}, loc("http://a.example/x.yaml#/components/schemas/X")), "the fragment names a part of the document, not another document")
	assert.True(t, RemoteRefAllowed([]string{"http://[::1]:8080/x.yaml"}, loc("http://[::1]:8080/x.yaml")))
	assert.False(t, RemoteRefAllowed([]string{"http://a.example/x.yaml/"}, loc("http://a.example/x.yaml")))
	assert.False(t, RemoteRefAllowed([]string{"http://A.example/x.yaml"}, loc("http://a.example/x.yaml")), "nothing is normalised")
	assert.False(t, RemoteRefAllowed([]string{"a.example"}, loc("http://a.example/x.yaml")))
}
