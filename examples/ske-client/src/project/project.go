// Package project provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/do87/oapi-codegen version (devel) DO NOT EDIT.
package project

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	skeclient "github.com/do87/oapi-codegen/examples/ske-client"
	"github.com/do87/oapi-codegen/pkg/runtime"
)

const (
	BearerAuthScopes = "bearerAuth.Scopes"
)

// Defines values for ProjectState.
const (
	STATE_CREATED     ProjectState = "STATE_CREATED"
	STATE_CREATING    ProjectState = "STATE_CREATING"
	STATE_DELETING    ProjectState = "STATE_DELETING"
	STATE_FAILED      ProjectState = "STATE_FAILED"
	STATE_UNSPECIFIED ProjectState = "STATE_UNSPECIFIED"
)

// Defines values for RuntimeErrorCode.
const (
	SKE_API_SERVER_ERROR      RuntimeErrorCode = "SKE_API_SERVER_ERROR"
	SKE_CONFIGURATION_PROBLEM RuntimeErrorCode = "SKE_CONFIGURATION_PROBLEM"
	SKE_INFRA_ERROR           RuntimeErrorCode = "SKE_INFRA_ERROR"
	SKE_QUOTA_EXCEEDED        RuntimeErrorCode = "SKE_QUOTA_EXCEEDED"
	SKE_RATE_LIMITS           RuntimeErrorCode = "SKE_RATE_LIMITS"
	SKE_REMAINING_RESOURCES   RuntimeErrorCode = "SKE_REMAINING_RESOURCES"
	SKE_TMP_AUTH_ERROR        RuntimeErrorCode = "SKE_TMP_AUTH_ERROR"
	SKE_UNREADY_NODES         RuntimeErrorCode = "SKE_UNREADY_NODES"
	SKE_UNSPECIFIED           RuntimeErrorCode = "SKE_UNSPECIFIED"
)

// Project defines model for Project.
type Project struct {
	ProjectID *string       `json:"projectId,omitempty"`
	State     *ProjectState `json:"state,omitempty"`
}

// ProjectState defines model for ProjectState.
type ProjectState string

// RuntimeError defines model for RuntimeError.
type RuntimeError struct {
	// Code - Code:    "SKE_UNSPECIFIED"
	//   Message: "An error occurred. Please open a support ticket if this error persists."
	// - Code:    "SKE_TMP_AUTH_ERROR"
	//   Message: "Authentication failed. This is a temporary error. Please wait while the system recovers."
	// - Code:    "SKE_QUOTA_EXCEEDED"
	//   Message: "Your project's resource quotas are exhausted. Please make sure your quota is sufficient for the ordered cluster."
	// - Code:    "SKE_RATE_LIMITS"
	//   Message: "While provisioning your cluster, request rate limits where incurred. Please wait while the system recovers."
	// - Code:    "SKE_INFRA_ERROR"
	//   Message: "An error occurred with the underlying infrastructure. Please open a support ticket if this error persists."
	// - Code:    "SKE_REMAINING_RESOURCES"
	//   Message: "There are remaining Kubernetes resources in your cluster that prevent deletion. Please make sure to remove them."
	// - Code:    "SKE_CONFIGURATION_PROBLEM"
	//   Message: "A configuration error occurred. Please open a support ticket if this error persists."
	// - Code:    "SKE_UNREADY_NODES"
	//   Message: "Not all worker nodes are ready. Please open a support ticket if this error persists."
	// - Code:    "SKE_API_SERVER_ERROR"
	//   Message: "The Kubernetes API server is not reporting readiness. Please open a support ticket if this error persists."
	Code    *RuntimeErrorCode `json:"code,omitempty"`
	Details *string           `json:"details,omitempty"`
	Message *string           `json:"message,omitempty"`
}

// RuntimeErrorCode - Code:    "SKE_UNSPECIFIED"
//
//		Message: "An error occurred. Please open a support ticket if this error persists."
//	  - Code:    "SKE_TMP_AUTH_ERROR"
//	    Message: "Authentication failed. This is a temporary error. Please wait while the system recovers."
//	  - Code:    "SKE_QUOTA_EXCEEDED"
//	    Message: "Your project's resource quotas are exhausted. Please make sure your quota is sufficient for the ordered cluster."
//	  - Code:    "SKE_RATE_LIMITS"
//	    Message: "While provisioning your cluster, request rate limits where incurred. Please wait while the system recovers."
//	  - Code:    "SKE_INFRA_ERROR"
//	    Message: "An error occurred with the underlying infrastructure. Please open a support ticket if this error persists."
//	  - Code:    "SKE_REMAINING_RESOURCES"
//	    Message: "There are remaining Kubernetes resources in your cluster that prevent deletion. Please make sure to remove them."
//	  - Code:    "SKE_CONFIGURATION_PROBLEM"
//	    Message: "A configuration error occurred. Please open a support ticket if this error persists."
//	  - Code:    "SKE_UNREADY_NODES"
//	    Message: "Not all worker nodes are ready. Please open a support ticket if this error persists."
//	  - Code:    "SKE_API_SERVER_ERROR"
//	    Message: "The Kubernetes API server is not reporting readiness. Please open a support ticket if this error persists."
type RuntimeErrorCode string

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client skeclient.HttpRequestDoer
}

