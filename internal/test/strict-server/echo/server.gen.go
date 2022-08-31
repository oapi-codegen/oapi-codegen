// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version (devel) DO NOT EDIT.
package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (POST /json)
	JSONExample(ctx echo.Context) error

	// (POST /multipart)
	MultipartExample(ctx echo.Context) error

	// (POST /multiple)
	MultipleRequestAndResponseTypes(ctx echo.Context) error

	// (POST /reusable-responses)
	ReusableResponses(ctx echo.Context) error

	// (POST /text)
	TextExample(ctx echo.Context) error

	// (POST /unknown)
	UnknownExample(ctx echo.Context) error

	// (POST /unspecified-content-type)
	UnspecifiedContentType(ctx echo.Context) error

	// (POST /urlencoded)
	URLEncodedExample(ctx echo.Context) error

	// (POST /with-headers)
	HeadersExample(ctx echo.Context, params HeadersExampleParams) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// JSONExample converts echo context to params.
func (w *ServerInterfaceWrapper) JSONExample(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.JSONExample(ctx)
	return err
}

// MultipartExample converts echo context to params.
func (w *ServerInterfaceWrapper) MultipartExample(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.MultipartExample(ctx)
	return err
}

// MultipleRequestAndResponseTypes converts echo context to params.
func (w *ServerInterfaceWrapper) MultipleRequestAndResponseTypes(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.MultipleRequestAndResponseTypes(ctx)
	return err
}

// ReusableResponses converts echo context to params.
func (w *ServerInterfaceWrapper) ReusableResponses(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ReusableResponses(ctx)
	return err
}

// TextExample converts echo context to params.
func (w *ServerInterfaceWrapper) TextExample(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.TextExample(ctx)
	return err
}

// UnknownExample converts echo context to params.
func (w *ServerInterfaceWrapper) UnknownExample(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.UnknownExample(ctx)
	return err
}

// UnspecifiedContentType converts echo context to params.
func (w *ServerInterfaceWrapper) UnspecifiedContentType(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.UnspecifiedContentType(ctx)
	return err
}

// URLEncodedExample converts echo context to params.
func (w *ServerInterfaceWrapper) URLEncodedExample(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.URLEncodedExample(ctx)
	return err
}

// HeadersExample converts echo context to params.
func (w *ServerInterfaceWrapper) HeadersExample(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params HeadersExampleParams

	headers := ctx.Request().Header
	// ------------- Required header parameter "header1" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("header1")]; found {
		var Header1 string
		n := len(valueList)
		if n != 1 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for header1, got %d", n))
		}

		err = runtime.BindStyledParameterWithLocation("simple", false, "header1", runtime.ParamLocationHeader, valueList[0], &Header1)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter header1: %s", err))
		}

		params.Header1 = Header1
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Header parameter header1 is required, but not found"))
	}
	// ------------- Optional header parameter "header2" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("header2")]; found {
		var Header2 int
		n := len(valueList)
		if n != 1 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Expected one value for header2, got %d", n))
		}

		err = runtime.BindStyledParameterWithLocation("simple", false, "header2", runtime.ParamLocationHeader, valueList[0], &Header2)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter header2: %s", err))
		}

		params.Header2 = &Header2
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.HeadersExample(ctx, params)
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

	router.POST(baseURL+"/json", wrapper.JSONExample)
	router.POST(baseURL+"/multipart", wrapper.MultipartExample)
	router.POST(baseURL+"/multiple", wrapper.MultipleRequestAndResponseTypes)
	router.POST(baseURL+"/reusable-responses", wrapper.ReusableResponses)
	router.POST(baseURL+"/text", wrapper.TextExample)
	router.POST(baseURL+"/unknown", wrapper.UnknownExample)
	router.POST(baseURL+"/unspecified-content-type", wrapper.UnspecifiedContentType)
	router.POST(baseURL+"/urlencoded", wrapper.URLEncodedExample)
	router.POST(baseURL+"/with-headers", wrapper.HeadersExample)

}

type BadrequestResponse struct {
}

