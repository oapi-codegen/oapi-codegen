// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen DO NOT EDIT.
package api

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
)

// Error defines model for Error.
type Error struct {

	// Error code
	Code int32 `json:"code"`

	// Error message
	Message string `json:"message"`
}

// NewPet defines model for NewPet.
type NewPet struct {

	// Name of the pet
	Name string `json:"name"`

	// Type of the pet
	Tag *string `json:"tag,omitempty"`
}

// Pet defines model for Pet.
type Pet struct {
	// Embedded struct due to allOf(#/components/schemas/NewPet)
	NewPet `yaml:",inline"`
	// Embedded fields due to inline allOf schema

	// Unique id of the pet
	Id int64 `json:"id"`
}

// FindPetsParams defines parameters for FindPets.
type FindPetsParams struct {

	// tags to filter by
	Tags *[]string `json:"tags,omitempty"`

	// maximum number of results to return
	Limit *int32 `json:"limit,omitempty"`
}

// AddPetJSONBody defines parameters for AddPet.
type AddPetJSONBody NewPet

// AddPetJSONRequestBody defines body for AddPet for application/json ContentType.
type AddPetJSONRequestBody AddPetJSONBody

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Returns all pets
	// (GET /pets)
	FindPets(w http.ResponseWriter, r *http.Request, params FindPetsParams)
	// Creates a new pet
	// (POST /pets)
	AddPet(w http.ResponseWriter, r *http.Request)
	// Deletes a pet by ID
	// (DELETE /pets/{id})
	DeletePet(w http.ResponseWriter, r *http.Request, id int64)
	// Returns a pet by ID
	// (GET /pets/{id})
	FindPetById(w http.ResponseWriter, r *http.Request, id int64)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
}

type MiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

// FindPets operation middleware
func (siw *ServerInterfaceWrapper) FindPets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params FindPetsParams

	// ------------- Optional query parameter "tags" -------------
	if paramValue := r.URL.Query().Get("tags"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "tags", r.URL.Query(), &params.Tags)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter tags: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "limit" -------------
	if paramValue := r.URL.Query().Get("limit"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "limit", r.URL.Query(), &params.Limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter limit: %s", err), http.StatusBadRequest)
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.FindPets(w, r, params)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// AddPet operation middleware
func (siw *ServerInterfaceWrapper) AddPet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.AddPet(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// DeletePet operation middleware
func (siw *ServerInterfaceWrapper) DeletePet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "id" -------------
	var id int64

	err = runtime.BindStyledParameter("simple", false, "id", chi.URLParam(r, "id"), &id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter id: %s", err), http.StatusBadRequest)
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.DeletePet(w, r, id)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// FindPetById operation middleware
func (siw *ServerInterfaceWrapper) FindPetById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "id" -------------
	var id int64

	err = runtime.BindStyledParameter("simple", false, "id", chi.URLParam(r, "id"), &id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter id: %s", err), http.StatusBadRequest)
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.FindPetById(w, r, id)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{})
}

type ChiServerOptions struct {
	BaseURL     string
	BaseRouter  chi.Router
	Middlewares []MiddlewareFunc
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
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
	}

	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/pets", wrapper.FindPets)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/pets", wrapper.AddPet)
	})
	r.Group(func(r chi.Router) {
		r.Delete(options.BaseURL+"/pets/{id}", wrapper.DeletePet)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/pets/{id}", wrapper.FindPetById)
	})

	return r
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
	"u/vdPwIAAP//v4qmX1cSAAA=",
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file.
func GetSwagger() (*openapi3.Swagger, error) {
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

	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error loading Swagger: %s", err)
	}
	return swagger, nil
}
