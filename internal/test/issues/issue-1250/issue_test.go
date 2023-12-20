package issue1250_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	issue1250 "github.com/deepmap/oapi-codegen/v2/internal/test/issues/issue-1250"
	"github.com/stretchr/testify/assert"
)

type testStrictServerInterface struct {
	t *testing.T
}

func (s *testStrictServerInterface) Test(ctx context.Context, request issue1250.TestRequestObject) (issue1250.TestResponseObject, error) {
	s1 := "foo"
	r := issue1250.Test{
		Field1: &s1,
	}
	r.Set("Extra", "bar")
	return issue1250.Test200JSONResponse(r), nil
}

func TestIssue1250(t *testing.T) {
	si := issue1250.NewStrictHandler(&testStrictServerInterface{
		t: t,
	}, nil)

	req := httptest.NewRequest("GET", "http://localhost/test", nil)
	rec := httptest.NewRecorder()
	si.Test(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"field1": "foo", "Extra": "bar"}`, rec.Body.String())
}
