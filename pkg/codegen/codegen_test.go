package codegen

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"go/format"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	examplePetstoreClient "github.com/deepmap/oapi-codegen/examples/petstore-expanded"
	examplePetstore "github.com/weberr13/oapi-codegen/examples/petstore-expanded/echo/api"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/golangci/lint-1"
	"github.com/onsi/gomega/gexec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExamplePetStoreCodeGeneration(t *testing.T) {

	// Input vars for code generation:
	packageName := "api"
	opts := Options{
		GenerateClient:     true,
		GenerateEchoServer: true,
		GenerateTypes:      true,
		EmbedSpec:          true,
	}

	// Get a spec from the example PetStore definition:
	swagger, err := examplePetstore.GetSwagger()
	assert.NoError(t, err)

	// Run our code generation:
	code, err := Generate(swagger, packageName, opts)
	assert.NoError(t, err)
	assert.NotEmpty(t, code)

	// Check that we have valid (formattable) code:
	_, err = format.Source([]byte(code))
	assert.NoError(t, err)

	// Check that we have a package:
	assert.Contains(t, code, "package api")

	// Check that the client method signatures return response structs:
	assert.Contains(t, code, "func (c *Client) FindPetById(ctx context.Context, id int64) (*http.Response, error) {")

	// Check that the property comments were generated
	assert.Contains(t, code, "// Unique id of the pet")

	// Check that the summary comment contains newlines
	assert.Contains(t, code, `// Deletes a pet by ID
	// (DELETE /pets/{id})
`)

	// Make sure the generated code is valid:
	linter := new(lint.Linter)
	problems, err := linter.Lint("test.gen.go", []byte(code))
	assert.NoError(t, err)
	assert.Len(t, problems, 0)
}

func TestExamplePetStoreParseFunction(t *testing.T) {

	bodyBytes := []byte(`{"id": 5, "name": "testpet", "tag": "cat"}`)

	cannedResponse := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(bodyBytes)),
		Header:     http.Header{},
	}
	cannedResponse.Header.Add("Content-type", "application/json")

	findPetByIDResponse, err := examplePetstoreClient.ParseFindPetByIdResponse(cannedResponse)
	assert.NoError(t, err)
	assert.NotNil(t, findPetByIDResponse.JSON200)
	assert.Equal(t, int64(5), findPetByIDResponse.JSON200.Id)
	assert.Equal(t, "testpet", findPetByIDResponse.JSON200.Name)
	assert.NotNil(t, findPetByIDResponse.JSON200.Tag)
	assert.Equal(t, "cat", *findPetByIDResponse.JSON200.Tag)
}
func TestFilterOperationsByTag(t *testing.T) {
	packageName := "testswagger"
	t.Run("include tags", func(t *testing.T) {
		opts := Options{
			GenerateClient:     true,
			GenerateEchoServer: true,
			GenerateTypes:      true,
			EmbedSpec:          true,
			IncludeTags:        []string{"hippo", "giraffe", "cat"},
		}

		// Get a spec from the test definition in this file:
		swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(testOpenAPIDefinition))
		assert.NoError(t, err)

		// Run our code generation:
		code, err := Generate(swagger, packageName, opts)
		assert.NoError(t, err)
		assert.NotEmpty(t, code)
		assert.NotContains(t, code, `"/test/:name"`)
		assert.Contains(t, code, `"/cat"`)
	})

	t.Run("exclude tags", func(t *testing.T) {
		opts := Options{
			GenerateClient:     true,
			GenerateEchoServer: true,
			GenerateTypes:      true,
			EmbedSpec:          true,
			ExcludeTags:        []string{"hippo", "giraffe", "cat"},
		}

		// Get a spec from the test definition in this file:
		swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData([]byte(testOpenAPIDefinition))
		assert.NoError(t, err)

		// Run our code generation:
		code, err := Generate(swagger, packageName, opts)
		assert.NoError(t, err)
		assert.NotEmpty(t, code)
		assert.Contains(t, code, `"/test/:name"`)
		assert.NotContains(t, code, `"/cat"`)
	})
}

var update = flag.Bool("update", false, "update .golden files")

