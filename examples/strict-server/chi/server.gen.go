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
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
)

// Badrequest defines model for badrequest.
type Badrequest string

// Example defines model for example.
type Example struct {
	Value *string `json:"value,omitempty"`
}

// JSONExampleJSONBody defines parameters for JSONExample.
type JSONExampleJSONBody Example

// MultipartExampleMultipartBody defines parameters for MultipartExample.
type MultipartExampleMultipartBody Example

// MultipleRequestAndResponseTypesJSONBody defines parameters for MultipleRequestAndResponseTypes.
type MultipleRequestAndResponseTypesJSONBody Example

// MultipleRequestAndResponseTypesFormdataBody defines parameters for MultipleRequestAndResponseTypes.
type MultipleRequestAndResponseTypesFormdataBody Example

// MultipleRequestAndResponseTypesMultipartBody defines parameters for MultipleRequestAndResponseTypes.
type MultipleRequestAndResponseTypesMultipartBody Example

// MultipleRequestAndResponseTypesTextBody defines parameters for MultipleRequestAndResponseTypes.
type MultipleRequestAndResponseTypesTextBody string

// TextExampleTextBody defines parameters for TextExample.
type TextExampleTextBody string

// URLEncodedExampleFormdataBody defines parameters for URLEncodedExample.
type URLEncodedExampleFormdataBody Example

// JSONExampleJSONRequestBody defines body for JSONExample for application/json ContentType.
type JSONExampleJSONRequestBody JSONExampleJSONBody

// MultipartExampleMultipartRequestBody defines body for MultipartExample for multipart/form-data ContentType.
type MultipartExampleMultipartRequestBody MultipartExampleMultipartBody

// MultipleRequestAndResponseTypesJSONRequestBody defines body for MultipleRequestAndResponseTypes for application/json ContentType.
type MultipleRequestAndResponseTypesJSONRequestBody MultipleRequestAndResponseTypesJSONBody

// MultipleRequestAndResponseTypesFormdataRequestBody defines body for MultipleRequestAndResponseTypes for application/x-www-form-urlencoded ContentType.
type MultipleRequestAndResponseTypesFormdataRequestBody MultipleRequestAndResponseTypesFormdataBody

// MultipleRequestAndResponseTypesMultipartRequestBody defines body for MultipleRequestAndResponseTypes for multipart/form-data ContentType.
type MultipleRequestAndResponseTypesMultipartRequestBody MultipleRequestAndResponseTypesMultipartBody

// MultipleRequestAndResponseTypesTextRequestBody defines body for MultipleRequestAndResponseTypes for text/plain ContentType.
type MultipleRequestAndResponseTypesTextRequestBody MultipleRequestAndResponseTypesTextBody

// TextExampleTextRequestBody defines body for TextExample for text/plain ContentType.
type TextExampleTextRequestBody TextExampleTextBody

// URLEncodedExampleFormdataRequestBody defines body for URLEncodedExample for application/x-www-form-urlencoded ContentType.
type URLEncodedExampleFormdataRequestBody URLEncodedExampleFormdataBody

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (POST /json)
	JSONExample(w http.ResponseWriter, r *http.Request)

	// (POST /multipart)
	MultipartExample(w http.ResponseWriter, r *http.Request)

	// (POST /multiple)
	MultipleRequestAndResponseTypes(w http.ResponseWriter, r *http.Request)

	// (POST /text)
	TextExample(w http.ResponseWriter, r *http.Request)

	// (POST /unknown)
	UnknownExample(w http.ResponseWriter, r *http.Request)

	// (POST /urlencoded)
	URLEncodedExample(w http.ResponseWriter, r *http.Request)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

