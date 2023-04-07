package testutil

import (
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
)

type FiberHandler struct {
	App *fiber.App
	T   *testing.T
}

func (h FiberHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response, err := h.App.Test(r)
	if err != nil {
		h.T.Fatalf("Fiber.Test: %s", err.Error())
	}

	defer response.Body.Close()

	b, err := io.ReadAll(response.Body)
	if err != nil {
		h.T.Fatalf("FiberHander.Response.Body: %s", err.Error())
	}

	for key, value := range response.Header {
		if len(value) > 0 {
			w.Header().Set(key, value[0])
		}
	}

	w.WriteHeader(response.StatusCode)
	_, err = w.Write(b)
	if err != nil {
		h.T.Fatalf("FiberHander.WriteHeader.: %s", err.Error())
	}
}
