// Package echo provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.0.0-00010101000000-000000000000 DO NOT EDIT.
package echo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	strictecho "github.com/oapi-codegen/runtime/strictmiddleware/echo"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (GET /get-multibody)
	GetGetMultibody(ctx echo.Context) error

	// (GET /object)
	GetObject(ctx echo.Context) error

	// (POST /post-multibody)
	PostPostMultibody(ctx echo.Context) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// GetGetMultibody converts echo context to params.
func (w *ServerInterfaceWrapper) GetGetMultibody(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.GetGetMultibody(ctx)
	return err
}

// GetObject converts echo context to params.
func (w *ServerInterfaceWrapper) GetObject(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.GetObject(ctx)
	return err
}

// PostPostMultibody converts echo context to params.
func (w *ServerInterfaceWrapper) PostPostMultibody(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.PostPostMultibody(ctx)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/get-multibody", wrapper.GetGetMultibody)
	router.GET(baseURL+"/object", wrapper.GetObject)
	router.POST(baseURL+"/post-multibody", wrapper.PostPostMultibody)

}

type GetGetMultibodyRequestObject struct {
}

type GetGetMultibodyResponseObject interface {
	VisitGetGetMultibodyResponse(w http.ResponseWriter) error
}

type GetGetMultibody200ApplicationLdPlusJSONProfilehttpswwwW3OrgnsactivitystreamsResponse string

func (response GetGetMultibody200ApplicationLdPlusJSONProfilehttpswwwW3OrgnsactivitystreamsResponse) VisitGetGetMultibodyResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type GetGetMultibody200ApplicationLdPlusJSONProfilehttpswwwW3Orgnsactivitystreams2Response string

func (response GetGetMultibody200ApplicationLdPlusJSONProfilehttpswwwW3Orgnsactivitystreams2Response) VisitGetGetMultibodyResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams2\"")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type GetObjectRequestObject struct {
}

type GetObjectResponseObject interface {
	VisitGetObjectResponse(w http.ResponseWriter) error
}

type GetObject200ApplicationLdPlusJSONProfilehttpswwwW3OrgnsactivitystreamsResponse string

func (response GetObject200ApplicationLdPlusJSONProfilehttpswwwW3OrgnsactivitystreamsResponse) VisitGetObjectResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type PostPostMultibodyRequestObject struct {
	ApplicationLdPlusJSONProfilehttpswwwW3OrgnsactivitystreamsBody  *PostPostMultibodyApplicationLdPlusJSONProfilehttpswwwW3OrgnsactivitystreamsRequestBody
	ApplicationLdPlusJSONProfilehttpswwwW3Orgnsactivitystreams2Body *PostPostMultibodyApplicationLdPlusJSONProfilehttpswwwW3Orgnsactivitystreams2RequestBody
}

type PostPostMultibodyResponseObject interface {
	VisitPostPostMultibodyResponse(w http.ResponseWriter) error
}

type PostPostMultibody200ApplicationLdPlusJSONProfilehttpswwwW3OrgnsactivitystreamsResponse string

func (response PostPostMultibody200ApplicationLdPlusJSONProfilehttpswwwW3OrgnsactivitystreamsResponse) VisitPostPostMultibodyResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type PostPostMultibody200ApplicationLdPlusJSONProfilehttpswwwW3Orgnsactivitystreams2Response string

func (response PostPostMultibody200ApplicationLdPlusJSONProfilehttpswwwW3Orgnsactivitystreams2Response) VisitPostPostMultibodyResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams2\"")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {

	// (GET /get-multibody)
	GetGetMultibody(ctx context.Context, request GetGetMultibodyRequestObject) (GetGetMultibodyResponseObject, error)

	// (GET /object)
	GetObject(ctx context.Context, request GetObjectRequestObject) (GetObjectResponseObject, error)

	// (POST /post-multibody)
	PostPostMultibody(ctx context.Context, request PostPostMultibodyRequestObject) (PostPostMultibodyResponseObject, error)
}

type StrictHandlerFunc = strictecho.StrictEchoHandlerFunc
type StrictMiddlewareFunc = strictecho.StrictEchoMiddlewareFunc

func NewStrictHandler(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares}
}

type strictHandler struct {
	ssi         StrictServerInterface
	middlewares []StrictMiddlewareFunc
}

// GetGetMultibody operation middleware
func (sh *strictHandler) GetGetMultibody(ctx echo.Context) error {
	var request GetGetMultibodyRequestObject

	handler := func(ctx echo.Context, request interface{}) (interface{}, error) {
		return sh.ssi.GetGetMultibody(ctx.Request().Context(), request.(GetGetMultibodyRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetGetMultibody")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return err
	} else if validResponse, ok := response.(GetGetMultibodyResponseObject); ok {
		return validResponse.VisitGetGetMultibodyResponse(ctx.Response())
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}

// GetObject operation middleware
func (sh *strictHandler) GetObject(ctx echo.Context) error {
	var request GetObjectRequestObject

	handler := func(ctx echo.Context, request interface{}) (interface{}, error) {
		return sh.ssi.GetObject(ctx.Request().Context(), request.(GetObjectRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetObject")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return err
	} else if validResponse, ok := response.(GetObjectResponseObject); ok {
		return validResponse.VisitGetObjectResponse(ctx.Response())
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}

// PostPostMultibody operation middleware
func (sh *strictHandler) PostPostMultibody(ctx echo.Context) error {
	var request PostPostMultibodyRequestObject

	if strings.HasPrefix(ctx.Request().Header.Get("Content-Type"), "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"") {
		var body PostPostMultibodyApplicationLdPlusJSONProfilehttpswwwW3OrgnsactivitystreamsRequestBody
		if err := ctx.Bind(&body); err != nil {
			return err
		}
		request.ApplicationLdPlusJSONProfilehttpswwwW3OrgnsactivitystreamsBody = &body
	}
	if strings.HasPrefix(ctx.Request().Header.Get("Content-Type"), "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams2\"") {
		var body PostPostMultibodyApplicationLdPlusJSONProfilehttpswwwW3Orgnsactivitystreams2RequestBody
		if err := ctx.Bind(&body); err != nil {
			return err
		}
		request.ApplicationLdPlusJSONProfilehttpswwwW3Orgnsactivitystreams2Body = &body
	}

	handler := func(ctx echo.Context, request interface{}) (interface{}, error) {
		return sh.ssi.PostPostMultibody(ctx.Request().Context(), request.(PostPostMultibodyRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "PostPostMultibody")
	}

	response, err := handler(ctx, request)

	if err != nil {
		return err
	} else if validResponse, ok := response.(PostPostMultibodyResponseObject); ok {
		return validResponse.VisitPostPostMultibodyResponse(ctx.Response())
	} else if response != nil {
		return fmt.Errorf("unexpected response type: %T", response)
	}
	return nil
}
