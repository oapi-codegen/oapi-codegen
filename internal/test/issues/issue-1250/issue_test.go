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

// (GET /test)
func (s *testStrictServerInterface) Test(ctx context.Context, request issue1250.TestRequestObject) (issue1250.TestResponseObject, error) {
	s1 := "foo"
	r := issue1250.Test{
		Field1: &s1,
	}
	return issue1250.Test200JSONResponse(r), nil
}

// (GET /test-union)
func (s *testStrictServerInterface) TestUnion(ctx context.Context, request issue1250.TestUnionRequestObject) (issue1250.TestUnionResponseObject, error) {
	s1 := "foo"
	var r issue1250.TestUnion
	r.FromTestUnion0(issue1250.TestUnion0{
		Field1: &s1,
	})
	return issue1250.TestUnion200JSONResponse(r), nil
}

// (GET /test-additional-properties)
func (s *testStrictServerInterface) TestAdditionalProperties(ctx context.Context, request issue1250.TestAdditionalPropertiesRequestObject) (issue1250.TestAdditionalPropertiesResponseObject, error) {
	s1 := "foo"
	r := issue1250.TestAdditionalProperties{
		Field1: &s1,
	}
	r.Set("Extra", "bar")
	return issue1250.TestAdditionalProperties200JSONResponse(r), nil
}

// (GET /test-additional-properties-with-union)
func (s *testStrictServerInterface) TestAdditionalPropertiesWithUnion(ctx context.Context, request issue1250.TestAdditionalPropertiesWithUnionRequestObject) (issue1250.TestAdditionalPropertiesWithUnionResponseObject, error) {
	s1 := "foo"
	var r issue1250.TestAdditionalPropertiesWithUnion
	r.FromTestAdditionalPropertiesWithUnion0(issue1250.TestAdditionalPropertiesWithUnion0{
		Field1: &s1,
	})
	r.Set("Extra", "bar")
	return issue1250.TestAdditionalPropertiesWithUnion200JSONResponse(r), nil
}

// (GET /test-ref)
func (s *testStrictServerInterface) TestRef(ctx context.Context, request issue1250.TestRefRequestObject) (issue1250.TestRefResponseObject, error) {
	s1 := "foo"
	r := issue1250.Test{
		Field1: &s1,
	}
	return issue1250.TestRef200JSONResponse{TestRespJSONResponse: issue1250.TestRespJSONResponse(r)}, nil
}

// (GET /test-union-ref)
func (s *testStrictServerInterface) TestUnionRef(ctx context.Context, request issue1250.TestUnionRefRequestObject) (issue1250.TestUnionRefResponseObject, error) {
	s1 := "foo"
	var r issue1250.TestUnion
	r.FromTestUnion0(issue1250.TestUnion0{
		Field1: &s1,
	})
	return issue1250.TestUnionRef200JSONResponse{TestUnionRespJSONResponse: issue1250.TestUnionRespJSONResponse(r)}, nil
}

// (GET /test-additional-properties-ref)
func (s *testStrictServerInterface) TestAdditionalPropertiesRef(ctx context.Context, request issue1250.TestAdditionalPropertiesRefRequestObject) (issue1250.TestAdditionalPropertiesRefResponseObject, error) {
	s1 := "foo"
	r := issue1250.TestAdditionalProperties{
		Field1: &s1,
	}
	r.Set("Extra", "bar")
	return issue1250.TestAdditionalPropertiesRef200JSONResponse{TestAdditionalPropertiesRespJSONResponse: issue1250.TestAdditionalPropertiesRespJSONResponse(r)}, nil
}

// (GET /test-additional-properties-with-union-ref)
func (s *testStrictServerInterface) TestAdditionalPropertiesWithUnionRef(ctx context.Context, request issue1250.TestAdditionalPropertiesWithUnionRefRequestObject) (issue1250.TestAdditionalPropertiesWithUnionRefResponseObject, error) {
	s1 := "foo"
	var r issue1250.TestAdditionalPropertiesWithUnion
	r.FromTestAdditionalPropertiesWithUnion0(issue1250.TestAdditionalPropertiesWithUnion0{
		Field1: &s1,
	})
	r.Set("Extra", "bar")
	return issue1250.TestAdditionalPropertiesWithUnionRef200JSONResponse{TestAdditionalPropertiesWithUnionRespJSONResponse: issue1250.TestAdditionalPropertiesWithUnionRespJSONResponse(r)}, nil
}

