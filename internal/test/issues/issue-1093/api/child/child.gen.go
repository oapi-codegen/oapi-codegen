// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.0.0-00010101000000-000000000000 DO NOT EDIT.
package api

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	externalRef0 "github.com/oapi-codegen/oapi-codegen/v2/internal/test/issues/issue-1093/api/parent"
	strictgin "github.com/oapi-codegen/runtime/strictmiddleware/gin"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (GET /pets)
	GetPets(c *gin.Context)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandler       func(*gin.Context, error, int)
}

type MiddlewareFunc func(c *gin.Context)

// GetPets operation middleware
func (siw *ServerInterfaceWrapper) GetPets(c *gin.Context) {

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.GetPets(c)
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

	router.GET(options.BaseURL+"/pets", wrapper.GetPets)
}

type GetPetsRequestObject struct {
}

type GetPetsResponseObject interface {
	VisitGetPetsResponse(w http.ResponseWriter) error
}

type GetPets200JSONResponse externalRef0.Pet

func (response GetPets200JSONResponse) VisitGetPetsResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {

	// (GET /pets)
	GetPets(ctx context.Context, request GetPetsRequestObject) (GetPetsResponseObject, error)
}

type StrictHandlerFunc = strictgin.StrictGinHandlerFunc
type StrictMiddlewareFunc = strictgin.StrictGinMiddlewareFunc

func NewStrictHandler(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares}
}

type strictHandler struct {
	ssi         StrictServerInterface
	middlewares []StrictMiddlewareFunc
}

// GetPets operation middleware
func (sh *strictHandler) GetPets(ctx *gin.Context) {
	var request GetPetsRequestObject

	handler := func(ctx *gin.Context, request interface{}) (interface{}, error) {
		return sh.ssi.GetPets(ctx, request.(GetPetsRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetPets")
	}

	response, err := handler(ctx, request)

	if err != nil {
		ctx.Error(err)
		ctx.Status(http.StatusInternalServerError)
	} else if validResponse, ok := response.(GetPetsResponseObject); ok {
		if err := validResponse.VisitGetPetsResponse(ctx.Writer); err != nil {
			ctx.Error(err)
		}
	} else if response != nil {
		ctx.Error(fmt.Errorf("unexpected response type: %T", response))
	}
}

// Base64 encoded, compressed with deflate, json marshaled Swagger object
const swaggerSpec = "" +
	"bFHBbtswDP0VgdvRiL3tph8YdguG3tKgUGU6VmBLLEkXCAL/e0G5QRqgJ5H24+Pje1eIZaaSMauAv4LE" +
	"EedQSwqMWV/2qLXjQsiasP7LYUZ79UIIHkQ55ROsDWg4ffN9bYDxbUmMPfjDNn1sbqjyesaosBos5aEY" +
	"QY8SOZGmksHD05jExTFNvRPC6NJMhVXcXHqcxA1cZqcjuk1yxUADmnQy/joIDbwjy8b3a9eZ2EKYAyXw" +
	"8GfX7TpogIKO9cCWcDPktJ3/qOc/6sJZHKHel8tFFK0MWvtFkN0YxIUYUcRpec5Ql3Iwnn89ePiLurdN" +
	"ZpBQybL5+7vr7IklK+YqIBBNKdbB9iym4haWVT8ZB/Dwo72n2X5G2X7JsVr8eEpwslR9wzK5mwYDGlSQ" +
	"zTTwhyssPIGHdjNzPa4fAQAA//8="

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	compressed, err := base64.StdEncoding.DecodeString(swaggerSpec)
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr := flate.NewReader(bytes.NewReader(compressed))
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(zr); err != nil {
		return nil, fmt.Errorf("read flate: %w", err)
	}
	if err := zr.Close(); err != nil {
		return nil, fmt.Errorf("close flate reader: %w", err)
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
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	for rawPath, rawFunc := range externalRef0.PathToRawSpec(path.Join(path.Dir(pathToFile), "parent.api.yaml")) {
		if _, ok := res[rawPath]; ok {
			// it is not possible to compare functions in golang, so always overwrite the old value
		}
		res[rawPath] = rawFunc
	}
	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
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
