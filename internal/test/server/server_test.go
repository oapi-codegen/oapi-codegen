package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParameters(t *testing.T) {
	m := ServerInterfaceMock{}

	m.CreateResource2Func = func(w http.ResponseWriter, r *http.Request) {
		params := ParamsForCreateResource2(r.Context())
		arg := r.Context().Value("inlineArgument").(int)

		assert.Equal(t, 99, *params.InlineQueryArgument)
		assert.Equal(t, 1, arg)
	}

	h := Handler(&m)

	req := httptest.NewRequest("POST", "http://openapitest.deepmap.ai/resource2/1?inline_query_argument=99", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.Equal(t, 1, len(m.CreateResource2Calls()))
}