// JSONExample operation middleware
func (siw *ServerInterfaceWrapper) JSONExample(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.JSONExample(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// MultipartExample operation middleware
func (siw *ServerInterfaceWrapper) MultipartExample(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.MultipartExample(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// MultipleRequestAndResponseTypes operation middleware
func (siw *ServerInterfaceWrapper) MultipleRequestAndResponseTypes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.MultipleRequestAndResponseTypes(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// TextExample operation middleware
func (siw *ServerInterfaceWrapper) TextExample(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.TextExample(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// UnknownExample operation middleware
func (siw *ServerInterfaceWrapper) UnknownExample(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.UnknownExample(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// URLEncodedExample operation middleware
func (siw *ServerInterfaceWrapper) URLEncodedExample(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.URLEncodedExample(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshalingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshalingParamError) Error() string {
	return fmt.Sprintf("Error unmarshaling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshalingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{})
}

type ChiServerOptions struct {
	BaseURL          string
	BaseRouter       chi.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r chi.Router) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r chi.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options ChiServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = chi.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/json", wrapper.JSONExample)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/multipart", wrapper.MultipartExample)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/multiple", wrapper.MultipleRequestAndResponseTypes)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/text", wrapper.TextExample)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/unknown", wrapper.UnknownExample)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/urlencoded", wrapper.URLEncodedExample)
	})

	return r
}

type JSONExampleRequestObject struct {
	Body *JSONExampleJSONRequestBody
}

type JSONExample200JSONResponse Example

func (t JSONExample200JSONResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((Example)(t))
}

type JSONExample400TextResponse Badrequest

func (t JSONExample400TextResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((Badrequest)(t))
}

type JSONExampledefaultResponse struct {
	StatusCode int
}

type MultipartExampleRequestObject struct {
	Body *multipart.Reader
}

type MultipartExample200MultipartformDataResponse struct {
	Body          io.Reader
	ContentLength int64
}

type MultipartExample400TextResponse Badrequest

func (t MultipartExample400TextResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((Badrequest)(t))
}

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

type MultipleRequestAndResponseTypes200TextResponse string

func (t MultipleRequestAndResponseTypes200TextResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((string)(t))
}

type MultipleRequestAndResponseTypes200JSONResponse Example

func (t MultipleRequestAndResponseTypes200JSONResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((Example)(t))
}

type MultipleRequestAndResponseTypes200FormdataResponse Example

func (t MultipleRequestAndResponseTypes200FormdataResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((Example)(t))
}

type MultipleRequestAndResponseTypes200ImagepngResponse struct {
	Body          io.Reader
	ContentLength int64
}

type MultipleRequestAndResponseTypes200MultipartformDataResponse struct {
	Body          io.Reader
	ContentLength int64
}

type TextExampleRequestObject struct {
	Body *TextExampleTextRequestBody
}

type TextExample200TextResponse string

func (t TextExample200TextResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((string)(t))
}

type TextExample400TextResponse Badrequest

func (t TextExample400TextResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((Badrequest)(t))
}

type TextExampledefaultResponse struct {
	StatusCode int
}

type UnknownExampleRequestObject struct {
	Body io.Reader
}

type UnknownExample200Videomp4Response struct {
	Body          io.Reader
	ContentLength int64
}

type UnknownExample400TextResponse Badrequest

func (t UnknownExample400TextResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((Badrequest)(t))
}

type UnknownExampledefaultResponse struct {
	StatusCode int
}

type URLEncodedExampleRequestObject struct {
	Body *URLEncodedExampleFormdataRequestBody
}

type URLEncodedExampledefaultResponse struct {
	StatusCode int
}

type URLEncodedExample200FormdataResponse Example

func (t URLEncodedExample200FormdataResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((Example)(t))
}

type URLEncodedExample400TextResponse Badrequest

func (t URLEncodedExample400TextResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((Badrequest)(t))
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {

	// (POST /json)
	JSONExample(ctx context.Context, request JSONExampleRequestObject) interface{}

	// (POST /multipart)
	MultipartExample(ctx context.Context, request MultipartExampleRequestObject) interface{}

	// (POST /multiple)
	MultipleRequestAndResponseTypes(ctx context.Context, request MultipleRequestAndResponseTypesRequestObject) interface{}

	// (POST /text)
	TextExample(ctx context.Context, request TextExampleRequestObject) interface{}

	// (POST /unknown)
	UnknownExample(ctx context.Context, request UnknownExampleRequestObject) interface{}

	// (POST /urlencoded)
	URLEncodedExample(ctx context.Context, request URLEncodedExampleRequestObject) interface{}
}

type StrictHandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request, args interface{}) interface{}

type StrictMiddlewareFunc func(f StrictHandlerFunc, operationID string) StrictHandlerFunc

func NewStrictHandler(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares}
}

type strictHandler struct {
	ssi         StrictServerInterface
	middlewares []StrictMiddlewareFunc
}

// JSONExample operation middleware
func (sh *strictHandler) JSONExample(w http.ResponseWriter, r *http.Request) {
	var request JSONExampleRequestObject

	var body JSONExampleJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "can't decode JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}
	request.Body = &body

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) interface{} {
		return sh.ssi.JSONExample(ctx, request.(JSONExampleRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "JSONExample")
	}

	response := handler(r.Context(), w, r, request)

	switch v := response.(type) {
	case JSONExample200JSONResponse:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		writeJSON(w, v)
	case JSONExample400TextResponse:
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(400)
		writeRaw(w, ([]byte)(v))
	case JSONExampledefaultResponse:
		w.WriteHeader(v.StatusCode)
	case error:
		http.Error(w, v.Error(), http.StatusInternalServerError)
	case nil:
	default:
		http.Error(w, fmt.Sprintf("Unexpected response type: %T", v), http.StatusInternalServerError)
	}
}

// MultipartExample operation middleware
func (sh *strictHandler) MultipartExample(w http.ResponseWriter, r *http.Request) {
	var request MultipartExampleRequestObject

	if reader, err := r.MultipartReader(); err != nil {
		http.Error(w, "can't decode multipart body: "+err.Error(), http.StatusBadRequest)
		return
	} else {
		request.Body = reader
	}

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) interface{} {
		return sh.ssi.MultipartExample(ctx, request.(MultipartExampleRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "MultipartExample")
	}

	response := handler(r.Context(), w, r, request)

	switch v := response.(type) {
	case MultipartExample200MultipartformDataResponse:
		w.Header().Set("Content-Type", "multipart/form-data")
		if v.ContentLength != 0 {
			w.Header().Set("Content-Length", fmt.Sprint(v.ContentLength))
		}
		w.WriteHeader(200)
		if closer, ok := v.Body.(io.ReadCloser); ok {
			defer closer.Close()
		}
		_, _ = io.Copy(w, v.Body)
	case MultipartExample400TextResponse:
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(400)
		writeRaw(w, ([]byte)(v))
	case MultipartExampledefaultResponse:
		w.WriteHeader(v.StatusCode)
	case error:
		http.Error(w, v.Error(), http.StatusInternalServerError)
	case nil:
	default:
		http.Error(w, fmt.Sprintf("Unexpected response type: %T", v), http.StatusInternalServerError)
	}
}

// MultipleRequestAndResponseTypes operation middleware
func (sh *strictHandler) MultipleRequestAndResponseTypes(w http.ResponseWriter, r *http.Request) {
	var request MultipleRequestAndResponseTypesRequestObject

	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		var body MultipleRequestAndResponseTypesJSONRequestBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "can't decode JSON body: "+err.Error(), http.StatusBadRequest)
			return
		}
		request.JSONBody = &body
	}
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "can't decode formdata: "+err.Error(), http.StatusBadRequest)
			return
		}
		var body MultipleRequestAndResponseTypesFormdataRequestBody
		if err := runtime.BindForm(&body, r.Form, nil, nil); err != nil {
			http.Error(w, "can't bind formdata: "+err.Error(), http.StatusBadRequest)
			return
		}
		request.FormdataBody = &body
	}
	if strings.HasPrefix(r.Header.Get("Content-Type"), "image/png") {
		request.Body = r.Body
	}
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if reader, err := r.MultipartReader(); err != nil {
			http.Error(w, "can't decode multipart body: "+err.Error(), http.StatusBadRequest)
			return
		} else {
			request.MultipartBody = reader
		}
	}
	if strings.HasPrefix(r.Header.Get("Content-Type"), "text/plain") {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "can't read body: "+err.Error(), http.StatusBadRequest)
			return
		}
		body := MultipleRequestAndResponseTypesTextRequestBody(data)
		request.TextBody = &body
	}

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) interface{} {
		return sh.ssi.MultipleRequestAndResponseTypes(ctx, request.(MultipleRequestAndResponseTypesRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "MultipleRequestAndResponseTypes")
	}

	response := handler(r.Context(), w, r, request)

	switch v := response.(type) {
	case MultipleRequestAndResponseTypes200TextResponse:
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)
		writeRaw(w, ([]byte)(v))
	case MultipleRequestAndResponseTypes200JSONResponse:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		writeJSON(w, v)
	case MultipleRequestAndResponseTypes200FormdataResponse:
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		w.WriteHeader(200)
		if form, err := runtime.MarshalForm(v, nil); err != nil {
			fmt.Fprintln(w, err)
		} else {
			writeRaw(w, []byte(form.Encode()))
		}
	case MultipleRequestAndResponseTypes200ImagepngResponse:
		w.Header().Set("Content-Type", "image/png")
		if v.ContentLength != 0 {
			w.Header().Set("Content-Length", fmt.Sprint(v.ContentLength))
		}
		w.WriteHeader(200)
		if closer, ok := v.Body.(io.ReadCloser); ok {
			defer closer.Close()
		}
		_, _ = io.Copy(w, v.Body)
	case MultipleRequestAndResponseTypes200MultipartformDataResponse:
		w.Header().Set("Content-Type", "multipart/form-data")
		if v.ContentLength != 0 {
			w.Header().Set("Content-Length", fmt.Sprint(v.ContentLength))
		}
		w.WriteHeader(200)
		if closer, ok := v.Body.(io.ReadCloser); ok {
			defer closer.Close()
		}
		_, _ = io.Copy(w, v.Body)
	case error:
		http.Error(w, v.Error(), http.StatusInternalServerError)
	case nil:
	default:
		http.Error(w, fmt.Sprintf("Unexpected response type: %T", v), http.StatusInternalServerError)
	}
}

