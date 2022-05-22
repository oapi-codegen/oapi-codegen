// Package middleware implements middleware function for go-chi or net/http,
// which validates incoming HTTP requests to make sure that they conform to the given OAPI 3.0 specification.
// When OAPI validation failes on the request, we return an HTTP/400.
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

// ErrorHandler is called when there is an error in validation
type ErrorHandler func(w http.ResponseWriter, message string, statusCode int)

// Options to customize request validation, openapi3filter specified options will be passed through.
type Options struct {
	Options      openapi3filter.Options
	ErrorHandler ErrorHandler
}

// OapiRequestValidator Creates middleware to validate request by swagger spec.
// This middleware is good for net/http either since go-chi is 100% compatible with net/http.
func OapiRequestValidator(swagger *openapi3.T) func(next http.Handler) http.Handler {
	return OapiRequestValidatorWithOptions(swagger, nil)
}

// OapiRequestValidatorWithOptions Creates middleware to validate request by swagger spec.
// This middleware is good for net/http either since go-chi is 100% compatible with net/http.
func OapiRequestValidatorWithOptions(swagger *openapi3.T, options *Options) func(next http.Handler) http.Handler {
	router, err := gorillamux.NewRouter(swagger)
	if err != nil {
		panic(err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// validate request
			if statusCode, err := validateRequest(r, router, options); err != nil {
				if options != nil && options.ErrorHandler != nil {
					options.ErrorHandler(w, err.Error(), statusCode)
				} else {
					http.Error(w, err.Error(), statusCode)
				}
				return
			}

			// serve
			next.ServeHTTP(w, r)
		})
	}

}

// This function is called from the middleware above and actually does the work
// of validating a request.
func validateRequest(r *http.Request, router routers.Router, options *Options) (int, error) {

	// Find route
	route, pathParams, err := router.FindRoute(r)
	if err != nil {
		return http.StatusBadRequest, err // We failed to find a matching route for the request.
	}

	// Validate request
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    r,
		PathParams: pathParams,
		Route:      route,
	}

	if options != nil {
		requestValidationInput.Options = &options.Options
	}

	if err := openapi3filter.ValidateRequest(context.Background(), requestValidationInput); err != nil {
		switch e := err.(type) {
		case *openapi3filter.RequestError:
			// We've got a bad request
			// Split up the verbose error by lines and return the first one
			// openapi errors seem to be multi-line with a decent message on the first
			errorLines := strings.Split(e.Error(), "\n")
			return http.StatusBadRequest, fmt.Errorf(errorLines[0])
		case *openapi3filter.SecurityRequirementsError:
			return http.StatusUnauthorized, err
		default:
			// This should never happen today, but if our upstream code changes,
			// we don't want to crash the server, so handle the unexpected error.
			return http.StatusInternalServerError, fmt.Errorf("error validating route: %s", err.Error())
		}
	}

	return http.StatusOK, nil
}
