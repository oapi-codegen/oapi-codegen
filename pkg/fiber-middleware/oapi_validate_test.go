package middleware

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed test_spec.yaml
var testSchema []byte

func doGet(t *testing.T, app *fiber.App, rawURL string) *http.Response {
	t.Helper()

	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("invalid URL %q: %v", rawURL, err)
	}

	req := httptest.NewRequest("GET", u.RequestURI(), nil)
	req.Header.Add("Accept", "application/json")
	req.Host = u.Host

	r, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request, URL=%q: %v", rawURL, err)
	}

	return r
}

func doPost(t *testing.T, app *fiber.App, rawURL string, jsonBody interface{}) *http.Response {
	t.Helper()

	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("invalid url %q: %v", rawURL, err)
	}

	buf, err := json.Marshal(jsonBody)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	req := httptest.NewRequest("POST", u.RequestURI(), bytes.NewReader(buf))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Host = u.Host

	r, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request, URL=%q: %v", rawURL, err)
	}

	return r
}

func TestOapiRequestValidator(t *testing.T) {

	swagger, err := openapi3.NewLoader().LoadFromData(testSchema)
	require.NoError(t, err, "Error initializing swagger")

	// Create a new fiber router
	app := fiber.New()

	// Set up an authenticator to check authenticated function. It will allow
	// access to "someScope", but disallow others.
	options := Options{
		ErrorHandler: func(c *fiber.Ctx, message string, statusCode int) {
			_ = c.Status(statusCode).SendString("test: " + message)
		},
		Options: openapi3filter.Options{
			AuthenticationFunc: func(c context.Context, input *openapi3filter.AuthenticationInput) error {
				// The fiber context should be propagated into here.
				gCtx := GetFiberContext(c)
				assert.NotNil(t, gCtx)
				// As should user data
				assert.EqualValues(t, "hi!", GetUserData(c))

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
	app.Use(OapiRequestValidatorWithOptions(swagger, &options))

	called := false

	// Install a request handler for /resource. We want to make sure it doesn't
	// get called.
	app.Get("/resource", func(c *fiber.Ctx) error {
		called = true
		return nil
	})
	// Let's send the request to the wrong server, this should fail validation
	{
		res := doGet(t, app, "https://not.deepmap.ai/resource")
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		assert.False(t, called, "Handler should not have been called")
	}

	// Let's send a good request, it should pass
	{
		res := doGet(t, app, "https://deepmap.ai/resource")
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.True(t, called, "Handler should have been called")
		called = false
	}

	// Send an out-of-spec parameter
	{
		res := doGet(t, app, "https://deepmap.ai/resource?id=500")
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	// Send a bad parameter type
	{
		res := doGet(t, app, "https://deepmap.ai/resource?id=foo")
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	// Add a handler for the POST message
	app.Post("/resource", func(c *fiber.Ctx) error {
		called = true
		return c.SendStatus(http.StatusNoContent)
	})

	called = false
	// Send a good request body
	{
		body := struct {
			Name string `json:"name"`
		}{
			Name: "Marcin",
		}
		res := doPost(t, app, "https://deepmap.ai/resource", body)
		assert.Equal(t, http.StatusNoContent, res.StatusCode)
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
		res := doPost(t, app, "https://deepmap.ai/resource", body)
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	app.Get("/protected_resource", func(c *fiber.Ctx) error {
		called = true
		return c.SendStatus(http.StatusNoContent)
	})

	// Call a protected function to which we have access
	{
		res := doGet(t, app, "https://deepmap.ai/protected_resource")
		assert.Equal(t, http.StatusNoContent, res.StatusCode)
		assert.True(t, called, "Handler should have been called")
		called = false
	}

	app.Get("/protected_resource2", func(c *fiber.Ctx) error {
		called = true
		return c.SendStatus(http.StatusNoContent)
	})
	// Call a protected function to which we don't have access
	{
		res := doGet(t, app, "https://deepmap.ai/protected_resource2")
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	app.Get("/protected_resource_401", func(c *fiber.Ctx) error {
		called = true
		return c.SendStatus(http.StatusNoContent)
	})
	// Call a protected function without credentials
	{
		res := doGet(t, app, "https://deepmap.ai/protected_resource_401")
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
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

	app := fiber.New()

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
	app.Use(OapiRequestValidatorWithOptions(swagger, &options))

	called := false

	// Install a request handler for /resource. We want to make sure it doesn't
	// get called.
	app.Get("/multiparamresource", func(c *fiber.Ctx) error {
		called = true
		return nil
	})

	// Let's send a good request, it should pass
	{
		res := doGet(t, app, "https://deepmap.ai/multiparamresource?id=50&id2=50")
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.True(t, called, "Handler should have been called")
		called = false
	}

	// Let's send a request with a missing parameter, it should return
	// a bad status
	{
		res := doGet(t, app, "https://deepmap.ai/multiparamresource?id=50")
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
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
		res := doGet(t, app, "https://deepmap.ai/multiparamresource")
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
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
		res := doGet(t, app, "https://deepmap.ai/multiparamresource?id=500")
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
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
		res := doGet(t, app, "https://deepmap.ai/multiparamresource?id=abc&id2=1")
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
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

	app := fiber.New()

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
	app.Use(OapiRequestValidatorWithOptions(swagger, &options))

	called := false

	// Install a request handler for /resource. We want to make sure it doesn't
	// get called.
	app.Get("/multiparamresource", func(c *fiber.Ctx) error {
		called = true
		return nil
	})

	// Let's send a good request, it should pass
	{
		res := doGet(t, app, "https://deepmap.ai/multiparamresource?id=50&id2=50")
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.True(t, called, "Handler should have been called")
		called = false
	}

	// Let's send a request with a missing parameter, it should return
	// a bad status
	{
		res := doGet(t, app, "https://deepmap.ai/multiparamresource?id=50")
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
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
		res := doGet(t, app, "https://deepmap.ai/multiparamresource")
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
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
		res := doGet(t, app, "https://deepmap.ai/multiparamresource?id=500")
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
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
		res := doGet(t, app, "https://deepmap.ai/multiparamresource?id=abc&id2=1")
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
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
