// Package middleware implements middleware function for server implementations,
// which validates incoming HTTP requests to make sure that they conform to the given OAPI 3.0 specification.
// When OAPI validation fails on the request, we return an HTTP/400.
package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

type ctxKeyFiberContext struct{}
type ctxKeyUserData struct{}

// OapiValidatorFromYamlFile creates a validator middleware from a YAML file path
func OapiValidatorFromYamlFile(path string) (fiber.Handler, error) {

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

// OapiRequestValidator is a fiber middleware function which validates incoming HTTP requests
// to make sure that they conform to the given OAPI 3.0 specification. When
// OAPI validation fails on the request, we return an HTTP/400 with error message
func OapiRequestValidator(swagger *openapi3.T) fiber.Handler {
	return OapiRequestValidatorWithOptions(swagger, nil)
}

// ErrorHandler is called when there is an error in validation
type ErrorHandler func(c *fiber.Ctx, message string, statusCode int)

// MultiErrorHandler is called when oapi returns a MultiError type
type MultiErrorHandler func(openapi3.MultiError) error

// Options to customize request validation. These are passed through to
// openapi3filter.
type Options struct {
	Options           openapi3filter.Options
	ErrorHandler      ErrorHandler
	ParamDecoder      openapi3filter.ContentParameterDecoder
	UserData          interface{}
	MultiErrorHandler MultiErrorHandler
}

// OapiRequestValidatorWithOptions creates a validator from a swagger object, with validation options
func OapiRequestValidatorWithOptions(swagger *openapi3.T, options *Options) fiber.Handler {

	router, err := gorillamux.NewRouter(swagger)
	if err != nil {
		panic(err)
	}

	return func(c *fiber.Ctx) error {

		err := ValidateRequestFromContext(c, router, options)
		if err != nil {
			if options != nil && options.ErrorHandler != nil {
				options.ErrorHandler(c, err.Error(), http.StatusBadRequest)
				// in case the handler didn't internally call Abort, stop the chain
				return nil
			} else {
				// note: I am not sure if this is the best way to handle this
				return fiber.NewError(http.StatusBadRequest, err.Error())
			}
		}
		return c.Next()
	}
}

// ValidateRequestFromContext is called from the middleware above and actually does the work
// of validating a request.
func ValidateRequestFromContext(c *fiber.Ctx, router routers.Router, options *Options) error {

	r, err := adaptor.ConvertRequest(c, false)
	if err != nil {
		return err
	}

	route, pathParams, err := router.FindRoute(r)

	// We failed to find a matching route for the request.
	if err != nil {
		switch e := err.(type) {
		case *routers.RouteError:
			// We've got a bad request, the path requested doesn't match
			// either server, or path, or something.
			return errors.New(e.Reason)
		default:
			// This should never happen today, but if our upstream code changes,
			// we don't want to crash the server, so handle the unexpected error.
			return fmt.Errorf("error validating route: %s", err.Error())
		}
	}

	// Validate request
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    r,
		PathParams: pathParams,
		Route:      route,
	}

	// Pass the fiber context into the request validator, so that any callbacks
	// which it invokes make it available.
	requestContext := context.WithValue(context.Background(), ctxKeyFiberContext{}, c) //nolint:staticcheck

	if options != nil {
		requestValidationInput.Options = &options.Options
		requestValidationInput.ParamDecoder = options.ParamDecoder
		requestContext = context.WithValue(requestContext, ctxKeyUserData{}, options.UserData) //nolint:staticcheck
	}

	err = openapi3filter.ValidateRequest(requestContext, requestValidationInput)
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
			return fmt.Errorf("error in openapi3filter.RequestError: %s", errorLines[0])
		case *openapi3filter.SecurityRequirementsError:
			return fmt.Errorf("error in openapi3filter.SecurityRequirementsError: %s", e.Error())
		default:
			// This should never happen today, but if our upstream code changes,
			// we don't want to crash the server, so handle the unexpected error.
			return fmt.Errorf("error validating request: %s", err)
		}
	}
	return nil
}

// GetFiberContext gets the fiber context from within requests. It returns
// nil if not found or wrong type.
func GetFiberContext(c context.Context) *fiber.Ctx {
	iface := c.Value(ctxKeyFiberContext{})
	if iface == nil {
		return nil
	}

	fiberCtx, ok := iface.(*fiber.Ctx)
	if ok {
		return fiberCtx
	}
	return nil
}

func GetUserData(c context.Context) interface{} {
	return c.Value(ctxKeyUserData{})
}

// getMultiErrorHandlerFromOptions attempts to get the MultiErrorHandler from the options. If it is not set,
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
// of all the errors. This method is called if there are no other
// methods defined on the options.
func defaultMultiErrorHandler(me openapi3.MultiError) error {
	return fmt.Errorf("multiple errors encountered: %s", me)
}
