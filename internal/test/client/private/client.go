package pkg

import (
	"context"
	"errors"
	"fmt"
)

// Client ...
type Client struct {
	gen generatedClientInterface
}

var _ ClientInterface = &Client{}

// ClientInterface ...
type ClientInterface interface {
	PostJSON(ctx context.Context, body PostJsonJSONRequestBody) (*SchemaObject, error)
}

// NewClient ...
func NewClient(server string, opts ...clientOption) (*Client, error) {
	genClient, err := newGeneratedClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{gen: genClient}, nil
}

// manually implement golang's `Error() string` interface so it can be passed as an `error`
func (e *Error) Error() string {
	return fmt.Sprintf("Error httpStatusCode: '%v', errorCode: '%s', message: '%s'", e.HttpStatusCode, e.ErrorCode, e.Message)
}

// PostJSON ...
func (c *Client) PostJSON(ctx context.Context, body PostJsonJSONRequestBody) (*SchemaObject, error) {
	resp, err := c.gen.PostJsonWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}

	switch {
	case resp.JSON200 != nil:
		return resp.JSON200, nil
	case resp.JSONDefault != nil:
		return nil, resp.JSONDefault
	default:
		return nil, errors.New("blah")
	}
}
