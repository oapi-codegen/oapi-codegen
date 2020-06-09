// Copyright 2019 DeepMap, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tidepool-org/oapi-codegen/pkg/testutil"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
)

var testSchema = `openapi: "3.0.0"
info:
  version: 1.0.0
  title: TestServer
servers:
  - url: http://deepmap.ai
paths:
  /resource:
    get:
      operationId: getResource
      parameters:
        - name: id
          in: query
          schema:
            type: integer
            minimum: 10
            maximum: 100
      responses:
        '200':
            content:
              application/json:
                schema:
                  properties:
                    name:
                      type: string
                    id:
                      type: integer
    post:
      operationId: createResource
      responses:
        '204':
          description: No content
      requestBody:
        required: true
        content:
          application/json:
            schema:
              properties:
                name:
                  type: string
  /protected_resource:
    get:
      operationId: getProtectedResource
      security:
        - BearerAuth:
          - someScope
      responses:
        '204':
          description: no content
  /protected_resource2:
    get:
      operationId: getProtectedResource
      security:
        - BearerAuth:
          - otherScope
      responses:
        '204':
          description: no content
  /protected_resource_401:
    get:
      operationId: getProtectedResource
      security:
        - BearerAuth:
          - unauthorized
      responses:
        '401':
          description: no content
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
`

func doGet(t *testing.T, e *echo.Echo, url string) *httptest.ResponseRecorder {
	response := testutil.NewRequest().Get(url).WithAcceptJson().Go(t, e)
	return response.Recorder
}

func doPost(t *testing.T, e *echo.Echo, url string, jsonBody interface{}) *httptest.ResponseRecorder {
	response := testutil.NewRequest().Post(url).WithJsonBody(jsonBody).Go(t, e)
	return response.Recorder
}

func TestOapiRequestValidator(t *testing.T) {
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(testSchema))
	assert.NoError(t, err, "Error initializing swagger")

	// Create a new echo router
	e := echo.New()

	// Set up an authenticator to check authenticated function. It will allow
	// access to "someScope", but disallow others.
	options := Options{
		Options: openapi3filter.Options{
			AuthenticationFunc: func(c context.Context, input *openapi3filter.AuthenticationInput) error {
				// The echo context should be propagated into here.
				eCtx := GetEchoContext(c)
				assert.NotNil(t, eCtx)
				// As should user data
				assert.EqualValues(t, "hi!", GetUserData(c))

				for _, s := range input.Scopes {
					if s == "someScope" {
						return nil
					}
					if s == "unauthorized" {
						return echo.ErrUnauthorized
					}
				}
				return errors.New("forbidden")
			},
		},
		UserData: "hi!",
	}

	// Install our OpenApi based request validator
	e.Use(OapiRequestValidatorWithOptions(swagger, &options))

	called := false

	// Install a request handler for /resource. We want to make sure it doesn't
	// get called.
	e.GET("/resource", func(c echo.Context) error {
		called = true
		return nil
	})
	// Let's send the request to the wrong server, this should fail validation
	{
		rec := doGet(t, e, "http://not.deepmap.ai/resource")
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, called, "Handler should not have been called")
	}

	// Let's send a good request, it should pass
	{
		rec := doGet(t, e, "http://deepmap.ai/resource")
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, called, "Handler should have been called")
		called = false
	}

	// Send an out-of-spec parameter
	{
		rec := doGet(t, e, "http://deepmap.ai/resource?id=500")
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	// Send a bad parameter type
	{
		rec := doGet(t, e, "http://deepmap.ai/resource?id=foo")
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	// Add a handler for the POST message
	e.POST("/resource", func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusNoContent)
	})

	called = false
	// Send a good request body
	{
		body := struct {
			Name string `json:"name"`
		}{
			Name: "Marcin",
		}
		rec := doPost(t, e, "http://deepmap.ai/resource", body)
		assert.Equal(t, http.StatusNoContent, rec.Code)
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
		rec := doPost(t, e, "http://deepmap.ai/resource", body)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	e.GET("/protected_resource", func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusNoContent)

	})

	// Call a protected function to which we have access
	{
		rec := doGet(t, e, "http://deepmap.ai/protected_resource")
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.True(t, called, "Handler should have been called")
		called = false
	}

	e.GET("/protected_resource2", func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusNoContent)
	})
	// Call a protected function to which we dont have access
	{
		rec := doGet(t, e, "http://deepmap.ai/protected_resource2")
		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.False(t, called, "Handler should not have been called")
		called = false
	}

	e.GET("/protected_resource_401", func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusNoContent)
	})
	// Call a protected function without credentials
	{
		rec := doGet(t, e, "http://deepmap.ai/protected_resource_401")
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.False(t, called, "Handler should not have been called")
		called = false
	}
}

func TestGetSkipperFromOptions(t *testing.T) {

	options := new(Options)
	assert.NotNil(t, getSkipperFromOptions(options))

	options = &Options{}
	assert.NotNil(t, getSkipperFromOptions(options))

	options = &Options{
		Skipper: echomiddleware.DefaultSkipper,
	}
	assert.NotNil(t, getSkipperFromOptions(options))
}
