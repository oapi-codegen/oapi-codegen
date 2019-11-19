// Package api provides primitives to interact the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen DO NOT EDIT.
package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

// Error defines model for Error.
type Error struct {
	Code    int32  `json:"code" xml:"code"`
	Message string `json:"message" xml:"message"`
}

// NewPet defines model for NewPet.
type NewPet struct {
	Name                 string  `json:"name" xml:"name"`
	Tag                  *string `json:"tag,omitempty" xml:"tag,omitempty"`
	AdditionalProperties map[string]struct {
		Description *string `json:"description,omitempty" xml:"description,omitempty"`
	} `json:"-"`
}

// Pet defines model for Pet.
type Pet struct {
	// Embedded struct due to allOf(#/components/schemas/NewPet)
	NewPet
	// Embedded fields due to inline allOf schema
	Id int64 `json:"id" xml:"id"`
}

// FindPetsParams defines parameters for FindPets.
type FindPetsParams struct {

	// tags to filter by
	Tags *[]string `json:"tags,omitempty" xml:"tags-list>tags,omitempty"`

	// maximum number of results to return
	Limit *int32 `json:"limit,omitempty" xml:"limit,omitempty"`
}

// addPetJSONBody defines parameters for AddPet.
type addPetJSONBody NewPet

// AddPetRequestBody defines body for AddPet for application/json ContentType.
type AddPetJSONRequestBody addPetJSONBody

// Getter for additional properties for NewPet. Returns the specified
// element and whether it was found
func (a NewPet) Get(fieldName string) (value struct {
	Description *string `json:"description,omitempty" xml:"description,omitempty"`
}, found bool) {
	if a.AdditionalProperties != nil {
		value, found = a.AdditionalProperties[fieldName]
	}
	return
}

// Setter for additional properties for NewPet
func (a *NewPet) Set(fieldName string, value struct {
	Description *string `json:"description,omitempty" xml:"description,omitempty"`
}) {
	if a.AdditionalProperties == nil {
		a.AdditionalProperties = make(map[string]struct {
			Description *string `json:"description,omitempty" xml:"description,omitempty"`
		})
	}
	a.AdditionalProperties[fieldName] = value
}

// Override default JSON handling for NewPet to handle AdditionalProperties
func (a *NewPet) UnmarshalJSON(b []byte) error {
	object := make(map[string]json.RawMessage)
	err := json.Unmarshal(b, &object)
	if err != nil {
		return err
	}

	if raw, found := object["name"]; found {
		err = json.Unmarshal(raw, &a.Name)
		if err != nil {
			return errors.Wrap(err, "error reading 'name'")
		}
		delete(object, "name")
	}

	if raw, found := object["tag"]; found {
		err = json.Unmarshal(raw, &a.Tag)
		if err != nil {
			return errors.Wrap(err, "error reading 'tag'")
		}
		delete(object, "tag")
	}

	if len(object) != 0 {
		a.AdditionalProperties = make(map[string]struct {
			Description *string `json:"description,omitempty" xml:"description,omitempty"`
		})
		for fieldName, fieldBuf := range object {
			var fieldVal struct {
				Description *string `json:"description,omitempty" xml:"description,omitempty"`
			}
			err := json.Unmarshal(fieldBuf, &fieldVal)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("error unmarshaling field %s", fieldName))
			}
			a.AdditionalProperties[fieldName] = fieldVal
		}
	}
	return nil
}

// Override default JSON handling for NewPet to handle AdditionalProperties
func (a NewPet) MarshalJSON() ([]byte, error) {
	var err error
	object := make(map[string]json.RawMessage)

	object["name"], err = json.Marshal(a.Name)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("error marshaling 'name'"))
	}

	if a.Tag != nil {
		object["tag"], err = json.Marshal(a.Tag)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("error marshaling 'tag'"))
		}
	}

	for fieldName, field := range a.AdditionalProperties {
		object[fieldName], err = json.Marshal(field)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("error marshaling '%s'", fieldName))
		}
	}
	return json.Marshal(object)
}

type ServerInterface interface {
	//  (GET /pets)
	FindPets(w http.ResponseWriter, r *http.Request)
	//  (POST /pets)
	AddPet(w http.ResponseWriter, r *http.Request)
	//  (DELETE /pets/{id})
	DeletePet(w http.ResponseWriter, r *http.Request)
	//  (GET /pets/{id})
	FindPetById(w http.ResponseWriter, r *http.Request)
}

// ParamsForFindPets operation parameters from context
func ParamsForFindPets(ctx context.Context) *FindPetsParams {
	return ctx.Value("FindPetsParams").(*FindPetsParams)
}

// FindPets operation middleware
func FindPetsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		ctx = context.WithValue(ctx, "FindPetsParams", &params)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AddPet operation middleware
func AddPetCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// DeletePet operation middleware
func DeletePetCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var err error

		// ------------- Path parameter "id" -------------
		var id int64

		err = runtime.BindStyledParameter("simple", false, "id", chi.URLParam(r, "id"), &id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter id: %s", err), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(r.Context(), "id", id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// FindPetById operation middleware
func FindPetByIdCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var err error

		// ------------- Path parameter "id" -------------
		var id int64

		err = runtime.BindStyledParameter("simple", false, "id", chi.URLParam(r, "id"), &id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter id: %s", err), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(r.Context(), "id", id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(FindPetsCtx)
		r.Get("/pets", si.FindPets)
	})
	r.Group(func(r chi.Router) {
		r.Use(AddPetCtx)
		r.Post("/pets", si.AddPet)
	})
	r.Group(func(r chi.Router) {
		r.Use(DeletePetCtx)
		r.Delete("/pets/{id}", si.DeletePet)
	})
	r.Group(func(r chi.Router) {
		r.Use(FindPetByIdCtx)
		r.Get("/pets/{id}", si.FindPetById)
	})

	return r
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
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
