package issue1329_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	issue1329 "github.com/deepmap/oapi-codegen/v2/internal/test/issues/issue-1329"
	"github.com/stretchr/testify/assert"
)

func TestIssue1329(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/bad+json")
		w.WriteHeader(204)
	}))
	defer ts.Close()

	c, err := issue1329.NewClientWithResponses(ts.URL)
	assert.NoError(t, err)
	res, err := c.TestWithApplicationTestPlusJSONBodyWithResponse(
		context.TODO(),
		issue1329.Test{
			Field1: "test1",
			Field2: "test2",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, 204, res.StatusCode())
}
