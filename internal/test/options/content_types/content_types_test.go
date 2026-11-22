package optionscontenttypes

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compile-time checks: the V1 short name drives the vendored-JSON type names,
// and the CSV mapping produces a model for text/csv even though its
// wire-level handling stays the untyped io.Reader passthrough.
var (
	_ AddPetV1RequestBody        = Pet{}
	_ UploadReportCSVBody        = ""
	_ UploadReportCSVRequestBody = ""
	_ AddPetResponseObject       = AddPet200V1Response{}
	_ UploadReportResponseObject = UploadReport200CSVResponse{}
	_ io.Reader                  = UploadReportRequestObject{}.Body
)

// The client deserializes the V1-mapped vendored JSON response into a typed
// field named after the short name.
func TestClientParsesV1MappedResponse(t *testing.T) {
	rsp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/vnd.mycompany.v1+json"}},
		Body:       io.NopCloser(strings.NewReader(`{"name":"fido"}`)),
	}

	parsed, err := ParseAddPetResponse(rsp)
	require.NoError(t, err)
	require.NotNil(t, parsed.V1200)
	assert.Equal(t, "fido", parsed.V1200.Name)
}

// The CSV-mapped response gets no typed client field — only the raw body —
// but the raw bytes are still captured.
func TestClientKeepsCSVMappedResponseRaw(t *testing.T) {
	rsp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/csv"}},
		Body:       io.NopCloser(strings.NewReader("a,b,c\n")),
	}

	parsed, err := ParseUploadReportResponse(rsp)
	require.NoError(t, err)
	assert.Equal(t, []byte("a,b,c\n"), parsed.GetBody())
}

// The strict envelope for the CSV-mapped response streams the body with the
// original media type.
func TestStrictCSVResponseVisit(t *testing.T) {
	w := httptest.NewRecorder()
	response := UploadReport200CSVResponse{Body: strings.NewReader("a,b,c\n")}
	require.NoError(t, response.VisitUploadReportResponse(w))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
	assert.Equal(t, "a,b,c\n", w.Body.String())
}
