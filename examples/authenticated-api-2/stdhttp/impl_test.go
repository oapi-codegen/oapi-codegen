package stdhttp

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	var s ServerInterface = &Server{}

	t.Run("unauthenticated", func(t *testing.T) {
		t.Run("does not require authentication", func(t *testing.T) {

			r := httptest.NewRequest("GET", "/unauthenticated", nil)
			rr := httptest.NewRecorder()

			s.Unauthenticated(rr, r)

			assert.Equal(t, http.StatusOK, rr.Code)

			fmt.Println(io.ReadAll(rr.Body))
		})
	})

	t.Run("apiKey", func(t *testing.T) {
		t.Run("returns **??** when no authentication", func(t *testing.T) {

			r := httptest.NewRequest("GET", "/apiKey", nil)
			rr := httptest.NewRecorder()

			s.Unauthenticated(rr, r)

			assert.NotEqual(t, http.StatusOK, rr.Code)

			fmt.Println(io.ReadAll(rr.Body))
		})
	})
}
