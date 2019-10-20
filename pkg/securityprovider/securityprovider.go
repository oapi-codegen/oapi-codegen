// Package securityprovider contains some default securityprovider
// implementations, which can be used as a RequestEditorFn of a
// client.
package securityprovider

import (
	"context"
	"fmt"
	"net/http"
)

const (
	// ErrSecurityProviderApiKeyInvalidIn indicates a usage of an invalid In.
	// Should be cookie, header or query
	ErrSecurityProviderApiKeyInvalidIn = SecurityProviderError("invalid 'in' specified for apiKey")
)

// SecurityProviderError defines error values of a security provider.
type SecurityProviderError string

// Error implements the error interface.
func (e SecurityProviderError) Error() string {
	return string(e)
}

// NewSecurityProviderBasicAuth provides a SecurityProvider, which can solve
// the BasicAuth challenge for api-calls.
func NewSecurityProviderBasicAuth(username, password string) (*SecurityProviderBasicAuth, error) {
	return &SecurityProviderBasicAuth{
		username: username,
		password: password,
	}, nil
}

// SecurityProviderBasicAuth sends a base64-encoded combination of
// username, password along with a request.
type SecurityProviderBasicAuth struct {
	username string
	password string
}

// Intercept will attach an Authorization header to the request and ensures that
// the username, password are base64 encoded and attached to the header.
func (s *SecurityProviderBasicAuth) Intercept(req *http.Request, ctx context.Context) error {
	req.SetBasicAuth(s.username, s.password)
	return nil
}

// NewSecurityProviderBearerToken provides a SecurityProvider, which can solve
// the Bearer Auth challende for api-calls.
func NewSecurityProviderBearerToken(token string) (*SecurityProviderBearerToken, error) {
	return &SecurityProviderBearerToken{
		token: token,
	}, nil
}

// SecurityProviderBearerToken sends a token as part of an
// Authorization: Bearer header along with a request.
type SecurityProviderBearerToken struct {
	token string
}

// Intercept will attach an Authorization header to the request
// and ensures that the bearer token is attached to the header.
func (s *SecurityProviderBearerToken) Intercept(req *http.Request, ctx context.Context) error {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.token))
	return nil
}

// NewSecurityProviderApiKey will attach a generic apiKey for a given name
// either to a cookie, header or as a query parameter.
func NewSecurityProviderApiKey(in, name, apiKey string) (*SecurityProviderApiKey, error) {
	interceptors := map[string]func(req *http.Request, ctx context.Context) error{
		"cookie": func(req *http.Request, ctx context.Context) error {
			req.AddCookie(&http.Cookie{Name: name, Value: apiKey})
			return nil
		},
		"header": func(req *http.Request, ctx context.Context) error {
			req.Header.Add(name, apiKey)
			return nil
		},
		"query": func(req *http.Request, ctx context.Context) error {
			query := req.URL.Query()
			query.Add(name, apiKey)
			req.URL.RawQuery = query.Encode()
			return nil
		},
	}

	interceptor, ok := interceptors[in]
	if !ok {
		return nil, ErrSecurityProviderApiKeyInvalidIn
	}

	return &SecurityProviderApiKey{
		interceptor: interceptor,
	}, nil
}

// SecurityProviderApiKey will attach an apiKey either to a
// cookie, header or query.
type SecurityProviderApiKey struct {
	interceptor func(req *http.Request, ctx context.Context) error
}

// Intercept will attach a cookie, header or query param for the configured
// name and apiKey.
func (s *SecurityProviderApiKey) Intercept(req *http.Request, ctx context.Context) error {
	return s.interceptor(req, ctx)
}
