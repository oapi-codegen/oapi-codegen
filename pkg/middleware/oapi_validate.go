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
	"os"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

const (
	EchoContextKey = "oapi-codegen/echo-context"
	UserDataKey    = "oapi-codegen/user-data"
)

// This is an Echo middleware function which validates incoming HTTP requests
// to make sure that they conform to the given OAPI 3.0 specification. When
// OAPI validation fails on the request, we return an HTTP/400.

// Create validator middleware from a YAML file path
func OapiValidatorFromYamlFile(path string) (echo.MiddlewareFunc, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %s", path, err)
	}

	swagger, err := openapi3.NewLoader().LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s as Swagger YAML: %s",
			path, err)
	}
	return OapiRequestValidator(swagger), nil
}

// Create a validator from a swagger object.
func OapiRequestValidator(swagger *openapi3.T) echo.MiddlewareFunc {
	return OapiRequestValidatorWithOptions(swagger, nil)
}

// ErrorHandler is called when there is an error in validation
type ErrorHandler func(c echo.Context, err *echo.HTTPError) error

// MultiErrorHandler is called when oapi returns a MultiError type
type MultiErrorHandler func(openapi3.MultiError) *echo.HTTPError

// Options to customize request validation. These are passed through to
// openapi3filter.
type Options struct {
	ErrorHandler      ErrorHandler
	Options           openapi3filter.Options
	ParamDecoder      openapi3filter.ContentParameterDecoder
	UserData          interface{}
	Skipper           echomiddleware.Skipper
	MultiErrorHandler MultiErrorHandler
}

// OapiRequestValidatorWithOptions creates a validator from a swagger object, with validation options
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
				if options != nil && options.ErrorHandler != nil {
					return options.ErrorHandler(c, err)
				}
				return err
			}
			return next(c)
		}
	}
}

// ValidateRequestFromContext is called from the middleware above and actually does the work
// of validating a request.
func ValidateRequestFromContext(ctx echo.Context, router routers.Router, options *Options) *echo.HTTPError {
	req := ctx.Request()
	route, pathParams, err := router.FindRoute(req)

	// We failed to find a matching route for the request.
	if err != nil {
		switch e := err.(type) {
		case *routers.RouteError:
			// We've got a bad request, the path requested doesn't match
			// either server, or path, or something.
			return echo.NewHTTPError(http.StatusBadRequest, e.Reason)
		default:
			// This should never happen today, but if our upstream code changes,
			// we don't want to crash the server, so handle the unexpected error.
			return echo.NewHTTPError(http.StatusInternalServerError,
				fmt.Sprintf("error validating route: %s", err.Error()))
		}
	}

	validationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}

	// Pass the Echo context into the request validator, so that any callbacks
	// which it invokes make it available.
	requestContext := context.WithValue(context.Background(), EchoContextKey, ctx)

	if options != nil {
		validationInput.Options = &options.Options
		validationInput.ParamDecoder = options.ParamDecoder
		requestContext = context.WithValue(requestContext, UserDataKey, options.UserData)
	}

	err = openapi3filter.ValidateRequest(requestContext, validationInput)
	if err != nil {
		me := openapi3.MultiError{}
		if errors.As(err, &me) {
			errFunc := getMultiErrorHandlerFromOptions(options)
			return errFunc(me)
		}

		switch e := err.(type) {
		case *openapi3filter.RequestError:
			// We've got a bad request
			// Split up the verbose error by lines and return the first one
			// openapi errors seem to be multi-line with a decent message on the first
			errorLines := strings.Split(e.Error(), "\n")
			return &echo.HTTPError{
				Code:     http.StatusBadRequest,
				Message:  errorLines[0],
				Internal: err,
			}
		case *openapi3filter.SecurityRequirementsError:
			for _, err := range e.Errors {
				httpErr, ok := err.(*echo.HTTPError)
				if ok {
					return httpErr
				}
			}
			return &echo.HTTPError{
				Code:     http.StatusForbidden,
				Message:  e.Error(),
				Internal: err,
			}
		default:
			// This should never happen today, but if our upstream code changes,
			// we don't want to crash the server, so handle the unexpected error.
			return &echo.HTTPError{
				Code:     http.StatusInternalServerError,
				Message:  fmt.Sprintf("error validating request: %s", err),
				Internal: err,
			}
		}
	}
	return nil
}

// Helper function to get the echo context from within requests. It returns
// nil if not found or wrong type.
func GetEchoContext(c context.Context) echo.Context {
	iface := c.Value(EchoContextKey)
	if iface == nil {
		return nil
	}
	eCtx, ok := iface.(echo.Context)
	if !ok {
		return nil
	}
	return eCtx
}

func GetUserData(c context.Context) interface{} {
	return c.Value(UserDataKey)
}

// attempt to get the skipper from the options whether it is set or not
func getSkipperFromOptions(options *Options) echomiddleware.Skipper {
	if options == nil {
		return echomiddleware.DefaultSkipper
	}

	if options.Skipper == nil {
		return echomiddleware.DefaultSkipper
	}

	return options.Skipper
}

// attempt to get the MultiErrorHandler from the options. If it is not set,
// return a default handler
func getMultiErrorHandlerFromOptions(options *Options) MultiErrorHandler {
	if options == nil {
		return defaultMultiErrorHandler
	}

	if options.MultiErrorHandler == nil {
		return defaultMultiErrorHandler
	}

	return options.MultiErrorHandler
}

// defaultMultiErrorHandler returns a StatusBadRequest (400) and a list
// of all of the errors. This method is called if there are no other
// methods defined on the options.
func defaultMultiErrorHandler(me openapi3.MultiError) *echo.HTTPError {
	return &echo.HTTPError{
		Code:     http.StatusBadRequest,
		Message:  me.Error(),
		Internal: me,
	}
}