type ReusableresponseResponseHeaders struct {
	Header1 string
	Header2 int
}
type ReusableresponseJSONResponse struct {
	Body Example

	Headers ReusableresponseResponseHeaders
}

func (t ReusableresponseJSONResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Body)
}

type JSONExampleRequestObject struct {
	Body *JSONExampleJSONRequestBody
}

type JSONExample200JSONResponse Example

func (t JSONExample200JSONResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((Example)(t))
}

type JSONExample400Response = BadrequestResponse

type JSONExampledefaultResponse struct {
	StatusCode int
}

type MultipartExampleRequestObject struct {
	Body *multipart.Reader
}

type MultipartExample200MultipartResponse func(writer *multipart.Writer) error

type MultipartExample400Response = BadrequestResponse

type MultipartExampledefaultResponse struct {
	StatusCode int
}

type MultipleRequestAndResponseTypesRequestObject struct {
	JSONBody      *MultipleRequestAndResponseTypesJSONRequestBody
	FormdataBody  *MultipleRequestAndResponseTypesFormdataRequestBody
	Body          io.Reader
	MultipartBody *multipart.Reader
	TextBody      *MultipleRequestAndResponseTypesTextRequestBody
}

type MultipleRequestAndResponseTypes200JSONResponse Example

func (t MultipleRequestAndResponseTypes200JSONResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((Example)(t))
}

type MultipleRequestAndResponseTypes200FormdataResponse Example

type MultipleRequestAndResponseTypes200ImagePngResponse struct {
	Body          io.Reader
	ContentLength int64
}

type MultipleRequestAndResponseTypes200MultipartResponse func(writer *multipart.Writer) error

type MultipleRequestAndResponseTypes200TextResponse string

type MultipleRequestAndResponseTypes400Response = BadrequestResponse

type ReusableResponsesRequestObject struct {
	Body *ReusableResponsesJSONRequestBody
}

type ReusableResponses200JSONResponse = ReusableresponseJSONResponse

type ReusableResponses400Response = BadrequestResponse

type ReusableResponsesdefaultResponse struct {
	StatusCode int
}

type TextExampleRequestObject struct {
	Body *TextExampleTextRequestBody
}

type TextExample200TextResponse string

type TextExample400Response = BadrequestResponse

type TextExampledefaultResponse struct {
	StatusCode int
}

type UnknownExampleRequestObject struct {
	Body io.Reader
}

type UnknownExample200VideoMp4Response struct {
	Body          io.Reader
	ContentLength int64
}

type UnknownExample400Response = BadrequestResponse

type UnknownExampledefaultResponse struct {
	StatusCode int
}

type UnspecifiedContentTypeRequestObject struct {
	ContentType string
	Body        io.Reader
}

type UnspecifiedContentType200VideoResponse struct {
	Body          io.Reader
	ContentType   string
	ContentLength int64
}

type UnspecifiedContentType400Response = BadrequestResponse

type UnspecifiedContentType401Response struct {
}

type UnspecifiedContentType403Response struct {
}

type UnspecifiedContentTypedefaultResponse struct {
	StatusCode int
}

type URLEncodedExampleRequestObject struct {
	Body *URLEncodedExampleFormdataRequestBody
}

type URLEncodedExample200FormdataResponse Example

type URLEncodedExample400Response = BadrequestResponse

type URLEncodedExampledefaultResponse struct {
	StatusCode int
}

type HeadersExampleRequestObject struct {
	Params HeadersExampleParams
	Body   *HeadersExampleJSONRequestBody
}

type HeadersExample200ResponseHeaders struct {
	Header1 string
	Header2 int
}

type HeadersExample200JSONResponse struct {
	Body    Example
	Headers HeadersExample200ResponseHeaders
}

func (t HeadersExample200JSONResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Body)
}

type HeadersExample400Response = BadrequestResponse

