package serversrouterstrailingslash

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type server struct {
	lastOp string
	lastID string
}

func (s *server) Test1(w http.ResponseWriter, r *http.Request, id string) {
	s.lastOp, s.lastID = "test1", id
	w.WriteHeader(http.StatusOK)
}

func (s *server) Test2(w http.ResponseWriter, r *http.Request) {
	s.lastOp = "test2"
	w.WriteHeader(http.StatusOK)
}

// TestTrailingSlashRoutes covers issue #2065: the two spec paths overlap
// ambiguously as ServeMux subtree patterns and used to panic inside
// HandlerWithOptions; as {$}-anchored patterns they register and route by
// OpenAPI's exact-path semantics.
func TestTrailingSlashRoutes(t *testing.T) {
	s := &server{}

	// Registration itself was the panic in the issue.
	var h http.Handler
	require.NotPanics(t, func() {
		h = Handler(s)
	})

	do := func(path string) (*server, *httptest.ResponseRecorder) {
		s.lastOp, s.lastID = "", ""
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return s, rec
	}

	// Exact paths route to their handlers.
	got, rec := do("/api/test/42/test2/")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test1", got.lastOp)
	assert.Equal(t, "42", got.lastID)

	got, rec = do("/api/test/test3/")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test2", got.lastOp)

	// The ambiguous path from the issue's panic message matches the
	// parameterized route exactly (id="test3"), not both.
	got, rec = do("/api/test/test3/test2/")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test1", got.lastOp)
	assert.Equal(t, "test3", got.lastID)

	// {$}-anchored patterns are exact matches, not subtrees: paths below a
	// trailing-slash route are 404, not swallowed by it.
	_, rec = do("/api/test/test3/deeper")
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// ServeMux still issues its trailing-slash redirect for the anchored
	// pattern: a request without the trailing slash is redirected to the
	// canonical path rather than 404ing. The redirect status differs
	// across Go releases (301 on older toolchains, 307 on newer), so pin
	// the redirect and its target rather than the exact code.
	_, rec = do("/api/test/test3")
	assert.True(t, rec.Code == http.StatusMovedPermanently || rec.Code == http.StatusTemporaryRedirect,
		"expected a trailing-slash redirect, got %d", rec.Code)
	assert.Equal(t, "/api/test/test3/", rec.Header().Get("Location"))
}
