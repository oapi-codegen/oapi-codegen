package main

import (
	"bytes"
	_ "embed"
	"go/format"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	examplePetstoreClient "github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded"
	examplePetstore "github.com/oapi-codegen/oapi-codegen/v2/examples/petstore-expanded/echo/api"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
	"github.com/stretchr/testify/assert"
	"golang.org/x/lint"
)

func checkLint(t *testing.T, filename string, code []byte) {
	linter := new(lint.Linter)
	problems, err := linter.Lint(filename, code)
	assert.NoError(t, err)
	assert.Len(t, problems, 0)
}

func TestExamplePetStoreCodeGeneration(t *testing.T) {

	// Input vars for code generation:
	packageName := "api"
	opts := codegen.Configuration{
		PackageName: packageName,
		Generate: codegen.GenerateOptions{
			EchoServer:   true,
			Client:       true,
			Models:       true,
			EmbeddedSpec: true,
		},
	}

	// Get a spec from the example PetStore definition:
	swagger, err := examplePetstore.GetSwagger()
	assert.NoError(t, err)

	// Run our code generation:
	code, err := codegen.Generate(swagger, opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, code)

	// Check that we have valid (formattable) code:
	_, err = format.Source([]byte(code))
	assert.NoError(t, err)

	// Check that we have a package:
	assert.Contains(t, code, "package api")

	// Check that the client method signatures return response structs:
	assert.Contains(t, code, "func (c *Client) FindPetByID(ctx context.Context, id int64, reqEditors ...RequestEditorFn) (*http.Response, error) {")

	// Check that the property comments were generated
	assert.Contains(t, code, "// Id Unique id of the pet")

	// Check that the summary comment contains newlines
	assert.Contains(t, code, `// Deletes a pet by ID
	// (DELETE /pets/{id})
`)

	// Make sure the generated code is valid:
	checkLint(t, "test.gen.go", []byte(code))
}

func TestExamplePetStoreCodeGenerationWithUserTemplates(t *testing.T) {

	userTemplates := map[string]string{"typedef.tmpl": "//blah\n//blah"}

	// Input vars for code generation:
	packageName := "api"
	opts := codegen.Configuration{
		PackageName: packageName,
		Generate: codegen.GenerateOptions{
			Models: true,
		},
		OutputOptions: codegen.OutputOptions{
			UserTemplates: userTemplates,
		},
	}

	// Get a spec from the example PetStore definition:
	swagger, err := examplePetstore.GetSwagger()
	assert.NoError(t, err)

	// Run our code generation:
	code, err := codegen.Generate(swagger, opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, code)

	// Check that we have valid (formattable) code:
	_, err = format.Source([]byte(code))
	assert.NoError(t, err)

	// Check that we have a package:
	assert.Contains(t, code, "package api")

	// Check that the built-in template has been overridden
	assert.Contains(t, code, "//blah")
}

func TestExamplePetStoreCodeGenerationWithFileUserTemplates(t *testing.T) {

	userTemplates := map[string]string{"typedef.tmpl": "../../../pkg/codegen/templates/typedef.tmpl"}

	// Input vars for code generation:
	packageName := "api"
	opts := codegen.Configuration{
		PackageName: packageName,
		Generate: codegen.GenerateOptions{
			Models: true,
		},
		OutputOptions: codegen.OutputOptions{
			UserTemplates: userTemplates,
		},
	}

	// Get a spec from the example PetStore definition:
	swagger, err := examplePetstore.GetSwagger()
	assert.NoError(t, err)

	// Run our code generation:
	code, err := codegen.Generate(swagger, opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, code)

	// Check that we have valid (formattable) code:
	_, err = format.Source([]byte(code))
	assert.NoError(t, err)

	// Check that we have a package:
	assert.Contains(t, code, "package api")

	// Check that the built-in template has been overridden
	assert.Contains(t, code, "// Package api provides primitives to interact with the openapi")
}

func TestExamplePetStoreCodeGenerationWithHTTPUserTemplates(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, writeErr := w.Write([]byte("//blah"))
		assert.NoError(t, writeErr)
	}))

	userTemplates := map[string]string{"typedef.tmpl": srv.URL}

	// Input vars for code generation:
	packageName := "api"
	opts := codegen.Configuration{
		PackageName: packageName,
		Generate: codegen.GenerateOptions{
			Models: true,
		},
		OutputOptions: codegen.OutputOptions{
			UserTemplates: userTemplates,
		},
	}

	// Get a spec from the example PetStore definition:
	swagger, err := examplePetstore.GetSwagger()
	assert.NoError(t, err)

	// Run our code generation:
	code, err := codegen.Generate(swagger, opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, code)

	// Check that we have valid (formattable) code:
	_, err = format.Source([]byte(code))
	assert.NoError(t, err)

	// Check that we have a package:
	assert.Contains(t, code, "package api")

	// Check that the built-in template has been overridden
	assert.Contains(t, code, "//blah")
}
func TestExamplePetStoreParseFunction(t *testing.T) {

	bodyBytes := []byte(`{"id": 5, "name": "testpet", "tag": "cat"}`)

	cannedResponse := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(bodyBytes)),
		Header:     http.Header{},
	}
	cannedResponse.Header.Add("Content-type", "application/json")

	findPetByIDResponse, err := examplePetstoreClient.ParseFindPetByIDResponse(cannedResponse)
	assert.NoError(t, err)
	assert.NotNil(t, findPetByIDResponse.JSON200)
	assert.Equal(t, int64(5), findPetByIDResponse.JSON200.Id)
	assert.Equal(t, "testpet", findPetByIDResponse.JSON200.Name)
	assert.NotNil(t, findPetByIDResponse.JSON200.Tag)
	assert.Equal(t, "cat", *findPetByIDResponse.JSON200.Tag)
}
