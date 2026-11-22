package responsescomponentinlineunion

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// strictServer returns the component-response envelopes under test. The
// status to return is selected per request through the X-Status header so a
// single handler covers every case.
type strictServer struct{}

func (strictServer) ListThings(ctx context.Context, request ListThingsRequestObject) (ListThingsResponseObject, error) {
	ve := ValidationError{
		Error:   "validation",
		Message: "field is required",
		Fields:  []string{"name"},
	}

	switch statusFromContext(ctx) {
	case "400":
		// Issue #2539: this envelope embeds BadRequestJSONResponse, which is
		// the union type declared for components/responses/BadRequest.
		var body BadRequest
		if err := body.FromValidationError(ve); err != nil {
			return nil, err
		}
		return ListThings400JSONResponse{BadRequestJSONResponse: body}, nil
	case "422":
		// Same union shape behind a response that also declares headers, so
		// the envelope is a struct with a Body field of the component's type.
		var body ServiceError
		if err := body.FromValidationError(ve); err != nil {
			return nil, err
		}
		return ListThings422JSONResponse{ServiceErrorJSONResponse{
			Body:    body,
			Headers: ServiceErrorResponseHeaders{XRequestId: "req-1"},
		}}, nil
	case "409":
		// The intersection of #2539 and #2549: the component response is an
		// allOf-merged union, so the envelope has to name the component's
		// model type *and* serialize through its MarshalJSON, or traceId
		// (merged in by the allOf) never reaches the wire.
		body := Conflict{TraceId: "trace-1"}
		if err := body.FromValidationError(ve); err != nil {
			return nil, err
		}
		return ListThings409JSONResponse{ConflictJSONResponse: body}, nil
	case "404":
		return ListThings404JSONResponse{NotFoundJSONResponse{Error: "not_found", Message: "no such thing"}}, nil
	case "default":
		// A non-fixed status code makes the envelope a struct with a Body
		// field; the body type must still be the component's model type
		// rather than an anonymous struct nobody can populate.
		var body BadRequest
		if err := body.FromValidationError(ve); err != nil {
			return nil, err
		}
		return ListThingsdefaultJSONResponse{Body: body, StatusCode: http.StatusTeapot}, nil
	default:
		return ListThings200JSONResponse{{Id: "1"}}, nil
	}
}

type statusKey struct{}

func statusFromContext(ctx context.Context) string {
	s, _ := ctx.Value(statusKey{}).(string)
	return s
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	handler := Handler(NewStrictHandler(strictServer{}, nil))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), statusKey{}, r.Header.Get("X-Status"))
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func get(t *testing.T, srv *httptest.Server, status string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/things", nil)
	require.NoError(t, err)
	req.Header.Set("X-Status", status)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

const wantValidationError = `{"error":"validation","fields":["name"],"message":"field is required"}`

// The strict envelope for a $ref to a component response with an inline
// oneOf must serialize the union payload, not an empty object.
func TestComponentInlineUnionResponseIsSerialized(t *testing.T) {
	srv := newTestServer(t)

	resp := get(t, srv, "400")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var got json.RawMessage
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.JSONEq(t, wantValidationError, string(got))
}

// With response headers the envelope carries the union in a Body field; the
// headers and the payload must both reach the wire.
func TestComponentInlineUnionResponseWithHeadersIsSerialized(t *testing.T) {
	srv := newTestServer(t)

	resp := get(t, srv, "422")
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	assert.Equal(t, "req-1", resp.Header.Get("X-Request-Id"))

	var got json.RawMessage
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.JSONEq(t, wantValidationError, string(got))
}

// The generated client decodes the component response into the declared
// union type, so the union accessors are usable on the wrapper field.
func TestClientDecodesComponentInlineUnionResponse(t *testing.T) {
	srv := newTestServer(t)

	client, err := NewClientWithResponses(srv.URL, WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
		req.Header.Set("X-Status", "400")
		return nil
	}))
	require.NoError(t, err)

	resp, err := client.ListThingsWithResponse(context.Background())
	require.NoError(t, err)
	require.NotNil(t, resp.JSON400)

	ve, err := resp.JSON400.AsValidationError()
	require.NoError(t, err)
	assert.Equal(t, []string{"name"}, ve.Fields)
}

// Control: a component response whose schema is a plain $ref keeps working.
func TestComponentRefResponseIsSerialized(t *testing.T) {
	srv := newTestServer(t)

	resp := get(t, srv, "404")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var got Error
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, Error{Error: "not_found", Message: "no such thing"}, got)
}

// The intersection shape from issues #2539 and #2549: an allOf-merged union
// declared inline in a component response and reached via $ref. Both the
// union branch and the properties merged in by the allOf must be written.
func TestComponentInlineAllOfUnionResponseIsSerialized(t *testing.T) {
	srv := newTestServer(t)

	resp := get(t, srv, "409")
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var got json.RawMessage
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.JSONEq(t, `{"traceId":"trace-1","error":"validation","fields":["name"],"message":"field is required"}`, string(got))

	// ...and the same bytes round-trip back into the model.
	var decoded Conflict
	require.NoError(t, json.Unmarshal(got, &decoded))
	assert.Equal(t, "trace-1", decoded.TraceId)
	ve, err := decoded.AsValidationError()
	require.NoError(t, err)
	assert.Equal(t, []string{"name"}, ve.Fields)
}

// A response with a non-fixed status code ($ref'd component response under
// "default") wraps the body in a Body field. That field must be typed as the
// component's model, not as an anonymous struct with an unexported union.
func TestComponentInlineUnionDefaultResponseIsSerialized(t *testing.T) {
	srv := newTestServer(t)

	resp := get(t, srv, "default")
	assert.Equal(t, http.StatusTeapot, resp.StatusCode)

	var got json.RawMessage
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.JSONEq(t, wantValidationError, string(got))
}
