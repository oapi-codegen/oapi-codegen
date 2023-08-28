package middleware

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/kataras/iris/v12"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed test_spec.yaml
var testSchema []byte

func doGet(t *testing.T, i *iris.Application, rawURL string) *httptest.ResponseRecorder {
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("Invalid url: %s", rawURL)
	}

	response := testutil.NewRequest().Get(u.RequestURI()).WithHost(u.Host).WithAcceptJson().GoWithHTTPHandler(t, i)
	return response.Recorder
}

func doPost(t *testing.T, i *iris.Application, rawURL string, jsonBody interface{}) *httptest.ResponseRecorder {
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("Invalid url: %s", rawURL)
	}

	response := testutil.NewRequest().Post(u.RequestURI()).WithHost(u.Host).WithJsonBody(jsonBody).GoWithHTTPHandler(t, i)
	return response.Recorder
}

func TestOapiRequestValidator(t *testing.T) {
	swagger, err := openapi3.NewLoader().LoadFromData(testSchema)
	require.NoError(t, err, "Error initializing swagger")

	// Create a new iris router
	i := iris.New()

	// Set up an authenticator to check authenticated function. It will allow
	// access to "someScope", but disallow others.
	options := Options{
		ErrorHandler: func(ctx iris.Context, message string, statusCode int) {
			ctx.StopWithText(statusCode, "test: "+message)
		},
		Options: openapi3filter.Options{
			AuthenticationFunc: func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
				// The iris context should be propagated into here.
				iCtx := GetIrisContext(ctx)
				assert.NotNil(t, iCtx)
				// As should user data
				assert.EqualValues(t, "hi!", GetUserData(ctx))

				for _, s := range input.Scopes {
					if s == "someScope" {
						return nil
					}
					if s == "unauthorized" {
						return errors.New("unauthorized")
					}
				}
				return errors.New("forbidden")
			},
		},
		UserData: "hi!",
	}

	// Install our OpenApi based request validator
	i.Use(OapiRequestValidatorWithOptions(swagger, &options))

	called := false

	// Install a request handler for /resource. We want to make sure it doesn't
	// get called.
	i.Get("/resource", func(ctx iris.Context) {
		called = true
	})

	// Add a handler for the POST message
	i.Post("/resource", func(ctx iris.Context) {
		called = true
		ctx.StatusCode(http.StatusNoContent)
	})

	i.Get("/protected_resource", func(ctx iris.Context) {
		called = true
		ctx.StatusCode(http.StatusNoContent)
	})

	i.Get("/protected_resource2", func(ctx iris.Context) {
		called = true
		ctx.StatusCode(http.StatusNoContent)
	})

	i.Get("/protected_resource_401", func(ctx iris.Context) {
		called = true
		ctx.StatusCode(http.StatusNoContent)
	})

	if err := i.Build(); err != nil {
		t.Fatalf("Error building iris: %s", err)
	}

	// Let's send the request to the wrong server, this should fail validation
	{
		res := doGet(t, i, "https://not.deepmap.ai/resource")
		assert.Equal(t, http.StatusBadRequest, res.Code)
		assert.False(t, called, "Handler should not have been called")
	}

	// Let's send a good request, it should pass
	{
		res := doGet(t, i, "https://deepmap.ai/resource")
		assert.Equal(t, http.StatusOK, res.Code)
		assert.True(t, called, "Handler should have been called")
		called = false
	}

	// Send an out-of-spec parameter
	{
		res := doGet(t, i, "https://deepmap.ai/resource?id=500")
		assert.Equal(t, http.StatusBadRequest, res.Code)
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	// Send a bad parameter type
	{
		res := doGet(t, i, "https://deepmap.ai/resource?id=foo")
		assert.Equal(t, http.StatusBadRequest, res.Code)
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	called = false
	// Send a good request body
	{
		body := struct {
			Name string `json:"name"`
		}{
			Name: "Marcin",
		}
		res := doPost(t, i, "https://deepmap.ai/resource", body)
		assert.Equal(t, http.StatusNoContent, res.Code)
		assert.True(t, called, "Handler should have been called")
		called = false
	}

	// Send a malformed body
	{
		body := struct {
			Name int `json:"name"`
		}{
			Name: 7,
		}
		res := doPost(t, i, "https://deepmap.ai/resource", body)
		assert.Equal(t, http.StatusBadRequest, res.Code)
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	// Call a protected function to which we have access
	{
		res := doGet(t, i, "https://deepmap.ai/protected_resource")
		assert.Equal(t, http.StatusNoContent, res.Code)
		assert.True(t, called, "Handler should have been called")
		called = false
	}

	// Call a protected function to which we don't have access
	{
		res := doGet(t, i, "https://deepmap.ai/protected_resource2")
		assert.Equal(t, http.StatusBadRequest, res.Code)
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	// Call a protected function without credentials
	{
		res := doGet(t, i, "https://deepmap.ai/protected_resource_401")
		assert.Equal(t, http.StatusBadRequest, res.Code)
		body, err := io.ReadAll(res.Body)
		if assert.NoError(t, err) {
			assert.Equal(t, "test: error in openapi3filter.SecurityRequirementsError: security requirements failed: unauthorized", string(body))
		}
		assert.False(t, called, "Handler should not have been called")
		called = false
	}
}

func TestOapiRequestValidatorWithOptionsMultiError(t *testing.T) {
	swagger, err := openapi3.NewLoader().LoadFromData(testSchema)
	require.NoError(t, err, "Error initializing swagger")

	i := iris.New()

	// Set up an authenticator to check authenticated function. It will allow
	// access to "someScope", but disallow others.
	options := Options{
		Options: openapi3filter.Options{
			ExcludeRequestBody:    false,
			ExcludeResponseBody:   false,
			IncludeResponseStatus: true,
			MultiError:            true,
		},
	}

	// register middleware
	i.Use(OapiRequestValidatorWithOptions(swagger, &options))

	called := false

	// Install a request handler for /resource. We want to make sure it doesn't
	// get called.
	i.Get("/multiparamresource", func(ctx iris.Context) {
		called = true
	})

	if err := i.Build(); err != nil {
		t.Fatalf("Error building iris: %s", err)
	}

	// Let's send a good request, it should pass
	{
		res := doGet(t, i, "https://deepmap.ai/multiparamresource?id=50&id2=50")
		assert.Equal(t, http.StatusOK, res.Code)
		assert.True(t, called, "Handler should have been called")
		called = false
	}

	// Let's send a request with a missing parameter, it should return
	// a bad status
	{
		res := doGet(t, i, "https://deepmap.ai/multiparamresource?id=50")
		assert.Equal(t, http.StatusBadRequest, res.Code)
		body, err := io.ReadAll(res.Body)
		if assert.NoError(t, err) {
			assert.Contains(t, string(body), "multiple errors encountered")
			assert.Contains(t, string(body), "parameter \"id2\"")
			assert.Contains(t, string(body), "value is required but missing")
		}
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	// Let's send a request with a 2 missing parameters, it should return
	// a bad status
	{
		res := doGet(t, i, "https://deepmap.ai/multiparamresource")
		assert.Equal(t, http.StatusBadRequest, res.Code)
		body, err := io.ReadAll(res.Body)
		if assert.NoError(t, err) {
			assert.Contains(t, string(body), "multiple errors encountered")
			assert.Contains(t, string(body), "parameter \"id\"")
			assert.Contains(t, string(body), "value is required but missing")
			assert.Contains(t, string(body), "parameter \"id2\"")
			assert.Contains(t, string(body), "value is required but missing")
		}
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	// Let's send a request with a 1 missing parameter, and another outside
	// or the parameters. It should return a bad status
	{
		res := doGet(t, i, "https://deepmap.ai/multiparamresource?id=500")
		assert.Equal(t, http.StatusBadRequest, res.Code)
		body, err := io.ReadAll(res.Body)
		if assert.NoError(t, err) {
			assert.Contains(t, string(body), "multiple errors encountered")
			assert.Contains(t, string(body), "parameter \"id\"")
			assert.Contains(t, string(body), "number must be at most 100")
			assert.Contains(t, string(body), "parameter \"id2\"")
			assert.Contains(t, string(body), "value is required but missing")
		}
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	// Let's send a request with a parameters that do not meet spec. It should
	// return a bad status
	{
		res := doGet(t, i, "https://deepmap.ai/multiparamresource?id=abc&id2=1")
		assert.Equal(t, http.StatusBadRequest, res.Code)
		body, err := io.ReadAll(res.Body)
		if assert.NoError(t, err) {
			assert.Contains(t, string(body), "multiple errors encountered")
			assert.Contains(t, string(body), "parameter \"id\"")
			assert.Contains(t, string(body), "value abc: an invalid integer: invalid syntax")
			assert.Contains(t, string(body), "parameter \"id2\"")
			assert.Contains(t, string(body), "number must be at least 10")
		}
		assert.False(t, called, "Handler should not have been called")
		called = false
	}
}

func TestOapiRequestValidatorWithOptionsMultiErrorAndCustomHandler(t *testing.T) {
	swagger, err := openapi3.NewLoader().LoadFromData(testSchema)
	require.NoError(t, err, "Error initializing swagger")

	i := iris.New()

	// Set up an authenticator to check authenticated function. It will allow
	// access to "someScope", but disallow others.
	options := Options{
		Options: openapi3filter.Options{
			ExcludeRequestBody:    false,
			ExcludeResponseBody:   false,
			IncludeResponseStatus: true,
			MultiError:            true,
		},
		MultiErrorHandler: func(me openapi3.MultiError) error {
			return fmt.Errorf("Bad stuff -  %s", me.Error())
		},
	}

	// register middleware
	i.Use(OapiRequestValidatorWithOptions(swagger, &options))

	called := false

	// Install a request handler for /resource. We want to make sure it doesn't
	// get called.
	i.Get("/multiparamresource", func(ctx iris.Context) {
		called = true
	})

	if err := i.Build(); err != nil {
		t.Fatalf("Error building iris: %s", err)
	}

	// Let's send a good request, it should pass
	{
		res := doGet(t, i, "https://deepmap.ai/multiparamresource?id=50&id2=50")
		assert.Equal(t, http.StatusOK, res.Code)
		assert.True(t, called, "Handler should have been called")
		called = false
	}

	// Let's send a request with a missing parameter, it should return
	// a bad status
	{
		res := doGet(t, i, "https://deepmap.ai/multiparamresource?id=50")
		assert.Equal(t, http.StatusBadRequest, res.Code)
		body, err := io.ReadAll(res.Body)
		if assert.NoError(t, err) {
			assert.Contains(t, string(body), "Bad stuff")
			assert.Contains(t, string(body), "parameter \"id2\"")
			assert.Contains(t, string(body), "value is required but missing")
		}
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	// Let's send a request with a 2 missing parameters, it should return
	// a bad status
	{
		res := doGet(t, i, "https://deepmap.ai/multiparamresource")
		assert.Equal(t, http.StatusBadRequest, res.Code)
		body, err := io.ReadAll(res.Body)
		if assert.NoError(t, err) {
			assert.Contains(t, string(body), "Bad stuff")
			assert.Contains(t, string(body), "parameter \"id\"")
			assert.Contains(t, string(body), "value is required but missing")
			assert.Contains(t, string(body), "parameter \"id2\"")
			assert.Contains(t, string(body), "value is required but missing")
		}
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	// Let's send a request with a 1 missing parameter, and another outside
	// or the parameters. It should return a bad status
	{
		res := doGet(t, i, "https://deepmap.ai/multiparamresource?id=500")
		assert.Equal(t, http.StatusBadRequest, res.Code)
		body, err := io.ReadAll(res.Body)
		if assert.NoError(t, err) {
			assert.Contains(t, string(body), "Bad stuff")
			assert.Contains(t, string(body), "parameter \"id\"")
			assert.Contains(t, string(body), "number must be at most 100")
			assert.Contains(t, string(body), "parameter \"id2\"")
			assert.Contains(t, string(body), "value is required but missing")
		}
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	// Let's send a request with a parameters that do not meet spec. It should
	// return a bad status
	{
		res := doGet(t, i, "https://deepmap.ai/multiparamresource?id=abc&id2=1")
		assert.Equal(t, http.StatusBadRequest, res.Code)
		body, err := io.ReadAll(res.Body)
		if assert.NoError(t, err) {
			assert.Contains(t, string(body), "Bad stuff")
			assert.Contains(t, string(body), "parameter \"id\"")
			assert.Contains(t, string(body), "value abc: an invalid integer: invalid syntax")
			assert.Contains(t, string(body), "parameter \"id2\"")
			assert.Contains(t, string(body), "number must be at least 10")
		}
		assert.False(t, called, "Handler should not have been called")
		called = false
	}
}
