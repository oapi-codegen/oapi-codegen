package bodiesformencoding

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// server is a strict server which records the request bodies it was handed,
// as the generated code bound them.
type server struct {
	subscribe   *SubscribeFormdataRequestBody
	unsubscribe *UnsubscribeFormdataRequestBody
	plain       *PlainFormdataRequestBody
	list        *SubscribeToListRequestObject
}

func (s *server) SubscribeToList(_ context.Context, request SubscribeToListRequestObject) (SubscribeToListResponseObject, error) {
	s.list = &request
	return SubscribeToList204Response{}, nil
}

func (s *server) Subscribe(_ context.Context, request SubscribeRequestObject) (SubscribeResponseObject, error) {
	s.subscribe = request.Body
	return Subscribe204Response{}, nil
}

func (s *server) Unsubscribe(_ context.Context, request UnsubscribeRequestObject) (UnsubscribeResponseObject, error) {
	s.unsubscribe = request.Body
	return Unsubscribe204Response{}, nil
}

func (s *server) Plain(_ context.Context, request PlainRequestObject) (PlainResponseObject, error) {
	s.plain = request.Body
	return Plain204Response{}, nil
}

// roundTrip serves one request from a generated client with a generated
// strict server, and returns what the server bound along with the body which
// went over the wire.
func roundTrip(t *testing.T, call func(c *ClientWithResponses) (int, error)) (*server, url.Values) {
	t.Helper()

	var wire url.Values
	bound := &server{}
	handler := Handler(NewStrictHandler(bound, nil))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		wire, err = url.ParseQuery(string(raw))
		require.NoError(t, err)
		r.Body = io.NopCloser(bytes.NewReader(raw))
		handler.ServeHTTP(w, r)
	}))
	defer ts.Close()

	client, err := NewClientWithResponses(ts.URL)
	require.NoError(t, err)
	status, err := call(client)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, status)
	return bound, wire
}

// Issue #2604: an array of primitives in the exploded form style is sent as
// repeated keys, whether the encoding spells that out (topic), declares only
// the style (ids) or only explode (scores).
func TestExplodedFormArraysRepeatTheKey(t *testing.T) {
	body := SubscribeFormdataRequestBody{
		Email:  "jane@example.com",
		Topic:  []string{"news", "releases, and more"},
		Ids:    &[]int{1, 2},
		Scores: &[]float32{0.5, 2},
	}
	bound, wire := roundTrip(t, func(c *ClientWithResponses) (int, error) {
		rsp, err := c.SubscribeWithFormdataBodyWithResponse(context.Background(), body)
		if err != nil {
			return 0, err
		}
		return rsp.StatusCode(), nil
	})

	assert.Equal(t, url.Values{
		"email":  {"jane@example.com"},
		"topic":  {"news", "releases, and more"},
		"ids":    {"1", "2"},
		"scores": {"0.5", "2"},
	}, wire)

	// The generated server reads repeated keys back as the same array.
	require.NotNil(t, bound.subscribe)
	assert.Equal(t, body, *bound.subscribe)
}

// The same, for a body which is a reference, and an array which comes from
// one of its allOf members.
func TestExplodedFormArrayInReferencedBody(t *testing.T) {
	body := UnsubscribeFormdataRequestBody{
		Email: "jane@example.com",
		Topic: &[]string{"news", "releases"},
	}
	bound, wire := roundTrip(t, func(c *ClientWithResponses) (int, error) {
		rsp, err := c.UnsubscribeWithFormdataBodyWithResponse(context.Background(), body)
		if err != nil {
			return 0, err
		}
		return rsp.StatusCode(), nil
	})

	assert.Equal(t, url.Values{
		"email": {"jane@example.com"},
		"topic": {"news", "releases"},
	}, wire)
	require.NotNil(t, bound.unsubscribe)
	assert.Equal(t, body, *bound.unsubscribe)
}

// A path parameter named like something the request builder could declare for
// the encoding must not break the generated code. That this package compiles
// is most of the test.
func TestExplodedFormArrayWithPathParameterNamedExplode(t *testing.T) {
	body := SubscribeToListFormdataRequestBody{Topic: &[]string{"news", "releases"}}
	bound, wire := roundTrip(t, func(c *ClientWithResponses) (int, error) {
		rsp, err := c.SubscribeToListWithFormdataBodyWithResponse(context.Background(), "weekly", body)
		if err != nil {
			return 0, err
		}
		return rsp.StatusCode(), nil
	})

	assert.Equal(t, url.Values{"topic": {"news", "releases"}}, wire)
	require.NotNil(t, bound.list)
	assert.Equal(t, "weekly", bound.list.Explode)
	require.NotNil(t, bound.list.Body)
	assert.Equal(t, body, *bound.list.Body)
}

// Every other encoding is ignored: those properties are sent with indexed and
// bracketed keys, as they were before the client looked at encodings at all,
// and the generated server still reads them back.
func TestOtherEncodingsAreUnchanged(t *testing.T) {
	role, age := "admin", 3
	body := SubscribeFormdataRequestBody{
		Email:      "jane@example.com",
		Topic:      []string{"news"},
		Unencoded:  &[]string{"a", "b"},
		Unexploded: &[]string{"a", "b"},
		Piped:      &[]string{"a", "b"},
		Deep:       &[]string{"a", "b"},
		Json:       &[]string{"a", "b"},
		Objects:    &[]Filter{{Role: &role, Age: &age}},
		Filter:     &Filter{Role: &role, Age: &age},
	}
	bound, wire := roundTrip(t, func(c *ClientWithResponses) (int, error) {
		rsp, err := c.SubscribeWithFormdataBodyWithResponse(context.Background(), body)
		if err != nil {
			return 0, err
		}
		return rsp.StatusCode(), nil
	})

	assert.Equal(t, url.Values{
		"email":            {"jane@example.com"},
		"topic":            {"news"},
		"unencoded[0]":     {"a"},
		"unencoded[1]":     {"b"},
		"unexploded[0]":    {"a"},
		"unexploded[1]":    {"b"},
		"piped[0]":         {"a"},
		"piped[1]":         {"b"},
		"deep[0]":          {"a"},
		"deep[1]":          {"b"},
		"json[0]":          {"a"},
		"json[1]":          {"b"},
		"objects[0][role]": {"admin"},
		"objects[0][age]":  {"3"},
		"filter[role]":     {"admin"},
		"filter[age]":      {"3"},
	}, wire)
	require.NotNil(t, bound.subscribe)
	assert.Equal(t, body, *bound.subscribe)
}

// A body with no encoding is untouched.
func TestNoEncodingIsUnchanged(t *testing.T) {
	body := PlainFormdataRequestBody{Topic: &[]string{"news", "releases"}}
	bound, wire := roundTrip(t, func(c *ClientWithResponses) (int, error) {
		rsp, err := c.PlainWithFormdataBodyWithResponse(context.Background(), body)
		if err != nil {
			return 0, err
		}
		return rsp.StatusCode(), nil
	})

	assert.Equal(t, url.Values{"topic[0]": {"news"}, "topic[1]": {"releases"}}, wire)
	require.NotNil(t, bound.plain)
	assert.Equal(t, body, *bound.plain)
}
