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
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
)

// This is an Echo middleware function which validates incoming HTTP requests
// to make sure that they conform to the given OAPI 3.0 specification. When
// OAPI validation failes on the request, we return an HTTP/400.

// Create validator middleware from a YAML file path
func OapiValidatorFromYamlFile(path string) (echo.MiddlewareFunc, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %s", path, err)
	}

	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s as Swagger YAML: %s",
			path, err)
	}
	return OapiRequestValidator(swagger), nil
}

// Create a validator from a swagger object.
func OapiRequestValidator(swagger *openapi3.Swagger) echo.MiddlewareFunc {
	router := openapi3filter.NewRouter().WithSwagger(swagger)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := ValidateRequestFromContext(c, router)
			if err != nil {
				return err
			}
			return next(c)
		}
	}
}

// This function is called from the middleware above and actually does the work
// of validating a request.
// TODO(marcin): kin-openapi, which we use for swagger validation, currently only
//   validates string parameters, and assumes that the rest are valid. I need to
//   handle param validation myself, or fix their code.
func ValidateRequestFromContext(ctx echo.Context, router *openapi3filter.Router) error {
	req := ctx.Request()
	route, pathParams, err := router.FindRoute(req.Method, req.URL)

	// We failed to find a matching route for the request.
	if err != nil {
		switch e := err.(type) {
		case *openapi3filter.RouteError:
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

	err = openapi3filter.ValidateRequest(context.Background(),
		&openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
		})
	if err != nil {
		switch e := err.(type) {
		case *openapi3filter.RequestError:
			// We've got a bad request
			return echo.NewHTTPError(http.StatusBadRequest, e.Reason)
		default:
			// This should never happen today, but if our upstream code changes,
			// we don't want to crash the server, so handle the unexpected error.
			return echo.NewHTTPError(http.StatusInternalServerError,
				fmt.Sprintf("error validating request: %s", err))
		}
	}
	return nil
}
