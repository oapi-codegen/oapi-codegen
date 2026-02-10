package server

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeServer struct {
	t      *testing.T
	called bool
}

// get every type optional
// (GET /every-type-optional)
func (s *fakeServer) GetEveryTypeOptional(w http.ResponseWriter, r *http.Request) {
	// not implemented
	w.WriteHeader(http.StatusTeapot)
}

// Get resource via simple path
// (GET /get-simple)
func (s *fakeServer) GetSimple(w http.ResponseWriter, r *http.Request) {
	// not implemented
	w.WriteHeader(http.StatusTeapot)
}

// Getter with referenced parameter and referenced response
// (GET /get-with-args)
func (s *fakeServer) GetWithArgs(w http.ResponseWriter, r *http.Request, params GetWithArgsParams) {
	// not implemented
	w.WriteHeader(http.StatusTeapot)
}

// Getter with referenced parameter and referenced response
// (GET /get-with-references/{global_argument}/{argument})
func (s *fakeServer) GetWithReferences(w http.ResponseWriter, r *http.Request, globalArgument int64, argument Argument) {
	// not implemented
	w.WriteHeader(http.StatusTeapot)
}

// Get an object by ID
// (GET /get-with-type/{content_type})
func (s *fakeServer) GetWithContentType(w http.ResponseWriter, r *http.Request, contentType GetWithContentTypeParamsContentType) {
	// not implemented
	w.WriteHeader(http.StatusTeapot)
}

// get with reserved keyword
// (GET /reserved-keyword)
func (s *fakeServer) GetReservedKeyword(w http.ResponseWriter, r *http.Request) {
	// not implemented
	w.WriteHeader(http.StatusTeapot)
}

// Create a resource
// (POST /resource/{argument})
func (s *fakeServer) CreateResource(w http.ResponseWriter, r *http.Request, argument Argument) {
	// not implemented
	w.WriteHeader(http.StatusTeapot)
}

// Create a resource with inline parameter
// (POST /resource2/{inline_argument})
func (s *fakeServer) CreateResource2(w http.ResponseWriter, r *http.Request, inlineArgument int, params CreateResource2Params) {
	assert.Equal(s.t, 99, *params.InlineQueryArgument)
	assert.Equal(s.t, 1, inlineArgument)
	s.called = true
}

// Update a resource with inline body. The parameter name is a reserved
// keyword, so make sure that gets prefixed to avoid syntax errors
// (PUT /resource3/{fallthrough})
func (s *fakeServer) UpdateResource3(w http.ResponseWriter, r *http.Request, pFallthrough int) {
	// not implemented
	w.WriteHeader(http.StatusTeapot)
}

// get response with reference
// (GET /response-with-reference)
func (s *fakeServer) GetResponseWithReference(w http.ResponseWriter, r *http.Request) {
	// not implemented
	w.WriteHeader(http.StatusTeapot)
}

func TestParameters(t *testing.T) {
	m := fakeServer{
		t: t,
	}

	h := Handler(&m)

	req := httptest.NewRequest("POST", "http://openapitest.deepmap.ai/resource2/1?inline_query_argument=99", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	assert.True(t, m.called)
}

func TestErrorHandlerFunc(t *testing.T) {
	m := fakeServer{}

	h := HandlerWithOptions(&m, ChiServerOptions{
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			w.Header().Set("Content-Type", "application/json")
			var requiredParamError *RequiredParamError
			assert.True(t, errors.As(err, &requiredParamError))
		},
	})

	s := httptest.NewServer(h)
	defer s.Close()

	req, err := http.DefaultClient.Get(s.URL + "/get-with-args")
	assert.Nil(t, err)
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
}

func TestErrorHandlerFuncBackwardsCompatible(t *testing.T) {
	m := fakeServer{}

	h := HandlerWithOptions(&m, ChiServerOptions{})

	s := httptest.NewServer(h)
	defer s.Close()

	req, err := http.DefaultClient.Get(s.URL + "/get-with-args")
	b, _ := io.ReadAll(req.Body)
	assert.Nil(t, err)
	assert.Equal(t, "text/plain; charset=utf-8", req.Header.Get("Content-Type"))
	assert.Equal(t, "Query argument required_argument is required, but not found\n", string(b))
}