func TestIssue1250(t *testing.T) {
	si := issue1250.NewStrictHandler(&testStrictServerInterface{
		t: t,
	}, nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !assert.Equal(t, "/test", r.URL.Path) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		si.Test(w, r)
	}))
	defer ts.Close()
	c, err := issue1250.NewClientWithResponses(ts.URL)
	assert.NoError(t, err)
	res, err := c.TestWithResponse(context.TODO())
	assert.NoError(t, err)
	s1 := "foo"
	assert.Equal(t, &issue1250.Test{
		Field1: &s1,
	}, res.JSON200)
}

func TestIssue1250Union(t *testing.T) {
	si := issue1250.NewStrictHandler(&testStrictServerInterface{
		t: t,
	}, nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !assert.Equal(t, "/test-union", r.URL.Path) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		si.TestUnion(w, r)
	}))
	defer ts.Close()
	c, err := issue1250.NewClientWithResponses(ts.URL)
	assert.NoError(t, err)
	res, err := c.TestUnionWithResponse(context.TODO())
	assert.NoError(t, err)
	s1 := "foo"
	assert.NotNil(t, res.JSON200)
	assert.EqualExportedValues(t, issue1250.TestUnion{}, *res.JSON200)
	u0, err := res.JSON200.AsTestUnion0()
	assert.NoError(t, err)
	assert.Equal(t, issue1250.TestUnion0{
		Field1: &s1,
	}, u0)
	u1, err := res.JSON200.AsTestUnion1()
	assert.NoError(t, err)
	assert.Equal(t, issue1250.TestUnion1{}, u1)
}

func TestIssue1250AdditionalProperties(t *testing.T) {
	si := issue1250.NewStrictHandler(&testStrictServerInterface{
		t: t,
	}, nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !assert.Equal(t, "/test-additional-properties", r.URL.Path) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		si.TestAdditionalProperties(w, r)
	}))
	defer ts.Close()
	c, err := issue1250.NewClientWithResponses(ts.URL)
	assert.NoError(t, err)
	res, err := c.TestAdditionalPropertiesWithResponse(context.TODO())
	assert.NoError(t, err)
	s1 := "foo"
	assert.Equal(t, &issue1250.TestAdditionalProperties{
		Field1: &s1,
		AdditionalProperties: map[string]interface{}{
			"Extra": "bar",
		},
	}, res.JSON200)
}

func TestIssue1250AdditionalPropertiesWithUnion(t *testing.T) {
	si := issue1250.NewStrictHandler(&testStrictServerInterface{
		t: t,
	}, nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !assert.Equal(t, "/test-additional-properties-with-union", r.URL.Path) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		si.TestAdditionalPropertiesWithUnion(w, r)
	}))
	defer ts.Close()
	c, err := issue1250.NewClientWithResponses(ts.URL)
	assert.NoError(t, err)
	res, err := c.TestAdditionalPropertiesWithUnionWithResponse(context.TODO())
	assert.NoError(t, err)
	s1 := "foo"
	assert.NotNil(t, res.JSON200)
	assert.EqualExportedValues(t, issue1250.TestAdditionalPropertiesWithUnion{
		AdditionalProperties: map[string]interface{}{
			"Extra":  "bar",
			"field1": s1,
		},
	}, *res.JSON200)
	u0, err := res.JSON200.AsTestAdditionalPropertiesWithUnion0()
	assert.NoError(t, err)
	assert.Equal(t, issue1250.TestAdditionalPropertiesWithUnion0{
		Field1: &s1,
	}, u0)
	u1, err := res.JSON200.AsTestAdditionalPropertiesWithUnion1()
	assert.NoError(t, err)
	assert.Equal(t, issue1250.TestAdditionalPropertiesWithUnion1{}, u1)
}

