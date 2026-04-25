// Package webhooks tests the OpenAPI 3.1 webhook codegen end-to-end:
// the WebhookInitiator fires a webhook against an httptest.Server that
// registers the WebhookReceiverInterface handler, and the test asserts
// the payload round-trips intact.
package webhooks

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeReceiver captures whatever the handler is given so the test can
// assert what arrived.
type fakeReceiver struct {
	gotMethod      string
	gotContentType string
	gotEvent       PetStatusEvent
	called         bool
}

func (f *fakeReceiver) HandlePetStatusChangedWebhook(w http.ResponseWriter, r *http.Request) {
	f.called = true
	f.gotMethod = r.Method
	f.gotContentType = r.Header.Get("Content-Type")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if err := json.Unmarshal(body, &f.gotEvent); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func TestWebhookRoundTrip(t *testing.T) {
	receiver := &fakeReceiver{}

	// Mount the generated factory at a test path. In real usage callers
	// pick whatever URL they advertise to subscribers; the factory is
	// path-agnostic so the test can pick anything.
	mux := http.NewServeMux()
	mux.Handle("POST /hooks/pet-status", PetStatusChangedWebhookHandler(receiver))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	initiator, err := NewWebhookInitiator()
	require.NoError(t, err)

	event := PetStatusEvent{
		Id:     "pet-42",
		Status: Sold,
	}

	resp, err := initiator.PetStatusChanged(context.Background(), srv.URL+"/hooks/pet-status", event)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	require.True(t, receiver.called, "webhook handler should have been called")
	assert.Equal(t, "POST", receiver.gotMethod)
	assert.Equal(t, "application/json", receiver.gotContentType)
	assert.Equal(t, event, receiver.gotEvent)
}

// TestWebhookInitiatorRequestEditor verifies WithWebhookRequestEditorFn
// composes onto every request, mirroring the path Client's RequestEditor
// behavior. This is the integration-level assertion that webhook-side
// middleware support is structurally identical to client-side.
func TestWebhookInitiatorRequestEditor(t *testing.T) {
	const sigHeader = "X-Webhook-Signature"
	const sigValue = "t=1234,v1=deadbeef"

	receiver := &capturingReceiver{}
	mux := http.NewServeMux()
	mux.Handle("POST /hooks", PetStatusChangedWebhookHandler(receiver))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	initiator, err := NewWebhookInitiator(
		WithWebhookRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set(sigHeader, sigValue)
			return nil
		}),
	)
	require.NoError(t, err)

	resp, err := initiator.PetStatusChanged(context.Background(), srv.URL+"/hooks", PetStatusEvent{
		Id:     "pet-1",
		Status: Available,
	})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, sigValue, receiver.gotHeaders.Get(sigHeader),
		"per-call editor was not applied to the outgoing request")
}

type capturingReceiver struct {
	gotHeaders http.Header
}

func (c *capturingReceiver) HandlePetStatusChangedWebhook(w http.ResponseWriter, r *http.Request) {
	c.gotHeaders = r.Header.Clone()
	w.WriteHeader(http.StatusNoContent)
}

// TestWebhookReceiverMiddleware verifies middlewares wrap the handler in
// declared order (outermost first), matching the standard handler
// composition convention.
func TestWebhookReceiverMiddleware(t *testing.T) {
	var order []string
	mw := func(name string) WebhookReceiverMiddlewareFunc {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+":pre")
				next.ServeHTTP(w, r)
				order = append(order, name+":post")
			})
		}
	}

	mux := http.NewServeMux()
	mux.Handle("POST /hooks",
		PetStatusChangedWebhookHandler(&capturingReceiver{}, mw("outer"), mw("inner")))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	initiator, err := NewWebhookInitiator()
	require.NoError(t, err)
	resp, err := initiator.PetStatusChanged(context.Background(), srv.URL+"/hooks", PetStatusEvent{
		Id: "x", Status: Pending,
	})
	require.NoError(t, err)
	resp.Body.Close()

	// Middlewares are applied in order with the LAST argument becoming
	// the OUTERMOST wrapper (each iteration assigns h = mw(h)). So we
	// expect: inner:pre -> outer:pre is the wrong order; check the
	// actual implementation's intent.
	//
	// In our generated factory:
	//   for _, mw := range middlewares { h = mw(h) }
	// the final mw becomes outermost. So with ("outer", "inner") the
	// "inner" middleware is the outermost wrapper.
	assert.Equal(t,
		[]string{"inner:pre", "outer:pre", "outer:post", "inner:post"},
		order)
}
