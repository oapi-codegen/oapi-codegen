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
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

type echoContextKeyType string
type userDataKeyType string

const (
	echoContextKey echoContextKeyType = "echo-context"
	userDataKey    userDataKeyType    = "user-data"
)

// This is an Echo middleware function which validates incoming HTTP requests
// to make sure that they conform to the given OAPI 3.0 specification. When
// OAPI validation fails on the request, we return an HTTP/400.

// Create validator middleware from a YAML file path.
func OapiValidatorFromYamlFile(path string) (echo.MiddlewareFunc, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", path, err)
	}

	swagger, err := openapi3.NewLoader().LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s as Swagger YAML: %w", path, err)
	}
	return OapiRequestValidator(swagger), nil
}

// Create a validator from a swagger object.
func OapiRequestValidator(swagger *openapi3.T) echo.MiddlewareFunc {
	return OapiRequestValidatorWithOptions(swagger, nil)
}

// Options to customize request validation. These are passed through to
// openapi3filter.
type Options struct {
	Options      openapi3filter.Options
	ParamDecoder openapi3filter.ContentParameterDecoder
	UserData     interface{}
	Skipper      echomiddleware.Skipper
}

// Create a validator from a swagger object, with validation options.
func OapiRequestValidatorWithOptions(swagger *openapi3.T, options *Options) echo.MiddlewareFunc {
	router, err := gorillamux.NewRouter(swagger)
	if err != nil {
		panic(err)
	}

	skipper := getSkipperFromOptions(options)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if skipper(c) {
				return next(c)
			}

			err := ValidateRequestFromContext(c, router, options)
			if err != nil {
				return err
			}
			return next(c)
		}
	}
}

// ValidateRequestFromContext is called from the middleware above and actually does the work
// of validating a request.
func ValidateRequestFromContext(ctx echo.Context, router routers.Router, options *Options) error {
	req := ctx.Request()
	route, pathParams, err := router.FindRoute(req)

	// We failed to find a matching route for the request.
	if err != nil {
		var re *routers.RouteError
		if errors.As(err, &re) {
			// We've got a bad request, the path requested doesn't match
			// either server, or path, or something.
			return echo.NewHTTPError(http.StatusBadRequest, re.Reason)
		}

		// This should never happen today, but if our upstream code changes,
		// we don't want to crash the server, so handle the unexpected error.
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Sprintf("error validating route: %s", err.Error()))
	}

	validationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}

	// Pass the Echo context into the request validator, so that any callbacks
	// which it invokes make it available.
	requestContext := context.WithValue(context.Background(), echoContextKey, ctx)

	if options != nil {
		validationInput.Options = &options.Options
		validationInput.ParamDecoder = options.ParamDecoder
		requestContext = context.WithValue(requestContext, userDataKey, options.UserData)
	}

	err = openapi3filter.ValidateRequest(requestContext, validationInput)
	if err != nil {
		var re *openapi3filter.RequestError
		if errors.As(err, &re) {
			// We've got a bad request
			// Split up the verbose error by lines and return the first one
			// openapi errors seem to be multi-line with a decent message on the first
			errorLines := strings.Split(re.Error(), "\n")
			return &echo.HTTPError{
				Code:     http.StatusBadRequest,
				Message:  errorLines[0],
				Internal: err,
			}
		}

		var sre *openapi3filter.SecurityRequirementsError
		if errors.As(err, &sre) {
			for _, err := range sre.Errors {
				var he *echo.HTTPError
				if errors.As(err, &he) {
					return he
				}
			}
			return &echo.HTTPError{
				Code:     http.StatusForbidden,
				Message:  sre.Error(),
				Internal: err,
			}
		}

		// This should never happen today, but if our upstream code changes,
		// we don't want to crash the server, so handle the unexpected error.
		return &echo.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  fmt.Sprintf("error validating request: %s", err),
			Internal: err,
		}
	}
	return nil
}

// GetEchoContents returns echo.Context from within requests. It returns
// nil if not found or wrong type.
func GetEchoContext(c context.Context) echo.Context {
	eCtx, ok := c.Value(echoContextKey).(echo.Context)
	if !ok {
		return nil
	}
	return eCtx
}

// GetUserData returns user data from within requests. It returns
// nil if not found.
func GetUserData(c context.Context) interface{} {
	return c.Value(userDataKey)
}

// attempt to get the skipper from the options whether it is set or not.
func getSkipperFromOptions(options *Options) echomiddleware.Skipper {
	if options == nil {
		return echomiddleware.DefaultSkipper
	}

	if options.Skipper == nil {
		return echomiddleware.DefaultSkipper
	}

	return options.Skipper
}
