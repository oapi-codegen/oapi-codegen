// Package pkg2 provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen/v2 version v2.0.0-00010101000000-000000000000 DO NOT EDIT.
package pkg2

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	strictgin "github.com/oapi-codegen/runtime/strictmiddleware/gin"
)

// TestSchema defines model for TestSchema.
type TestSchema struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"field2"`
}

// TestRespExtFixedJSON defines model for testRespExtFixedJSON.
type TestRespExtFixedJSON = TestSchema

// TestRespExtFixedSpecialJSON defines model for testRespExtFixedSpecialJSON.
type TestRespExtFixedSpecialJSON = TestSchema

// TestRespExtHeaderFixedJSON defines model for testRespExtHeaderFixedJSON.
type TestRespExtHeaderFixedJSON = TestSchema

// TestRespExtHeaderFixedSpecialJSON defines model for testRespExtHeaderFixedSpecialJSON.
type TestRespExtHeaderFixedSpecialJSON = TestSchema

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandler       func(*gin.Context, error, int)
}

type MiddlewareFunc func(c *gin.Context)

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

}

type TestRespExtFixedJSONJSONResponse TestSchema

type TestRespExtFixedMultipartMultipartResponse func(writer *multipart.Writer) error

type TestRespExtFixedMultipartRelatedMultipartResponse func(writer *multipart.Writer) error

type TestRespExtFixedNoContentResponse struct {
}

type TestRespExtFixedOtherApplicationtestResponse struct {
	Body io.Reader

	ContentLength int64
}

type TestRespExtFixedSpecialJSONApplicationTestPlusJSONResponse TestSchema

type TestRespExtFixedWildcardApplicationResponse struct {
	Body io.Reader

	ContentType   string
	ContentLength int64
}

type TestRespExtHeaderFixedJSONResponseHeaders struct {
	Header1 string
	Header2 int
}
type TestRespExtHeaderFixedJSONJSONResponse struct {
	Body TestSchema

	Headers TestRespExtHeaderFixedJSONResponseHeaders
}

type TestRespExtHeaderFixedMultipartResponseHeaders struct {
	Header1 string
	Header2 int
}
type TestRespExtHeaderFixedMultipartMultipartResponse struct {
	Body func(writer *multipart.Writer) error

	Headers TestRespExtHeaderFixedMultipartResponseHeaders
}

type TestRespExtHeaderFixedMultipartRelatedResponseHeaders struct {
	Header1 string
	Header2 int
}
type TestRespExtHeaderFixedMultipartRelatedMultipartResponse struct {
	Body func(writer *multipart.Writer) error

	Headers TestRespExtHeaderFixedMultipartRelatedResponseHeaders
}

type TestRespExtHeaderFixedNoContentResponseHeaders struct {
	Header1 string
	Header2 int
}
type TestRespExtHeaderFixedNoContentResponse struct {
	Headers TestRespExtHeaderFixedNoContentResponseHeaders
}

type TestRespExtHeaderFixedOtherResponseHeaders struct {
	Header1 string
	Header2 int
}
type TestRespExtHeaderFixedOtherApplicationtestResponse struct {
	Body io.Reader

	Headers       TestRespExtHeaderFixedOtherResponseHeaders
	ContentLength int64
}

type TestRespExtHeaderFixedSpecialJSONResponseHeaders struct {
	Header1 string
	Header2 int
}
type TestRespExtHeaderFixedSpecialJSONApplicationTestPlusJSONResponse struct {
	Body TestSchema

	Headers TestRespExtHeaderFixedSpecialJSONResponseHeaders
}

type TestRespExtHeaderFixedWildcardResponseHeaders struct {
	Header1 string
	Header2 int
}
type TestRespExtHeaderFixedWildcardApplicationResponse struct {
	Body io.Reader

	Headers       TestRespExtHeaderFixedWildcardResponseHeaders
	ContentType   string
	ContentLength int64
}

type TestRespExtHeaderMultipartResponseHeaders struct {
	Header1 string
	Header2 int
}
type TestRespExtHeaderMultipartMultipartResponse struct {
	Body func(writer *multipart.Writer) error

	Headers TestRespExtHeaderMultipartResponseHeaders
}

type TestRespExtHeaderMultipartRelatedResponseHeaders struct {
	Header1 string
	Header2 int
}
type TestRespExtHeaderMultipartRelatedMultipartResponse struct {
	Body func(writer *multipart.Writer) error

	Headers TestRespExtHeaderMultipartRelatedResponseHeaders
}

type TestRespExtHeaderNoContentResponseHeaders struct {
	Header1 string
	Header2 int
}
type TestRespExtHeaderNoContentResponse struct {
	Headers TestRespExtHeaderNoContentResponseHeaders
}

type TestRespExtHeaderOtherResponseHeaders struct {
	Header1 string
	Header2 int
}
type TestRespExtHeaderOtherApplicationtestResponse struct {
	Body io.Reader

	Headers       TestRespExtHeaderOtherResponseHeaders
	ContentLength int64
}

type TestRespExtHeaderWildcardResponseHeaders struct {
	Header1 string
	Header2 int
}
type TestRespExtHeaderWildcardApplicationResponse struct {
	Body io.Reader

	Headers       TestRespExtHeaderWildcardResponseHeaders
	ContentType   string
	ContentLength int64
}

type TestRespExtMultipartMultipartResponse func(writer *multipart.Writer) error

type TestRespExtMultipartRelatedMultipartResponse func(writer *multipart.Writer) error

type TestRespExtNoContentResponse struct {
}

type TestRespExtOtherApplicationtestResponse struct {
	Body io.Reader

	ContentLength int64
}

type TestRespExtWildcardApplicationResponse struct {
	Body io.Reader

	ContentType   string
	ContentLength int64
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {
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
