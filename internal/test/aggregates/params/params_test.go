package aggregatesparams

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// server reports the union parameters it bound, in their JSON form.
type server struct{}

func (server) GetThing(_ context.Context, req GetThingRequestObject) (GetThingResponseObject, error) {
	bound := GetThing200JSONResponse{}
	put := func(name string, v json.Marshaler) error {
		b, err := v.MarshalJSON()
		bound[name] = json.RawMessage(b)
		return err
	}
	if err := put("accept", req.Accept); err != nil {
		return nil, err
	}
	if req.Params.Q != nil {
		if err := put("q", req.Params.Q); err != nil {
			return nil, err
		}
	}
	if req.Params.R != nil {
		if err := put("r", req.Params.R); err != nil {
			return nil, err
		}
	}
	if req.Params.XAmount != nil {
		if err := put("amount", req.Params.XAmount); err != nil {
			return nil, err
		}
	}
	if req.Params.N != nil {
		if err := put("n", req.Params.N); err != nil {
			return nil, err
		}
	}
	return bound, nil
}

func (server) GetItem(_ context.Context, req GetItemRequestObject) (GetItemResponseObject, error) {
	bound := GetItem200JSONResponse{}
	put := func(name string, v json.Marshaler) error {
		b, err := v.MarshalJSON()
		bound[name] = json.RawMessage(b)
		return err
	}
	if err := put("id", req.Id); err != nil {
		return nil, err
	}
	if req.Params.Filter != nil {
		if err := put("filter", req.Params.Filter); err != nil {
			return nil, err
		}
	}
	if req.Params.Session != nil {
		if err := put("session", req.Params.Session); err != nil {
			return nil, err
		}
	}
	return bound, nil
}

func newClient(t *testing.T) *ClientWithResponses {
	t.Helper()
	srv := httptest.NewServer(Handler(NewStrictHandler(server{}, nil)))
	t.Cleanup(srv.Close)
	c, err := NewClientWithResponses(srv.URL)
	require.NoError(t, err)
	return c
}

func TestUnionParametersRoundTrip(t *testing.T) {
	c := newClient(t)

	t.Run("scalar branches", func(t *testing.T) {
		var accept GetThingParamsAccept
		require.NoError(t, accept.FromGetThingParamsAccept0(true))
		var q GetThingParamsQ
		require.NoError(t, q.FromGetThingParamsQ0(5))
		var r IntOrString
		require.NoError(t, r.FromIntOrString0(7))
		var amount GetThingParamsXAmount
		require.NoError(t, amount.FromGetThingParamsXAmount0(1.5))

		resp, err := c.GetThingWithResponse(context.Background(), accept, &GetThingParams{Q: &q, R: &r, XAmount: &amount})
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode(), string(resp.Body))
		assert.JSONEq(t, `{"accept": true, "q": 5, "r": 7, "amount": 1.5}`, string(resp.Body))
	})

	t.Run("string branches", func(t *testing.T) {
		var accept GetThingParamsAccept
		require.NoError(t, accept.FromGetThingParamsAccept1("text/plain"))
		var q GetThingParamsQ
		require.NoError(t, q.FromGetThingParamsQ1("hello"))
		var r IntOrString
		require.NoError(t, r.FromIntOrString1("abc"))
		var amount GetThingParamsXAmount
		require.NoError(t, amount.FromGetThingParamsXAmount1(false))

		resp, err := c.GetThingWithResponse(context.Background(), accept, &GetThingParams{Q: &q, R: &r, XAmount: &amount})
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode(), string(resp.Body))
		assert.JSONEq(t, `{"accept": "text/plain", "q": "hello", "r": "abc", "amount": false}`, string(resp.Body))
	})

	t.Run("optional parameters left out", func(t *testing.T) {
		var accept GetThingParamsAccept
		require.NoError(t, accept.FromGetThingParamsAccept1("x"))
		resp, err := c.GetThingWithResponse(context.Background(), accept, &GetThingParams{})
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode(), string(resp.Body))
		assert.JSONEq(t, `{"accept": "x"}`, string(resp.Body))
	})
}