type HeadersExampledefaultResponse struct {
	StatusCode int
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {

	// (POST /json)
	JSONExample(ctx context.Context, request JSONExampleRequestObject) interface{}

	// (POST /multipart)
	MultipartExample(ctx context.Context, request MultipartExampleRequestObject) interface{}

	// (POST /multiple)
	MultipleRequestAndResponseTypes(ctx context.Context, request MultipleRequestAndResponseTypesRequestObject) interface{}

	// (POST /reusable-responses)
	ReusableResponses(ctx context.Context, request ReusableResponsesRequestObject) interface{}

	// (POST /text)
	TextExample(ctx context.Context, request TextExampleRequestObject) interface{}

	// (POST /unknown)
	UnknownExample(ctx context.Context, request UnknownExampleRequestObject) interface{}

	// (POST /unspecified-content-type)
	UnspecifiedContentType(ctx context.Context, request UnspecifiedContentTypeRequestObject) interface{}

	// (POST /urlencoded)
	URLEncodedExample(ctx context.Context, request URLEncodedExampleRequestObject) interface{}

	// (POST /with-headers)
	HeadersExample(ctx context.Context, request HeadersExampleRequestObject) interface{}
}

type StrictHandlerFunc func(ctx echo.Context, args interface{}) interface{}

type StrictMiddlewareFunc func(f StrictHandlerFunc, operationID string) StrictHandlerFunc

func NewStrictHandler(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares}
}

type strictHandler struct {
	ssi         StrictServerInterface
	middlewares []StrictMiddlewareFunc
}

