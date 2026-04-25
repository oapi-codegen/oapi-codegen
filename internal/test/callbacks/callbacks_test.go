// Package callbacks tests OpenAPI callback codegen end-to-end. Spec is
// OpenAPI 3.0 to verify callbacks aren't accidentally gated on 3.1.
package callbacks

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

type fakeReceiver struct {
	gotMethod      string
	gotContentType string
	gotResult      TreePlantingResult
	called         bool
}

func (f *fakeReceiver) HandleTreePlantedCallback(w http.ResponseWriter, r *http.Request) {
	f.called = true
	f.gotMethod = r.Method
	f.gotContentType = r.Header.Get("Content-Type")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if err := json.Unmarshal(body, &f.gotResult); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func TestCallbackRoundTrip(t *testing.T) {
	receiver := &fakeReceiver{}

	mux := http.NewServeMux()
	mux.Handle("POST /tree-planted", TreePlantedCallbackHandler(receiver))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	initiator, err := NewCallbackInitiator()
	require.NoError(t, err)

	result := TreePlantingResult{
		Id:      "tree-42",
		Success: true,
	}

	resp, err := initiator.TreePlanted(context.Background(), srv.URL+"/tree-planted", result)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	require.True(t, receiver.called, "callback handler should have been called")
	assert.Equal(t, "POST", receiver.gotMethod)
	assert.Equal(t, "application/json", receiver.gotContentType)
	assert.Equal(t, result, receiver.gotResult)
}

// TestCallbackInitiatorRequestEditor verifies WithCallbackRequestEditorFn
// composes onto every outgoing request, mirroring the path Client and
// the WebhookInitiator behavior.
func TestCallbackInitiatorRequestEditor(t *testing.T) {
	const sigHeader = "X-Callback-Signature"
	const sigValue = "t=1,v1=abc"

	receiver := &capturingReceiver{}
	mux := http.NewServeMux()
	mux.Handle("POST /cb", TreePlantedCallbackHandler(receiver))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	initiator, err := NewCallbackInitiator(
		WithCallbackRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set(sigHeader, sigValue)
			return nil
		}),
	)
	require.NoError(t, err)

	resp, err := initiator.TreePlanted(context.Background(), srv.URL+"/cb", TreePlantingResult{
		Id:      "x",
		Success: false,
	})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, sigValue, receiver.gotHeaders.Get(sigHeader))
}

type capturingReceiver struct {
	gotHeaders http.Header
}

func (c *capturingReceiver) HandleTreePlantedCallback(w http.ResponseWriter, r *http.Request) {
	c.gotHeaders = r.Header.Clone()
	w.WriteHeader(http.StatusNoContent)
}
