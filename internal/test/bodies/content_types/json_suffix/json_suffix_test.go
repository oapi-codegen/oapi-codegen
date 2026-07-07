package jsonsuffix_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	jsonsuffix "github.com/oapi-codegen/oapi-codegen/v2/internal/test/bodies/content_types/json_suffix"
	"github.com/stretchr/testify/assert"
)

type testStrictServerInterface struct {
	t *testing.T
}

// (GET /test)
func (s *testStrictServerInterface) Test(ctx context.Context, request jsonsuffix.TestRequestObject) (jsonsuffix.TestResponseObject, error) {
	assert.Equal(s.t, "test1", request.Body.Field1)
	assert.Equal(s.t, "test2", request.Body.Field2)
	return jsonsuffix.Test204Response{}, nil
}

func TestIssue1298(t *testing.T) {
	g := gin.Default()
	jsonsuffix.RegisterHandlersWithOptions(g,
		jsonsuffix.NewStrictHandler(&testStrictServerInterface{t: t}, nil),
		jsonsuffix.GinServerOptions{})
	ts := httptest.NewServer(g)
	defer ts.Close()

	c, err := jsonsuffix.NewClientWithResponses(ts.URL)
	assert.NoError(t, err)
	res, err := c.TestWithApplicationTestPlusJSONBodyWithResponse(
		context.TODO(),
		jsonsuffix.Test{
			Field1: "test1",
			Field2: "test2",
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, res.StatusCode())
}
