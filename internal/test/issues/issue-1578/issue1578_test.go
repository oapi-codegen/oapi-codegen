package issue1578_test

import (
	"bytes"
	"context"
	issue1578 "github.com/deepmap/oapi-codegen/v2/internal/test/issues/issue-1578"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http/httptest"
	"testing"
)

var _ issue1578.StrictServerInterface = &testStrictServerInterface{}

type testStrictServerInterface struct {
}

func (t testStrictServerInterface) Test(_ context.Context, _ issue1578.TestRequestObject) (issue1578.TestResponseObject, error) {
	return issue1578.Test200JSONResponse(issue1578.ComplexObject{
		AdditionalProperties: map[string]interface{}{
			"message": "hello",
		},
	}), nil
}

func TestIssue1578(t *testing.T) {
	responseObject, _ := testStrictServerInterface{}.Test(context.Background(), issue1578.TestRequestObject{})
	response := &httptest.ResponseRecorder{
		Body: new(bytes.Buffer),
	}
	err := responseObject.VisitTestResponse(response)
	require.NoError(t, err)
	data, _ := io.ReadAll(response.Body)
	assert.JSONEq(t, `{"message":"hello"}`, string(data))
}