// Creates a new Client, with reasonable defaults
func NewClient(server string, httpClient skeclient.HttpRequestDoer) *Client {
	// create a client with sane default values
	client := Client{
		Server: server,
		Client: httpClient,
	}
	return &client
}

// The interface specification for the client above.
type ClientInterface interface {
	// DeleteProject request
	DeleteProject(ctx context.Context, projectID string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetProject request
	GetProject(ctx context.Context, projectID string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreateProject request
	CreateProject(ctx context.Context, projectID string, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) DeleteProject(ctx context.Context, projectID string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteProjectRequest(c.Server, projectID)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetProject(ctx context.Context, projectID string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetProjectRequest(c.Server, projectID)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateProject(ctx context.Context, projectID string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateProjectRequest(c.Server, projectID)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewDeleteProjectRequest generates requests for DeleteProject
func NewDeleteProjectRequest(server string, projectID string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "projectID", runtime.ParamLocationPath, projectID)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/v1/projects/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetProjectRequest generates requests for GetProject
func NewGetProjectRequest(server string, projectID string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "projectID", runtime.ParamLocationPath, projectID)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/v1/projects/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateProjectRequest generates requests for CreateProject
func NewCreateProjectRequest(server string, projectID string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "projectID", runtime.ParamLocationPath, projectID)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/v1/projects/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
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
func NewClientWithResponses(server string, httpClient skeclient.HttpRequestDoer) *ClientWithResponses {
	return &ClientWithResponses{NewClient(server, httpClient)}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// DeleteProject request
	DeleteProjectWithResponse(ctx context.Context, projectID string, reqEditors ...RequestEditorFn) (*DeleteProjectResponse, error)

	// GetProject request
	GetProjectWithResponse(ctx context.Context, projectID string, reqEditors ...RequestEditorFn) (*GetProjectResponse, error)

	// CreateProject request
	CreateProjectWithResponse(ctx context.Context, projectID string, reqEditors ...RequestEditorFn) (*CreateProjectResponse, error)
}

type DeleteProjectResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *map[string]interface{}
	JSON202      *map[string]interface{}
	JSON400      *map[string]interface{}
	JSONDefault  *RuntimeError
	HasError     error // Aggregated error
}

// Status returns HTTPResponse.Status
func (r DeleteProjectResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteProjectResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetProjectResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Project
	JSON404      *map[string]interface{}
	JSONDefault  *RuntimeError
	HasError     error // Aggregated error
}

// Status returns HTTPResponse.Status
func (r GetProjectResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetProjectResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateProjectResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Project
	JSON400      *map[string]interface{}
	JSONDefault  *RuntimeError
	HasError     error // Aggregated error
}

// Status returns HTTPResponse.Status
func (r CreateProjectResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateProjectResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// DeleteProjectWithResponse request returning *DeleteProjectResponse
func (c *ClientWithResponses) DeleteProjectWithResponse(ctx context.Context, projectID string, reqEditors ...RequestEditorFn) (*DeleteProjectResponse, error) {
	rsp, err := c.DeleteProject(ctx, projectID, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteProjectResponse(rsp)
}

// GetProjectWithResponse request returning *GetProjectResponse
func (c *ClientWithResponses) GetProjectWithResponse(ctx context.Context, projectID string, reqEditors ...RequestEditorFn) (*GetProjectResponse, error) {
	rsp, err := c.GetProject(ctx, projectID, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetProjectResponse(rsp)
}

// CreateProjectWithResponse request returning *CreateProjectResponse
func (c *ClientWithResponses) CreateProjectWithResponse(ctx context.Context, projectID string, reqEditors ...RequestEditorFn) (*CreateProjectResponse, error) {
	rsp, err := c.CreateProject(ctx, projectID, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateProjectResponse(rsp)
}

// ParseDeleteProjectResponse parses an HTTP response from a DeleteProjectWithResponse call
func ParseDeleteProjectResponse(rsp *http.Response) (*DeleteProjectResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteProjectResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 202:
		var dest map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON202 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest RuntimeError
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}

// ParseGetProjectResponse parses an HTTP response from a GetProjectWithResponse call
func ParseGetProjectResponse(rsp *http.Response) (*GetProjectResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetProjectResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Project
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 404:
		var dest map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON404 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest RuntimeError
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}

// ParseCreateProjectResponse parses an HTTP response from a CreateProjectWithResponse call
func ParseCreateProjectResponse(rsp *http.Response) (*CreateProjectResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreateProjectResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Project
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest RuntimeError
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}