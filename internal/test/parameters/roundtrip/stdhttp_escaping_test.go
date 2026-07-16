package parametersroundtrip_test

import (
	"net/http"
	"testing"

	stdhttpparams "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/roundtrip/stdhttp"
	stdhttpgen "github.com/oapi-codegen/oapi-codegen/v2/internal/test/parameters/roundtrip/stdhttp/gen"
)

// TestStdHttpStringEscaping registers only the /simpleString route directly,
// because the full stdhttp registration panics on the digit-leading "1param"
// route (#2306) and the shared round-trip test is skipped. This keeps the
// issue-2455 escaping coverage alive for net/http's ServeMux, whose
// r.PathValue returns decoded values.
func TestStdHttpStringEscaping(t *testing.T) {
	var s stdhttpparams.Server
	wrapper := stdhttpgen.ServerInterfaceWrapper{
		Handler: &s,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /simpleString/{param}", wrapper.GetSimpleString)
	testStringEscaping(t, mux, true)
}
