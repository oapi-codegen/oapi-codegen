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
	"github.com/gin-gonic/gin"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (POST /json)
	JSONExample(c *gin.Context)

	// (POST /multipart)
	MultipartExample(c *gin.Context)

	// (POST /multiple)
	MultipleRequestAndResponseTypes(c *gin.Context)

	// (GET /reserved-go-keyword-parameters/{type})
	ReservedGoKeywordParameters(c *gin.Context, pType string)

	// (POST /reusable-responses)
	ReusableResponses(c *gin.Context)

	// (POST /text)
	TextExample(c *gin.Context)

	// (POST /unknown)
	UnknownExample(c *gin.Context)

	// (POST /unspecified-content-type)
	UnspecifiedContentType(c *gin.Context)

	// (POST /urlencoded)
	URLEncodedExample(c *gin.Context)

	// (POST /with-headers)
	HeadersExample(c *gin.Context, params HeadersExampleParams)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandler       func(*gin.Context, error, int)
}

type MiddlewareFunc func(c *gin.Context)

// JSONExample operation middleware
func (siw *ServerInterfaceWrapper) JSONExample(c *gin.Context) {

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.JSONExample(c)
}

// MultipartExample operation middleware
func (siw *ServerInterfaceWrapper) MultipartExample(c *gin.Context) {

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.MultipartExample(c)
}

// MultipleRequestAndResponseTypes operation middleware
func (siw *ServerInterfaceWrapper) MultipleRequestAndResponseTypes(c *gin.Context) {

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.MultipleRequestAndResponseTypes(c)
}

// ReservedGoKeywordParameters operation middleware
func (siw *ServerInterfaceWrapper) ReservedGoKeywordParameters(c *gin.Context) {

	var err error

	// ------------- Path parameter "type" -------------
	var pType string

	err = runtime.BindStyledParameter("simple", false, "type", c.Param("type"), &pType)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter type: %s", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.ReservedGoKeywordParameters(c, pType)
}

// ReusableResponses operation middleware
func (siw *ServerInterfaceWrapper) ReusableResponses(c *gin.Context) {

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.ReusableResponses(c)
}

// TextExample operation middleware
func (siw *ServerInterfaceWrapper) TextExample(c *gin.Context) {

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.TextExample(c)
}

// UnknownExample operation middleware
func (siw *ServerInterfaceWrapper) UnknownExample(c *gin.Context) {

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.UnknownExample(c)
}

// UnspecifiedContentType operation middleware
func (siw *ServerInterfaceWrapper) UnspecifiedContentType(c *gin.Context) {

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.UnspecifiedContentType(c)
}

// URLEncodedExample operation middleware
func (siw *ServerInterfaceWrapper) URLEncodedExample(c *gin.Context) {

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.URLEncodedExample(c)
}

