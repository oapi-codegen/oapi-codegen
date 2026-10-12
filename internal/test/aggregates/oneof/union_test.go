package aggregatesoneof

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnionResponseMarshalsUnderlying is a regression test for
// https://github.com/oapi-codegen/oapi-codegen/issues/970.
//
// The strict-server response type for a content schema that is a $ref to a
// oneOf/anyOf union must encode as the union's JSON, not {}. Named defined
// types do not inherit methods, so we generate a delegating MarshalJSON.
func TestUnionResponseMarshalsUnderlying(t *testing.T) {
	var ev Event
	require.NoError(t, ev.FromOnetimeEvent(OnetimeEvent{
		Kind: Onetime,
		Name: "birthday",
	}))

	resp := GetEvent200JSONResponse(ev)

	got, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.JSONEq(t, `{"kind":"onetime","name":"birthday"}`, string(got),
		"union response must marshal via delegating MarshalJSON, not as {}")
}

// TestUnionResponseVisitWritesBody verifies the end-to-end strict-server Visit
// path — the HTTP response body must contain the union's JSON.
func TestUnionResponseVisitWritesBody(t *testing.T) {
	var ev Event
	require.NoError(t, ev.FromRepeatableEvent(RepeatableEvent{
		Kind:     Repeatable,
		Interval: "weekly",
	}))

	resp := GetEvent200JSONResponse(ev)
	w := httptest.NewRecorder()

	require.NoError(t, resp.VisitGetEventResponse(w))
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"kind":"repeatable","interval":"weekly"}`, w.Body.String())
}

// TestUnionResponseRoundtrip verifies the delegating UnmarshalJSON also works,
// so clients parsing a response body can recover the union value.
func TestUnionResponseRoundtrip(t *testing.T) {
	src := []byte(`{"kind":"onetime","name":"release"}`)

	var resp GetEvent200JSONResponse
	require.NoError(t, json.Unmarshal(src, &resp))

	ev := Event(resp)
	onetime, err := ev.AsOnetimeEvent()
	require.NoError(t, err)
	assert.Equal(t, Onetime, onetime.Kind)
	assert.Equal(t, "release", onetime.Name)
}

// TestAllOfUnionResponseVisitWritesBody is a regression test for
// https://github.com/oapi-codegen/oapi-codegen/issues/2549.
//
// Thing is an allOf whose second member is a oneOf, so the model stores the
// union in an unexported field and needs its own MarshalJSON. The strict
// response type must delegate to it, or the union branch is silently dropped.
func TestAllOfUnionResponseVisitWritesBody(t *testing.T) {
	thing := Thing{Id: "abc"}
	require.NoError(t, thing.FromCat(Cat{Meow: ptr("purr")}))

	resp := GetThing200JSONResponse(thing)
	w := httptest.NewRecorder()

	require.NoError(t, resp.VisitGetThingResponse(w))
	assert.JSONEq(t, `{"id":"abc","meow":"purr"}`, w.Body.String())

	var decoded GetThing200JSONResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &decoded))
	thing = Thing(decoded)
	assert.Equal(t, "abc", thing.Id)
	cat, err := thing.AsCat()
	require.NoError(t, err)
	assert.Equal(t, ptr("purr"), cat.Meow)
}

// TestAdditionalPropertiesResponseVisitWritesBody is a regression test for
// https://github.com/oapi-codegen/oapi-codegen/issues/2549.
//
// Bag has both fixed properties and additionalProperties, so the extra fields
// live in a map tagged `json:"-"` and are only written by Bag's MarshalJSON.
func TestAdditionalPropertiesResponseVisitWritesBody(t *testing.T) {
	bag := Bag{Kind: "demo", AdditionalProperties: map[string]any{"extra": float64(42)}}

	resp := GetBag200JSONResponse(bag)
	w := httptest.NewRecorder()

	require.NoError(t, resp.VisitGetBagResponse(w))
	assert.JSONEq(t, `{"kind":"demo","extra":42}`, w.Body.String())

	var decoded GetBag200JSONResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &decoded))
	assert.Equal(t, bag, Bag(decoded))
}

// TestUnionAndAdditionalPropertiesResponseVisitWritesBody is a control for
// https://github.com/oapi-codegen/oapi-codegen/issues/2549: a top-level oneOf
// alongside additionalProperties was already delegated by the #970 fix, and
// must stay that way.
func TestUnionAndAdditionalPropertiesResponseVisitWritesBody(t *testing.T) {
	both := Both{Kind: "demo", AdditionalProperties: map[string]any{"extra": float64(1)}}
	require.NoError(t, both.FromCat(Cat{Meow: ptr("purr")}))

	resp := GetBoth200JSONResponse(both)
	w := httptest.NewRecorder()

	require.NoError(t, resp.VisitGetBothResponse(w))
	assert.JSONEq(t, `{"kind":"demo","extra":1,"meow":"purr"}`, w.Body.String())
}

// TestInlineAllOfUnionResponseVisitWritesBody is a regression test for
// https://github.com/oapi-codegen/oapi-codegen/issues/2549.
//
// An inline allOf+oneOf at the response root is hoisted to a named body type,
// and the envelope is an alias of it, so MarshalJSON is inherited. The visitor
// must not bypass it by encoding the raw union field, which would drop the
// fixed properties merged in from the allOf.
func TestInlineAllOfUnionResponseVisitWritesBody(t *testing.T) {
	body := GetInlineThing200JSONResponseBody{Id: "abc"}
	require.NoError(t, body.FromCat(Cat{Meow: ptr("purr")}))

	resp := GetInlineThing200JSONResponse(body)
	w := httptest.NewRecorder()

	require.NoError(t, resp.VisitGetInlineThingResponse(w))
	assert.JSONEq(t, `{"id":"abc","meow":"purr"}`, w.Body.String())
}

// TestRenamedUnionResponseVisitWritesBody is a regression test for
// https://github.com/oapi-codegen/oapi-codegen/issues/2549.
//
// x-go-type-name moves the union model to a renamed type and leaves an alias
// behind, so the response envelope is a defined type over the renamed model
// and still needs a delegator.
func TestRenamedUnionResponseVisitWritesBody(t *testing.T) {
	var ev RenamedEvent
	require.NoError(t, ev.FromCat(Cat{Meow: ptr("purr")}))

	resp := GetRenamed200JSONResponse(ev)
	w := httptest.NewRecorder()

	require.NoError(t, resp.VisitGetRenamedResponse(w))
	assert.JSONEq(t, `{"meow":"purr"}`, w.Body.String())
}

func ptr[T any](v T) *T {
	return &v
}
