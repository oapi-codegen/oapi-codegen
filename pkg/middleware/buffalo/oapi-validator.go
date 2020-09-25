package oapimiddleware

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gobuffalo/buffalo"
)

const BuffaloContextKey = "oapi-codegen/buffalo-context"

var (
	swagger *openapi3.Swagger
	router  *openapi3filter.Router
)

// Create a validator from a swagger object.
func New(s *openapi3.Swagger) {
	swagger = s
	router = openapi3filter.NewRouter().WithSwagger(swagger)
}

// OAPIMiddleware validates against the Swagger definition
func OAPIMiddleware(next buffalo.Handler) buffalo.Handler {
	return func(ctx buffalo.Context) error {
		req := ctx.Request()
		route, pathParams, err := router.FindRoute(req.Method, req.URL)
		// We failed to find a matching route for the request.
		if err != nil {
			switch e := err.(type) {
			case *openapi3filter.RouteError:
				// We've got a bad request, the path requested doesn't match
				// either server, or path, or something.
				return ctx.Error(http.StatusBadRequest, errors.New(e.Reason))
			default:
				// This should never happen today, but if our upstream code changes,
				// we don't want to crash the server, so handle the unexpected error.
				return ctx.Error(http.StatusInternalServerError,
					errors.New(fmt.Sprintf("error validating route: %s", err.Error())))
			}
		}

		validationInput := &openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
		}

		// Pass the Echo context into the request validator, so that any callbacks
		// which it invokes make it available.
		requestContext := context.WithValue(context.Background(), BuffaloContextKey, ctx)

		err = openapi3filter.ValidateRequest(requestContext, validationInput)
		if err != nil {
			switch e := err.(type) {
			case *openapi3filter.RequestError:
				// We've got a bad request
				// Split up the verbose error by lines and return the first one
				// openapi errors seem to be multi-line with a decent message on the first
				errorLines := strings.Split(e.Error(), "\n")
				return ctx.Error(http.StatusBadRequest, errors.New(errorLines[0]))
			case *openapi3filter.SecurityRequirementsError:
				for _, err := range e.Errors {
					httpErr, ok := err.(*buffalo.HTTPError)
					if ok {
						return httpErr
					}
				}
				return ctx.Error(http.StatusForbidden, errors.New(e.Error()))
			default:
				// This should never happen today, but if our upstream code changes,
				// we don't want to crash the server, so handle the unexpected error.
				return ctx.Error(http.StatusInternalServerError,
					errors.New(fmt.Sprintf("error validating request: %s", err)))
			}
		}

		return next(ctx)
	}
}