//func TestExampleOpenAPICodeGeneration(t *testing.T) {
//
//	testCases := map[string]struct {
//		Spec string
//		Name string
//	}{
//		"standard test": {
//			Spec: "testdata/testOpenAPIDefinition.yaml",
//			Name: "testOpenAPIDefinition",
//		},
//		"discriminated oneOf test with enums": {
//			Spec: "testdata/testOpenAPIDefinitionWithOneOfDiscriminatorsAndEnums.yaml",
//			Name: "testOpenAPIDefinitionWithOneOfDiscriminatorsAndEnums",
//		},
//		"discriminated oneOf test with enums and mappings": {
//			Spec: "testdata/pets.yaml",
//			Name: "pets",
//		},
//	}
//	for testname, test := range testCases {
//		t.Run(testname, func(t *testing.T) {
//			// Input vars for code generation:
//			packageName := "testswagger"
//			opts := Options{
//				GenerateClient: true,
//				GenerateServer: true,
//				GenerateTypes:  true,
//				EmbedSpec:      true,
//			}
//			// load spec from testdata identified by spec
//			bytes, err := ioutil.ReadFile(test.Spec)
//			if err != nil {
//				t.Fatal(err)
//			}
//			// Get a spec from the test definition in this file:
//			swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(bytes)
//			assert.NoError(t, err)
//
//			// Run our code generation:
//			code, err := Generate(swagger, packageName, opts)
//			codeBytes := []byte(code)
//			assert.NoError(t, err)
//			assert.NotEmpty(t, code)
//
//			// Make sure the code is formattable
//			_, err = format.Source([]byte(code))
//			assert.NoError(t, err)
//
//			// Make sure the generated code is valid:
//			linter := new(lint.Linter)
//			problems, err := linter.Lint("test.gen.go", codeBytes)
//			assert.NoError(t, err)
//			assert.Len(t, problems, 0)
//
//			// if update flag is set, write to golden file
//			golden := filepath.Join("testdata/", test.Name+".golden")
//			if *update {
//				ioutil.WriteFile(golden, codeBytes, 0644)
//			}
//
//			// load the golden file and run formatting to ensure test does not fail due to different go format version
//			expected, _ := ioutil.ReadFile(golden)
//			expected, err = format.Source(expected)
//			assert.NoError(t, err)
//
//			// Compare generated code with golden file contents
//			assert.Equal(t, codeBytes, expected)
//		})
//	}
//}

func TestOneOfCodeGenerationErrors(t *testing.T) {

	testCases := map[string]struct {
		Spec           string
		ContentAsserts func(t *testing.T, code string, err error)
	}{
		"failed oneOf - anonymous nested schema": {
			Spec: "testdata/failedOneOfAnonymousSchema.yaml",
			ContentAsserts: func(t *testing.T, code string, err error) {
				assert.Empty(t, code)
				// Check that we have valid (formattable) code:
				//
				// Cannot instance an interface
				assert.NotNil(t, err, "expecting error")
				assert.Contains(t, err.Error(), `'Cat.echoChamberOneOf.sound' defines a oneOf property inside an anonymous schema`)
			},
		},
		"failed oneOf - missing discriminator": {
			Spec: "testdata/failedOneOfMissingDiscriminator.yaml",
			ContentAsserts: func(t *testing.T, code string, err error) {
				assert.Empty(t, code)

				assert.NotNil(t, err, "expecting error")
				assert.Contains(t, err.Error(), `error processing oneOf`)
				assert.Contains(t, err.Error(), `Schema 'CatAlive' does not have discriminator property`)
			},
		},
	}
	for testname, test := range testCases {
		t.Run(testname, func(t *testing.T) {
			// Input vars for code generation:
			packageName := "testswagger"
			opts := Options{
				GenerateClient: true,
				GenerateServer: true,
				GenerateTypes:  true,
				EmbedSpec:      true,
			}
			// load spec from testdata identified by spec
			bytes, err := ioutil.ReadFile(test.Spec)
			if err != nil {
				t.Fatal(err)
			}
			// Get a spec from the test definition in this file:
			swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(bytes)
			assert.NoError(t, err)

			// Run our code generation:
			code, err := Generate(swagger, packageName, opts)
			assert.Error(t, err)
			test.ContentAsserts(t, code, err)
		})
	}
}

func TestExampleOpenAPICodeGenerationOfDefaults(t *testing.T) {
	testSpec := "testdata/pets.yaml"
	_, code, err := loadSwaggerFromFileAndGenerateCode(t, testSpec, Options{GenerateTypes: true})
	assert.NoError(t, err)

	// ensure default values are stored as struct tags in generated structs
	assert.Regexp(t, "Type.*\\*string.*`.*default:\"black\"`", code)
	assert.Regexp(t, "Halls.*\\*int.*`.*default:\"7\"`", code)
	assert.Regexp(t, "Towers.*\\*int.*`.*default:\"1\"`", code)
	assert.NotRegexp(t, "Type.*\\*array.*`.*default:\".*\"`", code)
}

