package multijson_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	multijson "github.com/deepmap/oapi-codegen/internal/test/issues/issue-1208-1209"
	"github.com/stretchr/testify/assert"
)

func TestIssueMultiJson(t *testing.T) {
	// Status code 200, application/foo+json
	bodyFoo := []byte(`{"field1": "foo"}`)
	rawResponse := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(bodyFoo)),
		Header:     http.Header{},
	}
	rawResponse.Header.Add("Content-type", "application/foo+json")

	response, err := multijson.ParseTestResponse(rawResponse)
	assert.NoError(t, err)
	assert.NotNil(t, response.ApplicationfooJSON200)
	assert.NotNil(t, response.ApplicationfooJSON200.Field1)
	assert.Equal(t, "foo", *response.ApplicationfooJSON200.Field1)
	assert.Nil(t, response.ApplicationbarJSON200)
	assert.Nil(t, response.ApplicationfooJSON201)
	assert.Nil(t, response.ApplicationbarJSON201)

	// Status code 200, application/bar+json
	bodyBar := []byte(`{"field2": "bar"}`)
	rawResponse = &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(bodyBar)),
		Header:     http.Header{},
	}
	rawResponse.Header.Add("Content-type", "application/bar+json")

	response, err = multijson.ParseTestResponse(rawResponse)
	assert.NoError(t, err)
	assert.Nil(t, response.ApplicationfooJSON200)
	assert.NotNil(t, response.ApplicationbarJSON200)
	assert.NotNil(t, response.ApplicationbarJSON200.Field2)
	assert.Equal(t, "bar", *response.ApplicationbarJSON200.Field2)
	assert.Nil(t, response.ApplicationfooJSON201)
	assert.Nil(t, response.ApplicationbarJSON201)

	// Status code 200, application/qux+json
	bodyQux := []byte(`{"field4": "Qux"}`)
	rawResponse = &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(bodyQux)),
		Header:     http.Header{},
	}
	rawResponse.Header.Add("Content-type", "application/qux+json")

	response, err = multijson.ParseTestResponse(rawResponse)
	assert.NoError(t, err)
	assert.Nil(t, response.ApplicationfooJSON200)
	assert.Nil(t, response.ApplicationbarJSON200)
	assert.Nil(t, response.ApplicationfooJSON201)
	assert.Nil(t, response.ApplicationbarJSON201)

	// Status code 201, application/foo+json
	rawResponse = &http.Response{
		StatusCode: 201,
		Body:       io.NopCloser(bytes.NewReader(bodyFoo)),
		Header:     http.Header{},
	}
	rawResponse.Header.Add("Content-type", "application/foo+json")

	response, err = multijson.ParseTestResponse(rawResponse)
	assert.NoError(t, err)
	assert.Nil(t, response.ApplicationfooJSON200)
	assert.Nil(t, response.ApplicationbarJSON200)
	assert.NotNil(t, response.ApplicationfooJSON201)
	assert.NotNil(t, response.ApplicationfooJSON201.Field1)
	assert.Equal(t, "foo", *response.ApplicationfooJSON201.Field1)
	assert.Nil(t, response.ApplicationbarJSON201)

	// Status code 201, application/bar+json
	rawResponse = &http.Response{
		StatusCode: 201,
		Body:       io.NopCloser(bytes.NewReader(bodyBar)),
		Header:     http.Header{},
	}
	rawResponse.Header.Add("Content-type", "application/bar+json")

	response, err = multijson.ParseTestResponse(rawResponse)
	assert.NoError(t, err)
	assert.Nil(t, response.ApplicationfooJSON200)
	assert.Nil(t, response.ApplicationbarJSON200)
	assert.Nil(t, response.ApplicationfooJSON201)
	assert.NotNil(t, response.ApplicationbarJSON201)
	assert.NotNil(t, response.ApplicationbarJSON201.Field2)
	assert.Equal(t, "bar", *response.ApplicationbarJSON201.Field2)

	// Status code 201, application/qux+json
	rawResponse = &http.Response{
		StatusCode: 201,
		Body:       io.NopCloser(bytes.NewReader(bodyQux)),
		Header:     http.Header{},
	}
	rawResponse.Header.Add("Content-type", "application/qux+json")

	response, err = multijson.ParseTestResponse(rawResponse)
	assert.NoError(t, err)
	assert.Nil(t, response.ApplicationfooJSON200)
	assert.Nil(t, response.ApplicationbarJSON200)
	assert.Nil(t, response.ApplicationfooJSON201)
	assert.Nil(t, response.ApplicationbarJSON201)
}