// HeadersExample operation middleware
func (siw *ServerInterfaceWrapper) HeadersExample(c *gin.Context) {

	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params HeadersExampleParams

	headers := c.Request.Header

	// ------------- Required header parameter "header1" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("header1")]; found {
		var Header1 string
		n := len(valueList)
		if n != 1 {
			siw.ErrorHandler(c, fmt.Errorf("Expected one value for header1, got %d", n), http.StatusBadRequest)
			return
		}

		err = runtime.BindStyledParameterWithLocation("simple", false, "header1", runtime.ParamLocationHeader, valueList[0], &Header1)
		if err != nil {
			siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter header1: %s", err), http.StatusBadRequest)
			return
		}

		params.Header1 = Header1

	} else {
		siw.ErrorHandler(c, fmt.Errorf("Header parameter header1 is required, but not found: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional header parameter "header2" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("header2")]; found {
		var Header2 int
		n := len(valueList)
		if n != 1 {
			siw.ErrorHandler(c, fmt.Errorf("Expected one value for header2, got %d", n), http.StatusBadRequest)
			return
		}

		err = runtime.BindStyledParameterWithLocation("simple", false, "header2", runtime.ParamLocationHeader, valueList[0], &Header2)
		if err != nil {
			siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter header2: %s", err), http.StatusBadRequest)
			return
		}

		params.Header2 = &Header2

	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.HeadersExample(c, params)
}

// GinServerOptions provides options for the Gin server.
type GinServerOptions struct {
	BaseURL      string
	Middlewares  []MiddlewareFunc
	ErrorHandler func(*gin.Context, error, int)
}

// RegisterHandlers creates http.Handler with routing matching OpenAPI spec.
func RegisterHandlers(router gin.IRouter, si ServerInterface) {
	RegisterHandlersWithOptions(router, si, GinServerOptions{})
}

// RegisterHandlersWithOptions creates http.Handler with additional options
func RegisterHandlersWithOptions(router gin.IRouter, si ServerInterface, options GinServerOptions) {
	errorHandler := options.ErrorHandler
	if errorHandler == nil {
		errorHandler = func(c *gin.Context, err error, statusCode int) {
			c.JSON(statusCode, gin.H{"msg": err.Error()})
		}
	}

	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandler:       errorHandler,
	}

	router.POST(options.BaseURL+"/json", wrapper.JSONExample)
	router.POST(options.BaseURL+"/multipart", wrapper.MultipartExample)
	router.POST(options.BaseURL+"/multiple", wrapper.MultipleRequestAndResponseTypes)
	router.GET(options.BaseURL+"/reserved-go-keyword-parameters/:type", wrapper.ReservedGoKeywordParameters)
	router.POST(options.BaseURL+"/reusable-responses", wrapper.ReusableResponses)
	router.POST(options.BaseURL+"/text", wrapper.TextExample)
	router.POST(options.BaseURL+"/unknown", wrapper.UnknownExample)
	router.POST(options.BaseURL+"/unspecified-content-type", wrapper.UnspecifiedContentType)
	router.POST(options.BaseURL+"/urlencoded", wrapper.URLEncodedExample)
	router.POST(options.BaseURL+"/with-headers", wrapper.HeadersExample)
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

type JSONExampleRequestObject struct {
	Body *JSONExampleJSONRequestBody
}

type JSONExampleResponseObject interface {
	VisitJSONExampleResponse(w http.ResponseWriter) error
}

type JSONExample200JSONResponse Example

func (response JSONExample200JSONResponse) VisitJSONExampleResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type JSONExample400Response = BadrequestResponse

func (response JSONExample400Response) VisitJSONExampleResponse(w http.ResponseWriter) error {
	w.WriteHeader(400)
	return nil
}

type JSONExampledefaultResponse struct {
	StatusCode int
}

func (response JSONExampledefaultResponse) VisitJSONExampleResponse(w http.ResponseWriter) error {
	w.WriteHeader(response.StatusCode)
	return nil
}

type MultipartExampleRequestObject struct {
	Body *multipart.Reader
}

type MultipartExampleResponseObject interface {
	VisitMultipartExampleResponse(w http.ResponseWriter) error
}

type MultipartExample200MultipartResponse func(writer *multipart.Writer) error

func (response MultipartExample200MultipartResponse) VisitMultipartExampleResponse(w http.ResponseWriter) error {
	writer := multipart.NewWriter(w)
	w.Header().Set("Content-Type", writer.FormDataContentType())
	w.WriteHeader(200)

	defer writer.Close()
	return response(writer)
}

type MultipartExample400Response = BadrequestResponse

func (response MultipartExample400Response) VisitMultipartExampleResponse(w http.ResponseWriter) error {
	w.WriteHeader(400)
	return nil
}

type MultipartExampledefaultResponse struct {
	StatusCode int
}

func (response MultipartExampledefaultResponse) VisitMultipartExampleResponse(w http.ResponseWriter) error {
	w.WriteHeader(response.StatusCode)
	return nil
}

type MultipleRequestAndResponseTypesRequestObject struct {
	JSONBody      *MultipleRequestAndResponseTypesJSONRequestBody
	FormdataBody  *MultipleRequestAndResponseTypesFormdataRequestBody
	Body          io.Reader
	MultipartBody *multipart.Reader
	TextBody      *MultipleRequestAndResponseTypesTextRequestBody
}

type MultipleRequestAndResponseTypesResponseObject interface {
	VisitMultipleRequestAndResponseTypesResponse(w http.ResponseWriter) error
}

type MultipleRequestAndResponseTypes200JSONResponse Example

func (response MultipleRequestAndResponseTypes200JSONResponse) VisitMultipleRequestAndResponseTypesResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type MultipleRequestAndResponseTypes200FormdataResponse Example

func (response MultipleRequestAndResponseTypes200FormdataResponse) VisitMultipleRequestAndResponseTypesResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	w.WriteHeader(200)

	if form, err := runtime.MarshalForm(response, nil); err != nil {
		return err
	} else {
		_, err := w.Write([]byte(form.Encode()))
		return err
	}
}

type MultipleRequestAndResponseTypes200ImagepngResponse struct {
	Body          io.Reader
	ContentLength int64
}

func (response MultipleRequestAndResponseTypes200ImagepngResponse) VisitMultipleRequestAndResponseTypesResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "image/png")
	if response.ContentLength != 0 {
		w.Header().Set("Content-Length", fmt.Sprint(response.ContentLength))
	}
	w.WriteHeader(200)

	if closer, ok := response.Body.(io.ReadCloser); ok {
		defer closer.Close()
	}
	_, err := io.Copy(w, response.Body)
	return err
}

