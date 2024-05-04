package issue1298_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	issue1298 "github.com/deepmap/oapi-codegen/v2/internal/test/issues/issue-1298"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type testStrictServerInterface struct {
	t *testing.T
}

// (GET /test)
func (s *testStrictServerInterface) Test(ctx context.Context, request issue1298.TestRequestObject) (issue1298.TestResponseObject, error) {
	assert.Equal(s.t, "test1", request.Body.Field1)
	assert.Equal(s.t, "test2", request.Body.Field2)
	return issue1298.Test204Response{}, nil
}

func TestIssue1298(t *testing.T) {
	g := gin.Default()
	issue1298.RegisterHandlersWithOptions(g,
		issue1298.NewStrictHandler(&testStrictServerInterface{t: t}, nil),
		issue1298.GinServerOptions{})
	ts := httptest.NewServer(g)
	defer ts.Close()

	c, err := issue1298.NewClientWithResponses(ts.URL)
	assert.NoError(t, err)
	res, err := c.TestWithApplicationTestPlusJSONBodyWithResponse(
		context.TODO(),
		issue1298.Test{
			Field1: "test1",
			Field2: "test2",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, res.StatusCode())
}
