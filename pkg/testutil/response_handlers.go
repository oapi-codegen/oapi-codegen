package testutil

import (
	"encoding/json"
	"io"
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

// Deprecated: This has been replaced by https://pkg.go.dev/github.com/oapi-codegen/testutil#ResponseHandler
type ResponseHandler func(contentType string, raw io.Reader, obj interface{}, strict bool) error

// Deprecated: This has been replaced by https://pkg.go.dev/github.com/oapi-codegen/testutil#RegisterResponseHandler
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

// jsonHandler assumes that the response contains JSON and unmarshals it
// into the specified object.
func jsonHandler(_ string, r io.Reader, obj interface{}, strict bool) error {
	d := json.NewDecoder(r)
	if strict {
		d.DisallowUnknownFields()
	}
	return d.Decode(obj)
}