type MultipleRequestAndResponseTypes200MultipartResponse func(writer *multipart.Writer) error

func (response MultipleRequestAndResponseTypes200MultipartResponse) VisitMultipleRequestAndResponseTypesResponse(w http.ResponseWriter) error {
	writer := multipart.NewWriter(w)
	w.Header().Set("Content-Type", writer.FormDataContentType())
	w.WriteHeader(200)

	defer writer.Close()
	return response(writer)
}

type MultipleRequestAndResponseTypes200TextResponse string

func (response MultipleRequestAndResponseTypes200TextResponse) VisitMultipleRequestAndResponseTypesResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)

	_, err := w.Write([]byte(response))
	return err
}

type MultipleRequestAndResponseTypes400Response = BadrequestResponse

func (response MultipleRequestAndResponseTypes400Response) VisitMultipleRequestAndResponseTypesResponse(w http.ResponseWriter) error {
	w.WriteHeader(400)
	return nil
}

type ReservedGoKeywordParametersRequestObject struct {
	Type string `json:"type"`
}

type ReservedGoKeywordParametersResponseObject interface {
	VisitReservedGoKeywordParametersResponse(w http.ResponseWriter) error
}

type ReservedGoKeywordParameters200TextResponse string

func (response ReservedGoKeywordParameters200TextResponse) VisitReservedGoKeywordParametersResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)

	_, err := w.Write([]byte(response))
	return err
}

type ReusableResponsesRequestObject struct {
	Body *ReusableResponsesJSONRequestBody
}

type ReusableResponsesResponseObject interface {
	VisitReusableResponsesResponse(w http.ResponseWriter) error
}

type ReusableResponses200JSONResponse struct{ ReusableresponseJSONResponse }

func (response ReusableResponses200JSONResponse) VisitReusableResponsesResponse(w http.ResponseWriter) error {
	w.Header().Set("header1", fmt.Sprint(response.Headers.Header1))
	w.Header().Set("header2", fmt.Sprint(response.Headers.Header2))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response.Body)
}

type ReusableResponses400Response = BadrequestResponse

func (response ReusableResponses400Response) VisitReusableResponsesResponse(w http.ResponseWriter) error {
	w.WriteHeader(400)
	return nil
}

type ReusableResponsesdefaultResponse struct {
	StatusCode int
}

func (response ReusableResponsesdefaultResponse) VisitReusableResponsesResponse(w http.ResponseWriter) error {
	w.WriteHeader(response.StatusCode)
	return nil
}

type TextExampleRequestObject struct {
	Body *TextExampleTextRequestBody
}

type TextExampleResponseObject interface {
	VisitTextExampleResponse(w http.ResponseWriter) error
}

type TextExample200TextResponse string

func (response TextExample200TextResponse) VisitTextExampleResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)

	_, err := w.Write([]byte(response))
	return err
}

type TextExample400Response = BadrequestResponse

func (response TextExample400Response) VisitTextExampleResponse(w http.ResponseWriter) error {
	w.WriteHeader(400)
	return nil
}

type TextExampledefaultResponse struct {
	StatusCode int
}

func (response TextExampledefaultResponse) VisitTextExampleResponse(w http.ResponseWriter) error {
	w.WriteHeader(response.StatusCode)
	return nil
}

type UnknownExampleRequestObject struct {
	Body io.Reader
}

type UnknownExampleResponseObject interface {
	VisitUnknownExampleResponse(w http.ResponseWriter) error
}