// TestUnmarshalText pins how parameter text maps to union branches: text is
// only taken as a non-string type the union actually has.
func TestUnmarshalText(t *testing.T) {
	for _, tc := range []struct {
		text string
		want string
	}{
		{"5", `5`},
		{"-3", `-3`},
		{"5.5", `"5.5"`}, // not an integer, so the string branch
		{"true", `"true"`},
		{"abc", `"abc"`},
		{"", `""`},
		// Only text that is exactly one JSON value, with nothing around it,
		// is taken as a number.
		{"123}", `"123}"`},
		{"123]", `"123]"`},
		{" 5", `" 5"`},
		{"5 ", `"5 "`},
	} {
		var q GetThingParamsQ
		require.NoError(t, q.UnmarshalText([]byte(tc.text)), tc.text)
		b, err := q.MarshalJSON()
		require.NoError(t, err)
		assert.JSONEq(t, tc.want, string(b), "text %q", tc.text)
	}

	var amount GetThingParamsXAmount
	require.NoError(t, amount.UnmarshalText([]byte("1e3")))
	b, _ := amount.MarshalJSON()
	assert.JSONEq(t, `1000`, string(b))
	assert.Error(t, amount.UnmarshalText([]byte("abc")), "number|boolean has no string branch")
	assert.Error(t, amount.UnmarshalText([]byte("1}")), "malformed text is not a number")
	assert.Error(t, amount.UnmarshalText([]byte("true]")), "malformed text is not a boolean")
}

// TestSharedComponentAndCookieUnionParameters covers a union parameter shared
// by the path item (named IdParam, as the spec also has a schema named Id), a
// union declared under components/parameters, and a union in a cookie.
func TestSharedComponentAndCookieUnionParameters(t *testing.T) {
	c := newClient(t)

	var id IdParam
	require.NoError(t, id.FromId0(42))
	var filter Filter
	require.NoError(t, filter.FromFilter1("recent"))
	var session GetItemParamsSession
	require.NoError(t, session.FromGetItemParamsSession1(true))

	resp, err := c.GetItemWithResponse(context.Background(), id, &GetItemParams{Filter: &filter, Session: &session})
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode(), string(resp.Body))
	assert.JSONEq(t, `{"id": 42, "filter": "recent", "session": true}`, string(resp.Body))

	require.NoError(t, id.FromId1("abc"))
	require.NoError(t, filter.FromFilter0(7))
	resp, err = c.GetItemWithResponse(context.Background(), id, &GetItemParams{Filter: &filter})
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode(), string(resp.Body))
	assert.JSONEq(t, `{"id": "abc", "filter": 7}`, string(resp.Body))
}

// TestUnionOfUnionParameter: a union parameter whose branch is a $ref to
// another union of scalars binds like any scalar union.
func TestUnionOfUnionParameter(t *testing.T) {
	c := newClient(t)
	var accept GetThingParamsAccept
	require.NoError(t, accept.FromGetThingParamsAccept1("x"))

	for _, tc := range []struct {
		name string
		set  func(*GetThingParamsN) error
		want string
	}{
		{"integer through the referenced union", func(n *GetThingParamsN) error {
			var inner IntOrString
			if err := inner.FromIntOrString0(5); err != nil {
				return err
			}
			return n.FromIntOrString(inner)
		}, `5`},
		{"string through the referenced union", func(n *GetThingParamsN) error {
			var inner IntOrString
			if err := inner.FromIntOrString1("abc"); err != nil {
				return err
			}
			return n.FromIntOrString(inner)
		}, `"abc"`},
		{"boolean", func(n *GetThingParamsN) error { return n.FromGetThingParamsN1(true) }, `true`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var n GetThingParamsN
			require.NoError(t, tc.set(&n))
			resp, err := c.GetThingWithResponse(context.Background(), accept, &GetThingParams{N: &n})
			require.NoError(t, err)
			require.Equal(t, 200, resp.StatusCode(), string(resp.Body))
			assert.JSONEq(t, `{"accept": "x", "n": `+tc.want+`}`, string(resp.Body))
		})
	}
}
