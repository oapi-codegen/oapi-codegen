package bodiesrequests

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOptionalClientRequestBodies(t *testing.T) {
	const server = "https://example.com"
	params := &Issue9Params{Foo: "filter"}
	var nullJSON Issue9JSONRequestBody
	var emptyJSON Issue9JSONRequestBody = map[string]string{}
	form := Issue9FormdataRequestBody{Name: "a b"}
	emptyText := ""
	text := "hello"

	for _, tc := range []struct {
		name        string
		build       func() (*http.Request, error)
		contentType string
		body        string
	}{
		{"absent JSON", func() (*http.Request, error) { return NewIssue9Request(server, params, nil) }, "", ""},
		{"JSON null", func() (*http.Request, error) { return NewIssue9Request(server, params, &nullJSON) }, "application/json", "null"},
		{"empty JSON object", func() (*http.Request, error) { return NewIssue9Request(server, params, &emptyJSON) }, "application/json", "{}"},
		{"absent form", func() (*http.Request, error) { return NewIssue9RequestWithFormdataBody(server, params, nil) }, "", ""},
		{"form", func() (*http.Request, error) { return NewIssue9RequestWithFormdataBody(server, params, &form) }, "application/x-www-form-urlencoded", "name=a+b"},
		{"absent text", func() (*http.Request, error) { return NewIssue9RequestWithTextBody(server, params, nil) }, "", ""},
		{"empty text", func() (*http.Request, error) { return NewIssue9RequestWithTextBody(server, params, &emptyText) }, "text/plain", ""},
		{"text", func() (*http.Request, error) { return NewIssue9RequestWithTextBody(server, params, &text) }, "text/plain", "hello"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, err := tc.build()
			require.NoError(t, err)
			require.Equal(t, "filter", req.URL.Query().Get("foo"))
			if tc.contentType == "" {
				require.Nil(t, req.Body)
				require.NotContains(t, req.Header, "Content-Type")
				return
			}
			require.Equal(t, tc.contentType, req.Header.Get("Content-Type"))
			data, err := io.ReadAll(req.Body)
			require.NoError(t, req.Body.Close())
			require.NoError(t, err)
			require.Equal(t, tc.body, string(data))
		})
	}
}