type UnknownExample200Videomp4Response struct {
	Body          io.Reader
	ContentLength int64
}

func (response UnknownExample200Videomp4Response) VisitUnknownExampleResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "video/mp4")
	if response.ContentLength != 0 {
		w.Header().Set("Content-Length", fmt.Sprint(response.ContentLength))
	}
	w.WriteHeader(200)

	if closer, ok := response.Body.(io.ReadCloser); ok {
		defer closer.Close()
	}
	_, err := io.Copy(w, response.Body)
	return err
}

type UnknownExample400Response = BadrequestResponse

func (response UnknownExample400Response) VisitUnknownExampleResponse(w http.ResponseWriter) error {
	w.WriteHeader(400)
	return nil
}

type UnknownExampledefaultResponse struct {
	StatusCode int
}

func (response UnknownExampledefaultResponse) VisitUnknownExampleResponse(w http.ResponseWriter) error {
	w.WriteHeader(response.StatusCode)
	return nil
}

type UnspecifiedContentTypeRequestObject struct {
	ContentType string
	Body        io.Reader
}

type UnspecifiedContentTypeResponseObject interface {
	VisitUnspecifiedContentTypeResponse(w http.ResponseWriter) error
}

type UnspecifiedContentType200VideoResponse struct {
	Body          io.Reader
	ContentType   string
	ContentLength int64
}

func (response UnspecifiedContentType200VideoResponse) VisitUnspecifiedContentTypeResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", response.ContentType)
	if response.ContentLength != 0 {
		w.Header().Set("Content-Length", fmt.Sprint(response.ContentLength))
	}
	w.WriteHeader(200)

	if closer, ok := response.Body.(io.ReadCloser); ok {
		defer closer.Close()
	}
	_, err := io.Copy(w, response.Body)
	return err
}

type UnspecifiedContentType400Response = BadrequestResponse

func (response UnspecifiedContentType400Response) VisitUnspecifiedContentTypeResponse(w http.ResponseWriter) error {
	w.WriteHeader(400)
	return nil
}

type UnspecifiedContentType401Response struct {
}

func (response UnspecifiedContentType401Response) VisitUnspecifiedContentTypeResponse(w http.ResponseWriter) error {
	w.WriteHeader(401)
	return nil
}

type UnspecifiedContentType403Response struct {
}

func (response UnspecifiedContentType403Response) VisitUnspecifiedContentTypeResponse(w http.ResponseWriter) error {
	w.WriteHeader(403)
	return nil
}

type UnspecifiedContentTypedefaultResponse struct {
	StatusCode int
}

func (response UnspecifiedContentTypedefaultResponse) VisitUnspecifiedContentTypeResponse(w http.ResponseWriter) error {
	w.WriteHeader(response.StatusCode)
	return nil
}

type URLEncodedExampleRequestObject struct {
	Body *URLEncodedExampleFormdataRequestBody
}

type URLEncodedExampleResponseObject interface {
	VisitURLEncodedExampleResponse(w http.ResponseWriter) error
}

type URLEncodedExample200FormdataResponse Example

func (response URLEncodedExample200FormdataResponse) VisitURLEncodedExampleResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	w.WriteHeader(200)

	if form, err := runtime.MarshalForm(response, nil); err != nil {
		return err
	} else {
		_, err := w.Write([]byte(form.Encode()))
		return err
	}
}

type URLEncodedExample400Response = BadrequestResponse

func (response URLEncodedExample400Response) VisitURLEncodedExampleResponse(w http.ResponseWriter) error {
	w.WriteHeader(400)
	return nil
}

type URLEncodedExampledefaultResponse struct {
	StatusCode int
}

func (response URLEncodedExampledefaultResponse) VisitURLEncodedExampleResponse(w http.ResponseWriter) error {
	w.WriteHeader(response.StatusCode)
	return nil
}

type HeadersExampleRequestObject struct {
	Params HeadersExampleParams
	Body   *HeadersExampleJSONRequestBody
}

type HeadersExampleResponseObject interface {
	VisitHeadersExampleResponse(w http.ResponseWriter) error
}

type HeadersExample200ResponseHeaders struct {
	Header1 string
	Header2 int
}

type HeadersExample200JSONResponse struct {
	Body    Example
	Headers HeadersExample200ResponseHeaders
}