// JSONExample operation middleware
func (sh *strictHandler) JSONExample(ctx echo.Context) error {
	var request JSONExampleRequestObject

	var body JSONExampleJSONRequestBody
	if err := ctx.Bind(&body); err != nil {
		return err
	}
	request.Body = &body

	handler := func(ctx echo.Context, request interface{}) interface{} {
		return sh.ssi.JSONExample(ctx.Request().Context(), request.(JSONExampleRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "JSONExample")
	}

	response := handler(ctx, request)

	switch v := response.(type) {
	case JSONExample200JSONResponse:
		return ctx.JSON(200, v)
	case JSONExample400Response:
		return ctx.NoContent(400)
	case JSONExampledefaultResponse:
		return ctx.NoContent(v.StatusCode)
	case error:
		return v
	case nil:
	default:
		return fmt.Errorf("Unexpected response type: %T", v)
	}
	return nil
}

// MultipartExample operation middleware
func (sh *strictHandler) MultipartExample(ctx echo.Context) error {
	var request MultipartExampleRequestObject

	if reader, err := ctx.Request().MultipartReader(); err != nil {
		return err
	} else {
		request.Body = reader
	}

	handler := func(ctx echo.Context, request interface{}) interface{} {
		return sh.ssi.MultipartExample(ctx.Request().Context(), request.(MultipartExampleRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "MultipartExample")
	}

	response := handler(ctx, request)

	switch v := response.(type) {
	case MultipartExample200MultipartResponse:
		writer := multipart.NewWriter(ctx.Response())
		ctx.Response().Header().Set("Content-Type", writer.FormDataContentType())
		defer writer.Close()
		if err := v(writer); err != nil {
			return err
		}
	case MultipartExample400Response:
		return ctx.NoContent(400)
	case MultipartExampledefaultResponse:
		return ctx.NoContent(v.StatusCode)
	case error:
		return v
	case nil:
	default:
		return fmt.Errorf("Unexpected response type: %T", v)
	}
	return nil
}

// MultipleRequestAndResponseTypes operation middleware
func (sh *strictHandler) MultipleRequestAndResponseTypes(ctx echo.Context) error {
	var request MultipleRequestAndResponseTypesRequestObject

	if strings.HasPrefix(ctx.Request().Header.Get("Content-Type"), "application/json") {
		var body MultipleRequestAndResponseTypesJSONRequestBody
		if err := ctx.Bind(&body); err != nil {
			return err
		}
		request.JSONBody = &body
	}
	if strings.HasPrefix(ctx.Request().Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		if form, err := ctx.FormParams(); err == nil {
			var body MultipleRequestAndResponseTypesFormdataRequestBody
			if err := runtime.BindForm(&body, form, nil, nil); err != nil {
				return err
			}
			request.FormdataBody = &body
		} else {
			return err
		}
	}
	if strings.HasPrefix(ctx.Request().Header.Get("Content-Type"), "image/png") {
		request.Body = ctx.Request().Body
	}
	if strings.HasPrefix(ctx.Request().Header.Get("Content-Type"), "multipart/form-data") {
		if reader, err := ctx.Request().MultipartReader(); err != nil {
			return err
		} else {
			request.MultipartBody = reader
		}
	}
	if strings.HasPrefix(ctx.Request().Header.Get("Content-Type"), "text/plain") {
		data, err := io.ReadAll(ctx.Request().Body)
		if err != nil {
			return err
		}
		body := MultipleRequestAndResponseTypesTextRequestBody(data)
		request.TextBody = &body
	}

	handler := func(ctx echo.Context, request interface{}) interface{} {
		return sh.ssi.MultipleRequestAndResponseTypes(ctx.Request().Context(), request.(MultipleRequestAndResponseTypesRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "MultipleRequestAndResponseTypes")
	}

	response := handler(ctx, request)

	switch v := response.(type) {
	case MultipleRequestAndResponseTypes200JSONResponse:
		return ctx.JSON(200, v)
	case MultipleRequestAndResponseTypes200FormdataResponse:
		if form, err := runtime.MarshalForm(v, nil); err != nil {
			return err
		} else {
			return ctx.Blob(200, "application/x-www-form-urlencoded", []byte(form.Encode()))
		}
	case MultipleRequestAndResponseTypes200ImagePngResponse:
		if v.ContentLength != 0 {
			ctx.Response().Header().Set("Content-Length", fmt.Sprint(v.ContentLength))
		}
		if closer, ok := v.Body.(io.ReadCloser); ok {
			defer closer.Close()
		}
		return ctx.Stream(200, "image/png", v.Body)
	case MultipleRequestAndResponseTypes200MultipartResponse:
		writer := multipart.NewWriter(ctx.Response())
		ctx.Response().Header().Set("Content-Type", writer.FormDataContentType())
		defer writer.Close()
		if err := v(writer); err != nil {
			return err
		}
	case MultipleRequestAndResponseTypes200TextResponse:
		return ctx.Blob(200, "text/plain", []byte(v))
	case MultipleRequestAndResponseTypes400Response:
		return ctx.NoContent(400)
	case error:
		return v
	case nil:
	default:
		return fmt.Errorf("Unexpected response type: %T", v)
	}
	return nil
}

// ReusableResponses operation middleware
func (sh *strictHandler) ReusableResponses(ctx echo.Context) error {
	var request ReusableResponsesRequestObject

	var body ReusableResponsesJSONRequestBody
	if err := ctx.Bind(&body); err != nil {
		return err
	}
	request.Body = &body

	handler := func(ctx echo.Context, request interface{}) interface{} {
		return sh.ssi.ReusableResponses(ctx.Request().Context(), request.(ReusableResponsesRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "ReusableResponses")
	}

	response := handler(ctx, request)

	switch v := response.(type) {
	case ReusableResponses200JSONResponse:
		ctx.Response().Header().Set("header1", fmt.Sprint(v.Headers.Header1))
		ctx.Response().Header().Set("header2", fmt.Sprint(v.Headers.Header2))
		return ctx.JSON(200, v)
	case ReusableResponses400Response:
		return ctx.NoContent(400)
	case ReusableResponsesdefaultResponse:
		return ctx.NoContent(v.StatusCode)
	case error:
		return v
	case nil:
	default:
		return fmt.Errorf("Unexpected response type: %T", v)
	}
	return nil
}

// TextExample operation middleware
func (sh *strictHandler) TextExample(ctx echo.Context) error {
	var request TextExampleRequestObject

	data, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		return err
	}
	body := TextExampleTextRequestBody(data)
	request.Body = &body

	handler := func(ctx echo.Context, request interface{}) interface{} {
		return sh.ssi.TextExample(ctx.Request().Context(), request.(TextExampleRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "TextExample")
	}

	response := handler(ctx, request)

	switch v := response.(type) {
	case TextExample200TextResponse:
		return ctx.Blob(200, "text/plain", []byte(v))
	case TextExample400Response:
		return ctx.NoContent(400)
	case TextExampledefaultResponse:
		return ctx.NoContent(v.StatusCode)
	case error:
		return v
	case nil:
	default:
		return fmt.Errorf("Unexpected response type: %T", v)
	}
	return nil
}

// UnknownExample operation middleware
func (sh *strictHandler) UnknownExample(ctx echo.Context) error {
	var request UnknownExampleRequestObject

	request.Body = ctx.Request().Body

	handler := func(ctx echo.Context, request interface{}) interface{} {
		return sh.ssi.UnknownExample(ctx.Request().Context(), request.(UnknownExampleRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "UnknownExample")
	}

	response := handler(ctx, request)

	switch v := response.(type) {
	case UnknownExample200VideoMp4Response:
		if v.ContentLength != 0 {
			ctx.Response().Header().Set("Content-Length", fmt.Sprint(v.ContentLength))
		}
		if closer, ok := v.Body.(io.ReadCloser); ok {
			defer closer.Close()
		}
		return ctx.Stream(200, "video/mp4", v.Body)
	case UnknownExample400Response:
		return ctx.NoContent(400)
	case UnknownExampledefaultResponse:
		return ctx.NoContent(v.StatusCode)
	case error:
		return v
	case nil:
	default:
		return fmt.Errorf("Unexpected response type: %T", v)
	}
	return nil
}

// UnspecifiedContentType operation middleware
func (sh *strictHandler) UnspecifiedContentType(ctx echo.Context) error {
	var request UnspecifiedContentTypeRequestObject

	request.ContentType = ctx.Request().Header.Get("Content-Type")

	request.Body = ctx.Request().Body

	handler := func(ctx echo.Context, request interface{}) interface{} {
		return sh.ssi.UnspecifiedContentType(ctx.Request().Context(), request.(UnspecifiedContentTypeRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "UnspecifiedContentType")
	}

	response := handler(ctx, request)

	switch v := response.(type) {
	case UnspecifiedContentType200VideoResponse:
		if v.ContentLength != 0 {
			ctx.Response().Header().Set("Content-Length", fmt.Sprint(v.ContentLength))
		}
		if closer, ok := v.Body.(io.ReadCloser); ok {
			defer closer.Close()
		}
		return ctx.Stream(200, v.ContentType, v.Body)
	case UnspecifiedContentType400Response:
		return ctx.NoContent(400)
	case UnspecifiedContentType401Response:
		return ctx.NoContent(401)
	case UnspecifiedContentType403Response:
		return ctx.NoContent(403)
	case UnspecifiedContentTypedefaultResponse:
		return ctx.NoContent(v.StatusCode)
	case error:
		return v
	case nil:
	default:
		return fmt.Errorf("Unexpected response type: %T", v)
	}
	return nil
}

// URLEncodedExample operation middleware
func (sh *strictHandler) URLEncodedExample(ctx echo.Context) error {
	var request URLEncodedExampleRequestObject

	if form, err := ctx.FormParams(); err == nil {
		var body URLEncodedExampleFormdataRequestBody
		if err := runtime.BindForm(&body, form, nil, nil); err != nil {
			return err
		}
		request.Body = &body
	} else {
		return err
	}

	handler := func(ctx echo.Context, request interface{}) interface{} {
		return sh.ssi.URLEncodedExample(ctx.Request().Context(), request.(URLEncodedExampleRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "URLEncodedExample")
	}

	response := handler(ctx, request)

	switch v := response.(type) {
	case URLEncodedExample200FormdataResponse:
		if form, err := runtime.MarshalForm(v, nil); err != nil {
			return err
		} else {
			return ctx.Blob(200, "application/x-www-form-urlencoded", []byte(form.Encode()))
		}
	case URLEncodedExample400Response:
		return ctx.NoContent(400)
	case URLEncodedExampledefaultResponse:
		return ctx.NoContent(v.StatusCode)
	case error:
		return v
	case nil:
	default:
		return fmt.Errorf("Unexpected response type: %T", v)
	}
	return nil
}

// HeadersExample operation middleware
func (sh *strictHandler) HeadersExample(ctx echo.Context, params HeadersExampleParams) error {
	var request HeadersExampleRequestObject

	request.Params = params

	var body HeadersExampleJSONRequestBody
	if err := ctx.Bind(&body); err != nil {
		return err
	}
	request.Body = &body

	handler := func(ctx echo.Context, request interface{}) interface{} {
		return sh.ssi.HeadersExample(ctx.Request().Context(), request.(HeadersExampleRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "HeadersExample")
	}

	response := handler(ctx, request)

	switch v := response.(type) {
	case HeadersExample200JSONResponse:
		ctx.Response().Header().Set("header1", fmt.Sprint(v.Headers.Header1))
		ctx.Response().Header().Set("header2", fmt.Sprint(v.Headers.Header2))
		return ctx.JSON(200, v)
	case HeadersExample400Response:
		return ctx.NoContent(400)
	case HeadersExampledefaultResponse:
		return ctx.NoContent(v.StatusCode)
	case error:
		return v
	case nil:
	default:
		return fmt.Errorf("Unexpected response type: %T", v)
	}
	return nil
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/+xYS4/iOBD+K1btnkaB0D194rbTGmnfI9Ezp9UcirgAzya2166QRoj/vnJsaBjSCFo8",
	"pNXeEqde/qq+KsdLKExljSbNHoZLcOSt0Z7alzFKR//U5Dm8SfKFU5aV0TCEDyhH6dsqA0e1x3FJa/Ug",
	"XxjNpFtVtLZUBQbV/JsP+kvwxYwqDE8/OprAEH7IX0LJ41ef0zNWtiRYrVbZdxF8+g0ymBFKcm208fFu",
	"1zYvLMEQPDulpxCMRLH7TjGlmabkgrcgmoIIAus4hkuwzlhyrCJGcyxr6vaUVsz4GxUcd6D0xOxj+Wg0",
	"o9JeSDWZkCPNIoEngg0vfG2tcUxSjBcieChYeHJzcpABKw6BwdP2ukgBe8hgTs5HR3f9QX8Q8mUsabQK",
	"hvC+XcrAIs/aDW0SZE1X3n99+vSnUF5gzaZCVgWW5UJU6PwMy5KkUJpNiLEu2PehdeXazP8ik/rHhGUo",
	"m7aCPhi5uETFtIW5Vc/3g8GVCnOVwUN01mVjE1S+xbDWzATrsgP0L/pvbRotyDnj0s7yqi5ZWXS8naxd",
	"tP9YixwD+cZePjGu6klkvBDq5/J0U+BTM+gkydPMNF7MTCPYCElYikbxTKwVv2O30gKFV3paklgHlXVm",
	"sqTUc3/ScpT28jnYuDiXsh0rz72maXpt8mpXki6MJPk2s6rCKeVWT3fVg21kGMJ4waFs97vrmYooA6Zn",
	"zm2JSh8eHVdqJ/8jfTZiR7quzya9neR1E3dNKi8K1GIc+DjxgcRdvvZIOkqeRlsStxlxhzHaO61do2uG",
	"5L8+qT7T81FD6oxkvXY1ngpYHRdfxyxpHQPbG7l/BIpzJcnklX040fLNQPWWCjVRJHtpF70Y22st4dHo",
	"whHvDu1wAtaGxcZYOJjzjEREIBPeiIZEVXsWFr0XitsuUqp4uJe01zy+vET2GD2FyX5EVt9dKKfvbpXR",
	"h8Hd6SrvL1w3O8P3FT6Ofv8YZU79wznblD/xjHI+vzeiczhW97buALop/HMUeJnpBak5SYFaCkdcO01S",
	"zBWuf1v3uJkMvKTVosOKuPX61xLCCEkXC5CBxoo273epCJQLyLKrKTt0PXHQ1j1kh+4svv6Hf6gvedNz",
	"6TpdZRAvZWKx1K4MGWW2wzyPlzl93+B0Sq6vTI5Wwerr6t8AAAD//ygSomqZEwAA",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
