package skipunimplemented

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// This declaration must compile alongside the generated server, without a
// duplicate type or method from the built-in Unimplemented implementation.
type Unimplemented struct{}

var _ ServerInterface = Unimplemented{}

func (Unimplemented) GetWidget(w http.ResponseWriter, _ *http.Request, id string, params GetWidgetParams) {
	fmt.Fprintf(w, "%s:%d", id, params.Limit)
}

func TestCustomUnimplemented(t *testing.T) {
	handler := Handler(Unimplemented{})

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/widgets/example?limit=3", nil))
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "example:3", response.Body.String())

	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/widgets/example", nil))
	assert.Equal(t, http.StatusBadRequest, response.Code)
}