func (response HeadersExample200JSONResponse) VisitHeadersExampleResponse(w http.ResponseWriter) error {
	w.Header().Set("header1", fmt.Sprint(response.Headers.Header1))
	w.Header().Set("header2", fmt.Sprint(response.Headers.Header2))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response.Body)
}

type HeadersExample400Response = BadrequestResponse

func (response HeadersExample400Response) VisitHeadersExampleResponse(w http.ResponseWriter) error {
	w.WriteHeader(400)
	return nil
}

type HeadersExampledefaultResponse struct {
	StatusCode int
}

func (response HeadersExampledefaultResponse) VisitHeadersExampleResponse(w http.ResponseWriter) error {
	w.WriteHeader(response.StatusCode)
	return nil
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {

	// (POST /json)
	JSONExample(ctx context.Context, request JSONExampleRequestObject) (JSONExampleResponseObject, error)

	// (POST /multipart)
	MultipartExample(ctx context.Context, request MultipartExampleRequestObject) (MultipartExampleResponseObject, error)

	// (POST /multiple)
	MultipleRequestAndResponseTypes(ctx context.Context, request MultipleRequestAndResponseTypesRequestObject) (MultipleRequestAndResponseTypesResponseObject, error)

	// (GET /reserved-go-keyword-parameters/{type})
	ReservedGoKeywordParameters(ctx context.Context, request ReservedGoKeywordParametersRequestObject) (ReservedGoKeywordParametersResponseObject, error)

	// (POST /reusable-responses)
	ReusableResponses(ctx context.Context, request ReusableResponsesRequestObject) (ReusableResponsesResponseObject, error)

	// (POST /text)
	TextExample(ctx context.Context, request TextExampleRequestObject) (TextExampleResponseObject, error)

	// (POST /unknown)
	UnknownExample(ctx context.Context, request UnknownExampleRequestObject) (UnknownExampleResponseObject, error)

	// (POST /unspecified-content-type)
	UnspecifiedContentType(ctx context.Context, request UnspecifiedContentTypeRequestObject) (UnspecifiedContentTypeResponseObject, error)

	// (POST /urlencoded)
	URLEncodedExample(ctx context.Context, request URLEncodedExampleRequestObject) (URLEncodedExampleResponseObject, error)

	// (POST /with-headers)
	HeadersExample(ctx context.Context, request HeadersExampleRequestObject) (HeadersExampleResponseObject, error)
}

type StrictHandlerFunc func(ctx *gin.Context, args interface{}) (interface{}, error)

type StrictMiddlewareFunc func(f StrictHandlerFunc, operationID string) StrictHandlerFunc

func NewStrictHandler(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares}
}

type strictHandler struct {
	ssi         StrictServerInterface
	middlewares []StrictMiddlewareFunc
}