// TextExample operation middleware
func (sh *strictHandler) TextExample(w http.ResponseWriter, r *http.Request) {
	var request TextExampleRequestObject

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	body := TextExampleTextRequestBody(data)
	request.Body = &body

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) interface{} {
		return sh.ssi.TextExample(ctx, request.(TextExampleRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "TextExample")
	}

	response := handler(r.Context(), w, r, request)

	switch v := response.(type) {
	case TextExample200TextResponse:
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)
		writeRaw(w, ([]byte)(v))
	case TextExample400TextResponse:
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(400)
		writeRaw(w, ([]byte)(v))
	case TextExampledefaultResponse:
		w.WriteHeader(v.StatusCode)
	case error:
		http.Error(w, v.Error(), http.StatusInternalServerError)
	case nil:
	default:
		http.Error(w, fmt.Sprintf("Unexpected response type: %T", v), http.StatusInternalServerError)
	}
}

// UnknownExample operation middleware
func (sh *strictHandler) UnknownExample(w http.ResponseWriter, r *http.Request) {
	var request UnknownExampleRequestObject

	request.Body = r.Body

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) interface{} {
		return sh.ssi.UnknownExample(ctx, request.(UnknownExampleRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "UnknownExample")
	}

	response := handler(r.Context(), w, r, request)

	switch v := response.(type) {
	case UnknownExample200Videomp4Response:
		w.Header().Set("Content-Type", "video/mp4")
		if v.ContentLength != 0 {
			w.Header().Set("Content-Length", fmt.Sprint(v.ContentLength))
		}
		w.WriteHeader(200)
		if closer, ok := v.Body.(io.ReadCloser); ok {
			defer closer.Close()
		}
		_, _ = io.Copy(w, v.Body)
	case UnknownExample400TextResponse:
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(400)
		writeRaw(w, ([]byte)(v))
	case UnknownExampledefaultResponse:
		w.WriteHeader(v.StatusCode)
	case error:
		http.Error(w, v.Error(), http.StatusInternalServerError)
	case nil:
	default:
		http.Error(w, fmt.Sprintf("Unexpected response type: %T", v), http.StatusInternalServerError)
	}
}

