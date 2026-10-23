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
	spec := writeSpec(t, dir, "spec.yaml", "http://schemas.example.com:8080/remote.yaml#/components/schemas/Remote")

	cases := []struct {
		name    string
		allowed []string
		refused bool
	}{
		{name: "refused by default", allowed: nil, refused: true},
		{name: "refused with empty list", allowed: []string{}, refused: true},
		{name: "allowed by host", allowed: []string{"schemas.example.com"}},
		{name: "allowed by host and port", allowed: []string{"schemas.example.com:8080"}},
		{name: "allowed case-insensitively", allowed: []string{"Schemas.Example.COM"}},
		{name: "allowed by wildcard", allowed: []string{"*"}},
		{name: "other host refused", allowed: []string{"other.example.com"}, refused: true},
		{name: "other port refused", allowed: []string{"schemas.example.com:9090"}, refused: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			transport := &fakeTransport{}
			swagger, err := loadSwagger(spec, tc.allowed, &http.Client{Transport: transport})
			if tc.refused {
				require.ErrorContains(t, err, `refusing to resolve remote reference "http://schemas.example.com:8080/remote.yaml"`)
				assert.Empty(t, transport.requested, "no request must leave the process for a refused reference")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, []string{"http://schemas.example.com:8080/remote.yaml"}, transport.requested)
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
	require.ErrorContains(t, err, "refusing to resolve remote reference")
	assert.ErrorContains(t, err, `host "169.254.169.254" is not listed in allow-remote-refs-from`)

	_, err = LoadSwaggerWithOverlay(spec, LoadSwaggerWithOverlayOpts{})
	require.ErrorContains(t, err, "refusing to resolve remote reference")
}

func TestLoadSwagger_SpecFromURL(t *testing.T) {
	const top = "https://specs.example.com/v1/api.yaml"
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

	t.Run("relative reference follows the document's own host", func(t *testing.T) {
		transport := &fakeTransport{docs: map[string]string{top: topDoc("common.yaml#/components/schemas/Remote")}}
		swagger, err := loadSwagger(top, nil, &http.Client{Transport: transport})
		require.NoError(t, err)
		assert.Equal(t, []string{top, "https://specs.example.com/v1/common.yaml"}, transport.requested)
		remote := swagger.Components.Schemas["Top"].Value.Properties["remote"]
		require.NotNil(t, remote.Value)
		assert.Equal(t, "string", remote.Value.Type.Slice()[0])
	})

	t.Run("reference to another host is refused", func(t *testing.T) {
		transport := &fakeTransport{docs: map[string]string{top: topDoc("https://other.example.com/common.yaml#/components/schemas/Remote")}}
		_, err := loadSwagger(top, nil, &http.Client{Transport: transport})
		require.ErrorContains(t, err, `refusing to resolve remote reference "https://other.example.com/common.yaml"`)
		assert.Equal(t, []string{top}, transport.requested)
	})

	t.Run("reference to another host is followed when listed", func(t *testing.T) {
		transport := &fakeTransport{docs: map[string]string{top: topDoc("https://other.example.com/common.yaml#/components/schemas/Remote")}}
		_, err := loadSwagger(top, []string{"other.example.com"}, &http.Client{Transport: transport})
		require.NoError(t, err)
		assert.Equal(t, []string{top, "https://other.example.com/common.yaml"}, transport.requested)
	})
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

func TestRemoteRefHostAllowed(t *testing.T) {
	loc := func(s string) *url.URL {
		u, err := url.Parse(s)
		require.NoError(t, err)
		return u
	}
	assert.False(t, RemoteRefHostAllowed(nil, loc("http://a.example")))
	assert.True(t, RemoteRefHostAllowed([]string{"*"}, loc("http://a.example")))
	assert.True(t, RemoteRefHostAllowed([]string{" a.example "}, loc("http://a.example:443")))
	assert.True(t, RemoteRefHostAllowed([]string{"a.example:443"}, loc("http://a.example:443")))
	assert.False(t, RemoteRefHostAllowed([]string{"a.example:80"}, loc("http://a.example:443")))
	assert.False(t, RemoteRefHostAllowed([]string{"example"}, loc("http://a.example")))
}
