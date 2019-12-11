package testutil

import (
	"encoding/json"
	"sync"
)

func init() {
	knownHandlers = make(map[string]ResponseHandler)

	RegisterResponseHandler("application/json", jsonHandler)
}

var (
	knownHandlersMu sync.Mutex
	knownHandlers   map[string]ResponseHandler
)

type ResponseHandler func(raw []byte, obj interface{}) error

func RegisterResponseHandler(mime string, handler ResponseHandler) {
	knownHandlersMu.Lock()
	defer knownHandlersMu.Unlock()

	knownHandlers[mime] = handler
}

func getHandler(mime string) ResponseHandler {
	knownHandlersMu.Lock()
	defer knownHandlersMu.Unlock()

	return knownHandlers[mime]
}

// This function assumes that the response contains JSON and unmarshals it
// into the specified object.
func jsonHandler(raw []byte, obj interface{}) error {
	return json.Unmarshal(raw, obj)
}