func loadSwaggerFromFileAndGenerateCode(t *testing.T, testSpec string, opts Options) (*openapi3.Swagger, string, error) {
	// Input vars for code generation:
	packageName := "testswagger"

	// Get a spec from the test definition in this file:
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile(testSpec)
	if err != nil {
		return nil, "", err
	}
	// Run our code generation:
	code, err := Generate(swagger, packageName, opts)
	if err != nil {
		return nil, "", err
	}
	if code == "" {
		return nil, "", errors.New("empty code generated")
	}
	// Check that response structs contains fallbacks to interface for invalid types:
	// Here an invalid array with no items.
	assert.Contains(t, code, `
type getTestByNameResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]Test
	XML200       *[]Test
	JSON422      *[]interface{}
	XML422       *[]interface{}
	JSONDefault  *Error
}`)

	// Check that the helper methods are generated correctly:
	assert.Contains(t, code, "func (r getTestByNameResponse) Status() string {")
	assert.Contains(t, code, "func (r getTestByNameResponse) StatusCode() int {")
	assert.Contains(t, code, "func ParseGetTestByNameResponse(rsp *http.Response) (*getTestByNameResponse, error) {")

	// Check the client method signatures:
	assert.Contains(t, code, "type GetTestByNameParams struct {")
	assert.Contains(t, code, "Top *int `json:\"$top,omitempty\"`")
	assert.Contains(t, code, "func (c *Client) GetTestByName(ctx context.Context, name string, params *GetTestByNameParams) (*http.Response, error) {")
	assert.Contains(t, code, "func (c *ClientWithResponses) GetTestByNameWithResponse(ctx context.Context, name string, params *GetTestByNameParams) (*getTestByNameResponse, error) {")

	// Make sure the code is formattable
	_, err = format.Source([]byte(code))
	if err != nil {
		return nil, "", err
	}
	// Make sure the generated code is valid:
	linter := new(lint.Linter)
	problems, err := linter.Lint("test.gen.go", []byte(code))
	if err != nil {
		return nil, "", err
	}
	if len(problems) > 0 {
		return nil, "", errors.New("linting problems found")
	}
	return swagger, code, nil
}

func createTestServer(handler http.Handler) *httptest.Server {
	ts := httptest.NewUnstartedServer(handler)
	l, _ := net.Listen("tcp", "localhost:23456")
	ts.Listener.Close()
	ts.Listener = l
	return ts
}

func startTestServer(system http.FileSystem) func() {
	fs := http.FileServer(system)
	ts := createTestServer(fs)
	ts.Start()
	return ts.Close
}

// TestResolvedSchema tests that a schema containing remote references is resolved correctly.
// To do so, it starts an HTTP server serving files; the schema under test here contains a reference
// to a file located at the server we just started. This requires that the host address be known here
// at the test site, but also be coded  in the spec itself (see testResolvingSpec.yaml)
// Additionally, in order to ensure that the test is generating the proper spec, the generated code is compiled
// and executed and the resulting document is then compared with a golden file (which was inspected manually)
func TestResolvedSchema(t *testing.T) {

	cs := startTestServer(http.Dir("testdata"))
	defer cs()

	const mainStr = "\nfunc main(){ s, _ := GetSwaggerSpec(); fmt.Println(s)}"
	const testDir = "./testdata"
	const packageName = "main"
	testSpec := "testResolvingSpec.yaml"
	testSpecPath := filepath.Join(testDir, testSpec)

	// TODO:  these tests duplicate code that is in main() - we should refactor main at some point
	opts := Options{
		EmbedSpec: true,
	}
	// load spec from testdata identified by spec
	specBytes, err := ioutil.ReadFile(testSpecPath)
	require.NoError(t, err)

	// Get a spec from the test definition in this file:
	swagger, err := openapi3.NewSwaggerLoader(
		openapi3.WithAllowExternalRefs(true),
		openapi3.WithClearResolvedRefs(true)).LoadSwaggerFromData(specBytes)
	require.NoError(t, err)

	// Run our code generation:
	code, err := Generate(swagger, packageName, opts)
	require.NoError(t, err)
	require.NotEmpty(t, code)

	// append our main driver so that we can get the swagger string and dump it to stdout, and save the `program`
	code += mainStr
	tmpDir, err := ioutil.TempDir(testDir, "resolve")
	require.NoError(t, err)

	// defer clean up test dir/file
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, testSpec+".go")
	err = ioutil.WriteFile(testFile, []byte(code), 0644)
	require.NoError(t, err)

	// user gomega to build a temporary executable representing the spec
	testPath, _ := filepath.Abs(testFile)
	testPath = filepath.Dir(testPath)
	testBin, err := gexec.Build(testFile)
	require.NoError(t, err)
	defer gexec.CleanupBuildArtifacts()

	// run the executable
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	command := exec.Command(testBin)
	command.Stdout = stdOut
	command.Stderr = stdErr
	require.NoError(t, command.Start())
	require.NoError(t, command.Wait())

	ib := &bytes.Buffer{}
	require.NoError(t, json.Indent(ib, stdOut.Bytes(), "", "  "))

	golden := strings.TrimSuffix(testSpecPath, ".yaml") + ".json.golden"
	if *update {
		ioutil.WriteFile(golden, ib.Bytes(), 0644)
	}

	// load the golden file and run formatting to ensure test does not fail due to different go format version
	expected, _ := ioutil.ReadFile(golden)

	// Compare generated code with golden file contents
	require.Equal(t, ib.Bytes(), expected)

}