// URLEncodedExample operation middleware
func (sh *strictHandler) URLEncodedExample(w http.ResponseWriter, r *http.Request) {
	var request URLEncodedExampleRequestObject

	if err := r.ParseForm(); err != nil {
		http.Error(w, "can't decode formdata: "+err.Error(), http.StatusBadRequest)
		return
	}
	var body URLEncodedExampleFormdataRequestBody
	if err := runtime.BindForm(&body, r.Form, nil, nil); err != nil {
		http.Error(w, "can't bind formdata: "+err.Error(), http.StatusBadRequest)
		return
	}
	request.Body = &body

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) interface{} {
		return sh.ssi.URLEncodedExample(ctx, request.(URLEncodedExampleRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "URLEncodedExample")
	}

	response := handler(r.Context(), w, r, request)

	switch v := response.(type) {
	case URLEncodedExampledefaultResponse:
		w.WriteHeader(v.StatusCode)
	case URLEncodedExample200FormdataResponse:
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		w.WriteHeader(200)
		if form, err := runtime.MarshalForm(v, nil); err != nil {
			fmt.Fprintln(w, err)
		} else {
			writeRaw(w, []byte(form.Encode()))
		}
	case URLEncodedExample400TextResponse:
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(400)
		writeRaw(w, ([]byte)(v))
	case error:
		http.Error(w, v.Error(), http.StatusInternalServerError)
	case nil:
	default:
		http.Error(w, fmt.Sprintf("Unexpected response type: %T", v), http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		fmt.Fprintln(w, err)
	}
}

func writeRaw(w http.ResponseWriter, b []byte) {
	if _, err := w.Write(b); err != nil {
		fmt.Fprintln(w, err)
	}
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/+xWy27bOhD9FWLuXSqW22alXVNk0WeAJF0VXdDi2GJKcVhyZNkI9O8FJdmNXBlwUife",
	"dCc+5szhmYfmHnIqHVm0HCC7B4/BkQ3YLmZSefxZYeC4ysky2vaTccWpM1LbuAp5gaWMX/97nEMG/6W/",
	"QdPuNKQPwJqmSUBhyL12rMlCBhdSXW9Pkx5yhASvHUIGgb22C2gSwJUsncF45jw59Kw78ktpKhwxaZLN",
	"Ds3uMO/ZaDuneHnI6h1ZltoGofR8jh4ti14FETGCCJVz5BmVmK1F9JCzCOiX6CEB1hyJwc3DfdETDpDA",
	"En3oHL2aTCfT+BxyaKXTkMGbdisBJ7loH5TeBWr1dtRpMeT64ebqi9BByIqplKxzacxalNKHQhqDSmjL",
	"FDlWOYcJtK68jMbvVW9+2WuZQK/4Ban1Tuilc0bnrd2W0GEJsIlU0wo+SLTX0+lzuNlNsquPUeLzztkY",
	"xpbUIFsjzFxWZkT0r/aHpdoK9J58/7K0rAxrJz0/DNZQ7c+bK4dIvsVL5+TLMyVZPpPqx/J0UuH7ZjBa",
	"JDcF1UEUVAsmoVAaUWsuxMZwp7q1FVIEbRcGxYZUMhpJg333emvVdf+W24jx7LWUDFBWZ3Vdn7XBq7xB",
	"m5NC9TRYXcoFps4uhuYRWzJkMFtzTNs/u+uRkijZ+5fZdflC7eSf0uOF3dVehNjf725xdVCrO2LI/+pN",
	"L9Csqm5zv2a91SGyPTGDDlBxqRVSWrrzRyKfStRBKe7R9frTZXfnsfPO0Wr+kR3reH5PEZY4zrejb4Ds",
	"2z1U3kAGBbPL0rQbmSehlosF+ommNA6/zffmVwAAAP//pen/+5gMAAA=",
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