func TestIssue1250Ref(t *testing.T) {
	si := issue1250.NewStrictHandler(&testStrictServerInterface{
		t: t,
	}, nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !assert.Equal(t, "/test-ref", r.URL.Path) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		si.TestRef(w, r)
	}))
	defer ts.Close()
	c, err := issue1250.NewClientWithResponses(ts.URL)
	assert.NoError(t, err)
	res, err := c.TestRefWithResponse(context.TODO())
	assert.NoError(t, err)
	s1 := "foo"
	assert.Equal(t, &issue1250.Test{
		Field1: &s1,
	}, res.JSON200)
}

func TestIssue1250UnionRef(t *testing.T) {
	si := issue1250.NewStrictHandler(&testStrictServerInterface{
		t: t,
	}, nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !assert.Equal(t, "/test-union-ref", r.URL.Path) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		si.TestUnionRef(w, r)
	}))
	defer ts.Close()
	c, err := issue1250.NewClientWithResponses(ts.URL)
	assert.NoError(t, err)
	res, err := c.TestUnionRefWithResponse(context.TODO())
	assert.NoError(t, err)
	s1 := "foo"
	assert.NotNil(t, res.JSON200)
	assert.EqualExportedValues(t, issue1250.TestUnion{}, *res.JSON200)
	u0, err := res.JSON200.AsTestUnion0()
	assert.NoError(t, err)
	assert.Equal(t, issue1250.TestUnion0{
		Field1: &s1,
	}, u0)
	u1, err := res.JSON200.AsTestUnion1()
	assert.NoError(t, err)
	assert.Equal(t, issue1250.TestUnion1{}, u1)
}

func TestIssue1250AdditionalPropertiesRef(t *testing.T) {
	si := issue1250.NewStrictHandler(&testStrictServerInterface{
		t: t,
	}, nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !assert.Equal(t, "/test-additional-properties-ref", r.URL.Path) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		si.TestAdditionalPropertiesRef(w, r)
	}))
	defer ts.Close()
	c, err := issue1250.NewClientWithResponses(ts.URL)
	assert.NoError(t, err)
	res, err := c.TestAdditionalPropertiesRefWithResponse(context.TODO())
	assert.NoError(t, err)
	s1 := "foo"
	assert.Equal(t, &issue1250.TestAdditionalProperties{
		Field1: &s1,
		AdditionalProperties: map[string]interface{}{
			"Extra": "bar",
		},
	}, res.JSON200)
}

func TestIssue1250AdditionalPropertiesWithUnionRef(t *testing.T) {
	si := issue1250.NewStrictHandler(&testStrictServerInterface{
		t: t,
	}, nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !assert.Equal(t, "/test-additional-properties-with-union-ref", r.URL.Path) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		si.TestAdditionalPropertiesWithUnionRef(w, r)
	}))
	defer ts.Close()
	c, err := issue1250.NewClientWithResponses(ts.URL)
	assert.NoError(t, err)
	res, err := c.TestAdditionalPropertiesWithUnionRefWithResponse(context.TODO())
	assert.NoError(t, err)
	s1 := "foo"
	assert.NotNil(t, res.JSON200)
	assert.EqualExportedValues(t, issue1250.TestAdditionalPropertiesWithUnion{
		AdditionalProperties: map[string]interface{}{
			"Extra":  "bar",
			"field1": s1,
		},
	}, *res.JSON200)
	u0, err := res.JSON200.AsTestAdditionalPropertiesWithUnion0()
	assert.NoError(t, err)
	assert.Equal(t, issue1250.TestAdditionalPropertiesWithUnion0{
		Field1: &s1,
	}, u0)
	u1, err := res.JSON200.AsTestAdditionalPropertiesWithUnion1()
	assert.NoError(t, err)
	assert.Equal(t, issue1250.TestAdditionalPropertiesWithUnion1{}, u1)
}
