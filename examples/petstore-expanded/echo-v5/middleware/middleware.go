// Package middleware provides OpenAPI request validation middleware for Echo v5.
//
// Adapted from github.com/oapi-codegen/echo-middleware (Echo v4) for use
// with github.com/labstack/echo/v5, which has breaking API changes
// (pointer context, NewHTTPError signature, no HTTPError.Internal field).
//
// This is intentionally inlined in the example because there is no published
// echo-v5 middleware package yet.
// TODO: make an echo-v5 middleware repo
package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/labstack/echo/v5"
	echomiddleware "github.com/labstack/echo/v5/middleware"
)

const (
	EchoContextKey = "oapi-codegen/echo-context"
	UserDataKey    = "oapi-codegen/user-data"
)

// OapiValidatorFromYamlFile creates validator middleware from a YAML file path.
func OapiValidatorFromYamlFile(path string) (echo.MiddlewareFunc, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", path, err)
	}

	spec, err := openapi3.NewLoader().LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s as OpenAPI YAML: %w", path, err)
	}
	return OapiRequestValidator(spec), nil
}

// OapiRequestValidator creates middleware to validate incoming requests against
// the given OpenAPI 3.x spec with default configuration.
func OapiRequestValidator(spec *openapi3.T) echo.MiddlewareFunc {
	return OapiRequestValidatorWithOptions(spec, nil)
}

// ErrorHandler is called when there is an error in validation.
type ErrorHandler func(c *echo.Context, err *echo.HTTPError) error

// MultiErrorHandler is called when the OpenAPI filter returns a MultiError.
type MultiErrorHandler func(openapi3.MultiError) *echo.HTTPError

// Options to customize request validation.
type Options struct {
	ErrorHandler          ErrorHandler
	Options               openapi3filter.Options
	ParamDecoder          openapi3filter.ContentParameterDecoder
	UserData              any
	Skipper               echomiddleware.Skipper
	MultiErrorHandler     MultiErrorHandler
	SilenceServersWarning bool
	DoNotValidateServers  bool
	Prefix                string
}

// OapiRequestValidatorWithOptions creates middleware with explicit configuration.
func OapiRequestValidatorWithOptions(spec *openapi3.T, options *Options) echo.MiddlewareFunc {
	if options != nil && options.DoNotValidateServers {
		spec.Servers = nil
	}

	if spec.Servers != nil && (options == nil || !options.SilenceServersWarning) {
		log.Println("WARN: OapiRequestValidatorWithOptions called with an OpenAPI spec that has `Servers` set. This may lead to an HTTP 400 with `no matching operation was found` when sending a valid request, as the validator performs `Host` header validation. If you're expecting `Host` header validation, you can silence this warning by setting `Options.SilenceServersWarning = true`. See https://github.com/oapi-codegen/oapi-codegen/issues/882 for more information.")
	}

	router, err := gorillamux.NewRouter(spec)
	if err != nil {
		panic(err)
	}

	skipper := getSkipperFromOptions(options)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if skipper(c) {
				return next(c)
			}

			err := validateRequestFromContext(c, router, options)
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

// validateRequestFromContext does the work of validating a request.
func validateRequestFromContext(ctx *echo.Context, router routers.Router, options *Options) *echo.HTTPError {
	req := ctx.Request()

	if options != nil && options.Prefix != "" {
		clone := req.Clone(req.Context())
		clone.URL.Path = strings.TrimPrefix(clone.URL.Path, options.Prefix)
		req = clone
	}

	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		if errors.Is(err, routers.ErrMethodNotAllowed) {
			return echo.NewHTTPError(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		}

		switch e := err.(type) {
		case *routers.RouteError:
			return echo.NewHTTPError(http.StatusNotFound, e.Reason)
		default:
			return echo.NewHTTPError(http.StatusInternalServerError,
				fmt.Sprintf("error validating route: %s", err.Error()))
		}
	}

	for k, v := range pathParams {
		if unescaped, err := url.PathUnescape(v); err == nil {
			pathParams[k] = unescaped
		}
	}

	validationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}

	requestContext := context.WithValue(context.Background(), EchoContextKey, ctx) //nolint:staticcheck

	if options != nil {
		validationInput.Options = &options.Options
		validationInput.ParamDecoder = options.ParamDecoder
		requestContext = context.WithValue(requestContext, UserDataKey, options.UserData) //nolint:staticcheck
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
			errorLines := strings.Split(e.Error(), "\n")
			return echo.NewHTTPError(http.StatusBadRequest, errorLines[0])
		case *openapi3filter.SecurityRequirementsError:
			for _, err := range e.Errors {
				httpErr, ok := err.(*echo.HTTPError)
				if ok {
					return httpErr
				}
			}
			return echo.NewHTTPError(http.StatusForbidden, e.Error())
		default:
			return echo.NewHTTPError(http.StatusInternalServerError,
				fmt.Sprintf("error validating request: %s", err))
		}
	}
	return nil
}

// GetEchoContext gets the echo context from within requests. It returns
// nil if not found or wrong type.
func GetEchoContext(c context.Context) *echo.Context {
	iface := c.Value(EchoContextKey)
	if iface == nil {
		return nil
	}
	eCtx, ok := iface.(*echo.Context)
	if !ok {
		return nil
	}
	return eCtx
}

// GetUserData gets the user data from the context.
func GetUserData(c context.Context) any {
	return c.Value(UserDataKey)
}

func getSkipperFromOptions(options *Options) echomiddleware.Skipper {
	if options == nil {
		return echomiddleware.DefaultSkipper
	}

	if options.Skipper == nil {
		return echomiddleware.DefaultSkipper
	}

	return options.Skipper
}

func getMultiErrorHandlerFromOptions(options *Options) MultiErrorHandler {
	if options == nil {
		return defaultMultiErrorHandler
	}

	if options.MultiErrorHandler == nil {
		return defaultMultiErrorHandler
	}

	return options.MultiErrorHandler
}

func defaultMultiErrorHandler(me openapi3.MultiError) *echo.HTTPError {
	return echo.NewHTTPError(http.StatusBadRequest, me.Error())
}