// JSONExample operation middleware
func (sh *strictHandler) JSONExample(ctx *gin.Context) {
	var request JSONExampleRequestObject

	var body JSONExampleJSONRequestBody
	if err := ctx.ShouldBind(&body); err != nil {
		ctx.Status(http.StatusBadRequest)
		ctx.Error(err)
		return
	}
	request.Body = &body

	handler := func(ctx *gin.Context, request interface{}) (interface{}, error) {
		return sh.ssi.JSONExample(ctx, request.(JSONExampleRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "JSONExample")
	}

	response, err := handler(ctx, request)

	if err != nil {
		ctx.Error(err)
		ctx.Status(http.StatusInternalServerError)
	} else if validResponse, ok := response.(JSONExampleResponseObject); ok {
		if err := validResponse.VisitJSONExampleResponse(ctx.Writer); err != nil {
			ctx.Error(err)
		}
	} else if response != nil {
		ctx.Error(fmt.Errorf("Unexpected response type: %T", response))
	}
}

// MultipartExample operation middleware
func (sh *strictHandler) MultipartExample(ctx *gin.Context) {
	var request MultipartExampleRequestObject

	if reader, err := ctx.Request.MultipartReader(); err == nil {
		request.Body = reader
	} else {
		ctx.Error(err)
		return
	}

	handler := func(ctx *gin.Context, request interface{}) (interface{}, error) {
		return sh.ssi.MultipartExample(ctx, request.(MultipartExampleRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "MultipartExample")
	}

	response, err := handler(ctx, request)

	if err != nil {
		ctx.Error(err)
		ctx.Status(http.StatusInternalServerError)
	} else if validResponse, ok := response.(MultipartExampleResponseObject); ok {
		if err := validResponse.VisitMultipartExampleResponse(ctx.Writer); err != nil {
			ctx.Error(err)
		}
	} else if response != nil {
		ctx.Error(fmt.Errorf("Unexpected response type: %T", response))
	}
}

// MultipleRequestAndResponseTypes operation middleware
func (sh *strictHandler) MultipleRequestAndResponseTypes(ctx *gin.Context) {
	var request MultipleRequestAndResponseTypesRequestObject

	if strings.HasPrefix(ctx.GetHeader("Content-Type"), "application/json") {
		var body MultipleRequestAndResponseTypesJSONRequestBody
		if err := ctx.ShouldBind(&body); err != nil {
			ctx.Status(http.StatusBadRequest)
			ctx.Error(err)
			return
		}
		request.JSONBody = &body
	}
	if strings.HasPrefix(ctx.GetHeader("Content-Type"), "application/x-www-form-urlencoded") {
		if err := ctx.Request.ParseForm(); err != nil {
			ctx.Error(err)
			return
		}
		var body MultipleRequestAndResponseTypesFormdataRequestBody
		if err := runtime.BindForm(&body, ctx.Request.Form, nil, nil); err != nil {
			ctx.Error(err)
			return
		}
		request.FormdataBody = &body
	}
	if strings.HasPrefix(ctx.GetHeader("Content-Type"), "image/png") {
		request.Body = ctx.Request.Body
	}
	if strings.HasPrefix(ctx.GetHeader("Content-Type"), "multipart/form-data") {
		if reader, err := ctx.Request.MultipartReader(); err == nil {
			request.MultipartBody = reader
		} else {
			ctx.Error(err)
			return
		}
	}
	if strings.HasPrefix(ctx.GetHeader("Content-Type"), "text/plain") {
		data, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.Error(err)
			return
		}
		body := MultipleRequestAndResponseTypesTextRequestBody(data)
		request.TextBody = &body
	}

	handler := func(ctx *gin.Context, request interface{}) (interface{}, error) {
		return sh.ssi.MultipleRequestAndResponseTypes(ctx, request.(MultipleRequestAndResponseTypesRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "MultipleRequestAndResponseTypes")
	}

	response, err := handler(ctx, request)

	if err != nil {
		ctx.Error(err)
		ctx.Status(http.StatusInternalServerError)
	} else if validResponse, ok := response.(MultipleRequestAndResponseTypesResponseObject); ok {
		if err := validResponse.VisitMultipleRequestAndResponseTypesResponse(ctx.Writer); err != nil {
			ctx.Error(err)
		}
	} else if response != nil {
		ctx.Error(fmt.Errorf("Unexpected response type: %T", response))
	}
}

// ReservedGoKeywordParameters operation middleware
func (sh *strictHandler) ReservedGoKeywordParameters(ctx *gin.Context, pType string) {
	var request ReservedGoKeywordParametersRequestObject

	request.Type = pType

	handler := func(ctx *gin.Context, request interface{}) (interface{}, error) {
		return sh.ssi.ReservedGoKeywordParameters(ctx, request.(ReservedGoKeywordParametersRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "ReservedGoKeywordParameters")
	}

	response, err := handler(ctx, request)

	if err != nil {
		ctx.Error(err)
		ctx.Status(http.StatusInternalServerError)
	} else if validResponse, ok := response.(ReservedGoKeywordParametersResponseObject); ok {
		if err := validResponse.VisitReservedGoKeywordParametersResponse(ctx.Writer); err != nil {
			ctx.Error(err)
		}
	} else if response != nil {
		ctx.Error(fmt.Errorf("Unexpected response type: %T", response))
	}
}

// ReusableResponses operation middleware
func (sh *strictHandler) ReusableResponses(ctx *gin.Context) {
	var request ReusableResponsesRequestObject

	var body ReusableResponsesJSONRequestBody
	if err := ctx.ShouldBind(&body); err != nil {
		ctx.Status(http.StatusBadRequest)
		ctx.Error(err)
		return
	}
	request.Body = &body

	handler := func(ctx *gin.Context, request interface{}) (interface{}, error) {
		return sh.ssi.ReusableResponses(ctx, request.(ReusableResponsesRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "ReusableResponses")
	}

	response, err := handler(ctx, request)

	if err != nil {
		ctx.Error(err)
		ctx.Status(http.StatusInternalServerError)
	} else if validResponse, ok := response.(ReusableResponsesResponseObject); ok {
		if err := validResponse.VisitReusableResponsesResponse(ctx.Writer); err != nil {
			ctx.Error(err)
		}
	} else if response != nil {
		ctx.Error(fmt.Errorf("Unexpected response type: %T", response))
	}
}

// TextExample operation middleware
func (sh *strictHandler) TextExample(ctx *gin.Context) {
	var request TextExampleRequestObject

	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Error(err)
		return
	}
	body := TextExampleTextRequestBody(data)
	request.Body = &body

	handler := func(ctx *gin.Context, request interface{}) (interface{}, error) {
		return sh.ssi.TextExample(ctx, request.(TextExampleRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "TextExample")
	}

	response, err := handler(ctx, request)

	if err != nil {
		ctx.Error(err)
		ctx.Status(http.StatusInternalServerError)
	} else if validResponse, ok := response.(TextExampleResponseObject); ok {
		if err := validResponse.VisitTextExampleResponse(ctx.Writer); err != nil {
			ctx.Error(err)
		}
	} else if response != nil {
		ctx.Error(fmt.Errorf("Unexpected response type: %T", response))
	}
}

// UnknownExample operation middleware
func (sh *strictHandler) UnknownExample(ctx *gin.Context) {
	var request UnknownExampleRequestObject

	request.Body = ctx.Request.Body

	handler := func(ctx *gin.Context, request interface{}) (interface{}, error) {
		return sh.ssi.UnknownExample(ctx, request.(UnknownExampleRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "UnknownExample")
	}

	response, err := handler(ctx, request)

	if err != nil {
		ctx.Error(err)
		ctx.Status(http.StatusInternalServerError)
	} else if validResponse, ok := response.(UnknownExampleResponseObject); ok {
		if err := validResponse.VisitUnknownExampleResponse(ctx.Writer); err != nil {
			ctx.Error(err)
		}
	} else if response != nil {
		ctx.Error(fmt.Errorf("Unexpected response type: %T", response))
	}
}

// UnspecifiedContentType operation middleware
func (sh *strictHandler) UnspecifiedContentType(ctx *gin.Context) {
	var request UnspecifiedContentTypeRequestObject

	request.ContentType = ctx.ContentType()

	request.Body = ctx.Request.Body

	handler := func(ctx *gin.Context, request interface{}) (interface{}, error) {
		return sh.ssi.UnspecifiedContentType(ctx, request.(UnspecifiedContentTypeRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "UnspecifiedContentType")
	}

	response, err := handler(ctx, request)

	if err != nil {
		ctx.Error(err)
		ctx.Status(http.StatusInternalServerError)
	} else if validResponse, ok := response.(UnspecifiedContentTypeResponseObject); ok {
		if err := validResponse.VisitUnspecifiedContentTypeResponse(ctx.Writer); err != nil {
			ctx.Error(err)
		}
	} else if response != nil {
		ctx.Error(fmt.Errorf("Unexpected response type: %T", response))
	}
}

// URLEncodedExample operation middleware
func (sh *strictHandler) URLEncodedExample(ctx *gin.Context) {
	var request URLEncodedExampleRequestObject

	if err := ctx.Request.ParseForm(); err != nil {
		ctx.Error(err)
		return
	}
	var body URLEncodedExampleFormdataRequestBody
	if err := runtime.BindForm(&body, ctx.Request.Form, nil, nil); err != nil {
		ctx.Error(err)
		return
	}
	request.Body = &body

	handler := func(ctx *gin.Context, request interface{}) (interface{}, error) {
		return sh.ssi.URLEncodedExample(ctx, request.(URLEncodedExampleRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "URLEncodedExample")
	}

	response, err := handler(ctx, request)

	if err != nil {
		ctx.Error(err)
		ctx.Status(http.StatusInternalServerError)
	} else if validResponse, ok := response.(URLEncodedExampleResponseObject); ok {
		if err := validResponse.VisitURLEncodedExampleResponse(ctx.Writer); err != nil {
			ctx.Error(err)
		}
	} else if response != nil {
		ctx.Error(fmt.Errorf("Unexpected response type: %T", response))
	}
}

// HeadersExample operation middleware
func (sh *strictHandler) HeadersExample(ctx *gin.Context, params HeadersExampleParams) {
	var request HeadersExampleRequestObject

	request.Params = params

	var body HeadersExampleJSONRequestBody
	if err := ctx.ShouldBind(&body); err != nil {
		ctx.Status(http.StatusBadRequest)
		ctx.Error(err)
		return
	}
	request.Body = &body

	handler := func(ctx *gin.Context, request interface{}) (interface{}, error) {
		return sh.ssi.HeadersExample(ctx, request.(HeadersExampleRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "HeadersExample")
	}

	response, err := handler(ctx, request)

	if err != nil {
		ctx.Error(err)
		ctx.Status(http.StatusInternalServerError)
	} else if validResponse, ok := response.(HeadersExampleResponseObject); ok {
		if err := validResponse.VisitHeadersExampleResponse(ctx.Writer); err != nil {
			ctx.Error(err)
		}
	} else if response != nil {
		ctx.Error(fmt.Errorf("Unexpected response type: %T", response))
	}
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/+xYX2/bNhD/KgS3p0KynDRPeluDotu6rYOTPg15oMWzzFYiuePJimHouw8UJf+pFS/O",
	"7BgY+mbRd787/u4Pj1zxzJTWaNDkeLriCM4a7aD9mAqJ8HcFjvyXBJehsqSM5il/J+Sk+6+JOELlxLSA",
	"Xt3LZ0YT6FZVWFuoTHjV5Ivz+ivusjmUwv/6EWHGU/5DsnElCf+6BB5FaQvgTdNE33jw6SOP+ByEBGy9",
	"DT+vdrFpaYGn3BEqnXMPEsSuB8WUJsgBvTUv2jnhBXo/0hW3aCwgqcDRQhQVDFvqVsz0C2QUdqD0zOxz",
	"eWs0CaUdk2o2AwRNrCOPeQzHXGWtQQLJpkvmLWTEHOACkEecFHnH+N32OuscdjziC0AXDF2NxqOxj5ex",
	"oIVVPOVv26WIW0HzdkPrAFkzFPdf7z79wZRjoiJTClKZKIolKwW6uSgKkExpMt7HKiM34q0pbCP/i+zU",
	"33dc+rRpM+idkctzZEybmFv5fD0ev1JiNhG/CcaGMNZOJVsV1sLMRFUMkP5Zf9Wm1gwQDXY7S8qqIGUF",
	"0nawdtn+vRd5DuVrvGRmsIylIHEm1k9l6aLEd81gsEju5qZ2bG5qRoZJEAWrFc1Zr/hNdSvNBHNK5wWw",
	"3qloMJIFdD33Jy0n3V7uPcbZaynaQXmM67qO2+BVWIDOjAT5MlhVihwSq/NddY8tiKd8uiSftvvd9URJ",
	"FHGCR0psIZQ+fHS8Ujv5zvTJCjuUK0J7JMo4N/FXWNYGZWwFihII0CUrb73xwDkMlPKfa0mWCc2mwLQo",
	"QTIxI0D2wbAO0u2V7KSz+8F8DCIbqPa8XX+kf624p6Q9g3nEvQGeBlZCXSv0QSesIDpA28O/5ud/CkDP",
	"Zpj04h1Tw22wb1Fr6hBmzrfEocgN8BcsTbYkLjMwHM64vdn3Nc4gH8mnz/17eHzWkX/C1vfatX0sYVVY",
	"fJqzTus5tL2wkz6DxYWSYJLS3hyJfDFSnYVMzRTIuNtFHHx7qiXcGp0h0O4I5O8T2hBbg/lrDs2BBQYi",
	"5gyrgZWVI2aFc0xR20UKFa5KEvaax+eNZ7fB0v2mnR6K6pszxfTNpSJ6M746XuXtmfNmZ5R5oh4nv70P",
	"MsfeF082Mx058Z3O7oXK2V9S4q0XleES/jkIbM70DNTCT0RaMgSqUINkCyX6R4C92uwANmEdmoWCG5tp",
	"qH/dOWYgig5iXfPo0AvQw//4eeKc72bnztMm4uGJKyRLhYWPKJFNkyQ8jY1cLfIccKRMIqzizUPzTwAA",
	"AP//O0NNuucUAAA=",
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
