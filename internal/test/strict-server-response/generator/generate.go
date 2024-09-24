package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"golang.org/x/tools/imports"
	"gopkg.in/yaml.v3"
)

func init() {
	imports.LocalPrefix = "github.com/oapi-codegen/oapi-codegen/v2"
}

type ServerType string

const (
	ChiServer     ServerType = "chi"
	EchoServer    ServerType = "echo"
	FiberServer   ServerType = "fiber"
	GinServer     ServerType = "gin"
	GorillaServer ServerType = "gorilla"
	IrisServer    ServerType = "iris"
	StdHttpServer ServerType = "std-http"
)

var AllServers []ServerType = []ServerType{
	ChiServer,
	EchoServer,
	FiberServer,
	GinServer,
	GorillaServer,
	IrisServer,
	StdHttpServer,
}

func main() {
	fTestGos := make(map[ServerType]*bytes.Buffer)

	for _, server := range AllServers {
		if err := os.MkdirAll(string(server)+"/pkg1", 0o755); err != nil {
			panic(err)
		}
		if f, err := os.Create(string(server) + "/pkg1/config.yaml"); err != nil {
			panic(err)
		} else {
			fmt.Fprintf(f, `# Code generated by generator/generate.go DO NOT EDIT.
package: pkg1
generate:
  models: true
  client: true
  %s-server: true
  strict-server: true
output: %s/pkg1/pkg1.gen.go
import-mapping:
  pkg2.yaml: github.com/oapi-codegen/oapi-codegen/v2/internal/test/strict-server-response/%s/pkg2
`, server, server, server)
		}

		if err := os.MkdirAll(string(server)+"/pkg2", 0o755); err != nil {
			panic(err)
		}
		if f, err := os.Create(string(server) + "/pkg2/config.yaml"); err != nil {
			panic(err)
		} else {
			fmt.Fprintf(f, `# Code generated by generator/generate.go DO NOT EDIT.
package: pkg2
generate:
  models: true
  client: true
  %s-server: true
  strict-server: true
output-options:
  skip-prune: true
output: %s/pkg2/pkg2.gen.go
`, server, server)
		}

		fTestGos[server] = &bytes.Buffer{}
		fmt.Fprintf(fTestGos[server], `// Code generated by generator/generate.go DO NOT EDIT.

			package pkg1_test

			import (
				"github.com/gin-gonic/gin"
				"github.com/gofiber/fiber/v2"
				"github.com/gofiber/fiber/v2/middleware/adaptor"
				"github.com/kataras/iris/v12"
				"github.com/labstack/echo/v4"
				"github.com/stretchr/testify/assert"

				"github.com/oapi-codegen/oapi-codegen/v2/internal/test/strict-server-response/%s/pkg1"
				"github.com/oapi-codegen/oapi-codegen/v2/internal/test/strict-server-response/%s/pkg2"
			)

			type strictServerInterface struct{}

			`, server, server)
	}

	paths := map[string]any{}
	responses := map[int]map[string]any{
		1: {},
		2: {},
	}

	for _, ref := range []bool{false, true} {
		for _, extRef := range []bool{false, true} {
			if extRef {
				// Issue #1405
				continue
			}
			for _, header := range []bool{false, true} {
				for _, fixedStatusCode := range []bool{false, true} {
					for _, content := range []struct {
						name    []string
						content string
						tag     string
					}{
						{name: []string{"JSON"}, content: "application/json", tag: "JSON"},
						{name: []string{"Special", "JSON"}, content: "application/test+json", tag: "ApplicationTestPlusJSON"},
						{name: []string{"Formdata"}, content: "application/x-www-form-urlencoded", tag: "Formdata"},
						{name: []string{"Multipart"}, content: "multipart/form-data", tag: "Multipart"},
						{name: []string{"Multipart", "Related"}, content: "multipart/related", tag: "Multipart"},
						// Issue #1403
						// {name: []string{"Wildcard", "Multipart"}, content: "multipart/*", tag: "Multipart"},
						{name: []string{"Text"}, content: "text/plain", tag: "Text"},
						{name: []string{"Other"}, content: "application/test", tag: "Applicationtest"},
						{name: []string{"Wildcard"}, content: "application/*", tag: "Application"},
						{name: []string{"NoContent"}},
					} {
						if content.content == "text/plain" && (header || !fixedStatusCode || ref) {
							// Issue #1401
							continue
						}

						if content.content == "text/plain" {
							// Issue #1406
							continue
						}

						if content.content == "application/x-www-form-urlencoded" && ref {
							// Issue #1402
							continue
						}

						if strings.Contains(content.content, "json") && extRef && !fixedStatusCode {
							// Issue #1202
							continue
						}

						if header && content.content == "" {
							// Issue #1407
							continue
						}

						generateOneTest(fTestGos, paths, responses, ref, extRef, header, fixedStatusCode, content)
					}
				}
			}
		}
	}

	specs := map[int]map[string]any{
		1: {
			"openapi": "3.0.1",
			"components": map[string]any{
				"schemas": map[string]any{
					"TestSchema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"field1": map[string]any{
								"type": "string",
							},
							"field2": map[string]any{
								"type": "integer",
							},
						},
						"required": []any{
							"field1",
							"field2",
						},
					},
				},
				"responses": responses[1],
			},
			"paths": paths,
		},
		2: {
			"openapi": "3.0.1",
			"components": map[string]any{
				"schemas": map[string]any{
					"TestSchema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"field1": map[string]any{
								"type": "string",
							},
							"field2": map[string]any{
								"type": "integer",
							},
						},
						"required": []any{
							"field1",
							"field2",
						},
					},
				},
				"responses": responses[2],
			},
		},
	}

	for pkgId, spec := range specs {
		if buf, err := yaml.Marshal(spec); err != nil {
			panic(err)
		} else if fYAML, err := os.Create(fmt.Sprintf("pkg%d.yaml", pkgId)); err != nil {
			panic(err)
		} else {
			defer func() {
				if err := fYAML.Close(); err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			}()
			fmt.Fprintln(fYAML, "# Code generated by generator/generate.go DO NOT EDIT.")
			if _, err := fYAML.Write(buf); err != nil {
				panic(err)
			}
		}
	}

	for _, server := range AllServers {
		testGoName := string(server) + "/pkg1/pkg1_test.go"
		if buf, err := imports.Process(testGoName, fTestGos[server].Bytes(), nil); err != nil {
			panic(err)
		} else if fTestGoOut, err := os.Create(testGoName); err != nil {
			panic(err)
		} else {
			defer func() {
				if err := fTestGoOut.Close(); err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			}()
			if _, err := fTestGoOut.Write(buf); err != nil {
				panic(err)
			}
		}
	}
}

func generateOneTest(fTestGos map[ServerType]*bytes.Buffer, paths map[string]any, responses map[int]map[string]any,
	ref bool, extRef bool, header bool, fixedStatusCode bool, content struct {
		name    []string
		content string
		tag     string
	}) {
	nameSlice := []string{"test"}
	var pkgId int = 1
	if ref {
		if extRef {
			nameSlice = append(nameSlice, "Ext")
			pkgId = 2
		} else {
			nameSlice = append(nameSlice, "Ref")
		}
	} else {
		if extRef {
			return
		}
	}
	if header {
		nameSlice = append(nameSlice, "Header")
	}
	if fixedStatusCode {
		nameSlice = append(nameSlice, "Fixed")
	}
	nameSlice = append(nameSlice, content.name...)

	response := map[string]any{}

	if content.content != "" {
		response["content"] = map[string]any{
			content.content: map[string]any{
				"schema": map[string]any{
					"$ref": "#/components/schemas/TestSchema",
				},
			},
		}
	}

	if header {
		response["headers"] = map[string]any{
			"header1": map[string]any{
				"schema": map[string]any{
					"type": "string",
				},
			},
			"header2": map[string]any{
				"schema": map[string]any{
					"type": "integer",
				},
				"required": true,
			},
		}
	}

	var statusCode string
	if fixedStatusCode {
		if content.content == "" {
			statusCode = "204"
		} else {
			statusCode = "200"
		}
	} else {
		statusCode = "default"
	}

	var responseInResponses map[string]any
	if ref {
		responseName := "testResp" + strings.Join(nameSlice[1:], "")
		responses[pkgId][responseName] = response
		var ref string
		if pkgId == 1 {
			ref = fmt.Sprintf("#/components/responses/%s", responseName)
		} else {
			ref = fmt.Sprintf("pkg%d.yaml#/components/responses/%s", pkgId, responseName)
		}
		responseInResponses = map[string]any{
			"$ref": ref,
		}
	} else {
		responseInResponses = response
	}

	pathSlice := make([]string, 0, len(nameSlice))
	for _, s := range nameSlice {
		pathSlice = append(pathSlice, strings.ToLower(s))
	}
	path := "/" + strings.Join(pathSlice, "-")
	paths[path] = map[string]any{
		"get": map[string]any{
			"operationId": strings.Join(nameSlice, ""),
			"responses": map[string]any{
				statusCode: responseInResponses,
			},
		},
	}

	method := "Test" + strings.Join(nameSlice[1:], "")
	body := ""
	var statusCodeRes int
	if content.content == "" {
		statusCodeRes = 204
	} else {
		statusCodeRes = 200
	}
	switch {
	case content.content == "":
	case content.content == "text/plain":
		body = "\"bar\""
	case strings.HasPrefix(content.content, "multipart/"):
		body = fmt.Sprintf(`func(writer *multipart.Writer) error {
			if p, err := writer.CreatePart(textproto.MIMEHeader{"Content-Type": []string{"application/json"}}); err != nil {
				return err
			} else {
				return json.NewEncoder(p).Encode(pkg%d.TestSchema{
					Field1: "bar",
					Field2: 456,
				})
			}
}`, pkgId)
	case content.content == "application/test" || content.content == "application/*":
		body = "bytes.NewReader(buf)"
	default:
		body = fmt.Sprintf(`pkg%d.TestSchema{
			Field1: "bar",
			Field2: 456,
}`, pkgId)
	}
	contentRes := content.content
	if strings.HasSuffix(contentRes, "/*") {
		contentRes = strings.TrimSuffix(contentRes, "/*") + "/baz"
	}

	for _, server := range AllServers {
		fTestGo := fTestGos[server]
		fmt.Fprintf(fTestGo, "func (s strictServerInterface) %s(ctx context.Context, request pkg1.%sRequestObject) (pkg1.%sResponseObject, error) {\n", method, method, method)
		switch content.content {
		case "application/test", "application/*":
			fmt.Fprintf(fTestGo, "buf := []byte(\"bar\")\n")
		}
		var resRet string
		if ref && fixedStatusCode {
			resRet = fmt.Sprintf("pkg%d.TestResp%s%sResponse", pkgId, strings.Join(nameSlice[1:], ""), content.tag)
		} else {
			resRet = fmt.Sprintf("pkg1.%s%s%sResponse", method, statusCode, content.tag)
		}

		if header || !fixedStatusCode ||
			content.content == "application/test" || content.content == "application/*" ||
			content.content == "multipart/*" || content.content == "" {
			resRet += " {\n"
			if body != "" {
				resRet += fmt.Sprintf("Body: %s,\n", body)
			}
			if header {
				var headerType string
				if ref {
					headerType = fmt.Sprintf("pkg%d.TestResp%sResponseHeaders", pkgId, strings.Join(nameSlice[1:], ""))
				} else {
					headerType = fmt.Sprintf("pkg1.%s%sResponseHeaders", method, statusCode)
				}
				resRet += fmt.Sprintf(`Headers: %s {
					Header2: 123,
				},
				`, headerType)
			}
			if !fixedStatusCode {
				resRet += fmt.Sprintf("StatusCode: %d,\n", statusCodeRes)
			}
			if strings.HasSuffix(content.content, "/*") {
				resRet += fmt.Sprintf("ContentType: \"%s\",\n", contentRes)
			}
			switch content.content {
			case "application/test", "application/*":
				resRet += "ContentLength: int64(len(buf)),\n"
			}
			resRet += "}"
		} else {
			resRet += "(" + body + ")"
		}
		if ref && fixedStatusCode {
			if (strings.HasPrefix(content.content, "multipart/") && !header) ||
				(content.content == "" && extRef) {
				resRet = fmt.Sprintf("pkg1.%s%s%sResponse(%s)", method, statusCode, content.tag, resRet)
			} else if content.content != "" {
				resRet = fmt.Sprintf("pkg1.%s%s%sResponse{%s}", method, statusCode, content.tag, resRet)
			}
		}
		fmt.Fprintf(fTestGo, "return %s, nil\n", resRet)
		fmt.Fprintf(fTestGo, "}\n")
		fmt.Fprintf(fTestGo, "\n")

		var handlerSetup string
		var handlerCall string
		switch server {
		case ChiServer:
			handlerSetup = `hh := pkg1.Handler(pkg1.NewStrictHandler(strictServerInterface{}, nil))`
			handlerCall = `hh.ServeHTTP(w, r)`
		case EchoServer:
			handlerSetup = `e := echo.New()
				pkg1.RegisterHandlers(e, pkg1.NewStrictHandler(strictServerInterface{}, nil))`
			handlerCall = `e.Server.Handler.ServeHTTP(w, r)`
		case FiberServer:
			handlerSetup = `app := fiber.New()
				pkg1.RegisterHandlers(app, pkg1.NewStrictHandler(strictServerInterface{}, nil))
				hhf := adaptor.FiberApp(app)`
			handlerCall = `hhf(w, r)`
		case GinServer:
			handlerSetup = `g := gin.New()
				pkg1.RegisterHandlers(g, pkg1.NewStrictHandler(strictServerInterface{}, nil))`
			handlerCall = `g.Handler().ServeHTTP(w, r)`
		case GorillaServer:
			handlerSetup = `hh := pkg1.Handler(pkg1.NewStrictHandler(strictServerInterface{}, nil))`
			handlerCall = `hh.ServeHTTP(w, r)`
		case IrisServer:
			handlerSetup = `app := iris.New()
				pkg1.RegisterHandlers(app, pkg1.NewStrictHandler(strictServerInterface{}, nil))`
			handlerCall = `app.ServeHTTP(w, r)`
		case StdHttpServer:
			handlerSetup = `hh := pkg1.Handler(pkg1.NewStrictHandler(strictServerInterface{}, nil))`
			handlerCall = `hh.ServeHTTP(w, r)`
		}

		fmt.Fprintf(fTestGo, `func %s(t *testing.T) {
			%s

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !assert.Equal(t, "%s", r.URL.Path) {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				%s
			}))
			defer ts.Close()

			c, err := pkg1.NewClientWithResponses(ts.URL)
			assert.NoError(t, err)
			res, err := c.%sWithResponse(context.TODO())
			assert.NoError(t, err)
			assert.Equal(t, %d, res.StatusCode())
		`, method, handlerSetup, path, handlerCall, method, statusCodeRes)
		if header {
			// Issue #1301
			fmt.Fprintf(fTestGo, "// assert.Empty(t, res.HTTPResponse.Header.Values(\"header1\"))\n")
			fmt.Fprintf(fTestGo, "assert.Equal(t, []string{\"123\"}, res.HTTPResponse.Header.Values(\"header2\"))\n")
		}
		if !strings.HasPrefix(contentRes, "multipart/") {
			if (strings.Contains(content.content, "json") && (server == FiberServer || server == IrisServer) /* Issue #1408 */) ||
				(content.content == "" && server == FiberServer /* Issue #1409 */) {
				fmt.Fprint(fTestGo, "// ")
			}
			fmt.Fprintf(fTestGo, "assert.Equal(t, \"%s\", res.HTTPResponse.Header.Get(\"Content-Type\"))\n", contentRes)
		}
		switch {
		case content.content == "":
			fmt.Fprintf(fTestGo, "assert.Equal(t, []byte{}, res.Body)\n")
		case strings.HasPrefix(content.content, "multipart/"):
			fmt.Fprintf(fTestGo, `mediaType, params, err := mime.ParseMediaType(res.HTTPResponse.Header.Get("Content-Type"))
				if assert.NoError(t, err) {
					assert.Equal(t, "%s", mediaType)
					assert.NotEmpty(t, params["boundary"])
					reader := multipart.NewReader(bytes.NewReader(res.Body), params["boundary"])
					jsonExist := false
					for {
						if p, err := reader.NextPart(); err == io.EOF {
							break
						} else {
							assert.NoError(t, err)
							switch p.Header.Get("Content-Type") {
							case "application/json":
								var j pkg%d.TestSchema
								err := json.NewDecoder(p).Decode(&j)
								assert.NoError(t, err)
								assert.Equal(t, pkg%d.TestSchema{
									Field1: "bar",
									Field2: 456,
								}, j)
								jsonExist = true
							default:
								assert.Fail(t, "Bad Content-Type: %%s", p.Header.Get("Content-Type"))
							}
						}
					}
					assert.True(t, jsonExist)
				}
	`, contentRes, pkgId, pkgId)
		case content.content == "application/x-www-form-urlencoded":
			fmt.Fprintf(fTestGo, `form, err := url.ParseQuery(string(res.Body))
				assert.NoError(t, err)
				assert.Equal(t, url.Values{
					"field1": []string{"bar"},
					"field2": []string{"456"},
				}, form)
	`)
		case content.content == "application/json":
			fmt.Fprintf(fTestGo, "assert.Equal(t, &pkg%d.TestSchema{\n", pkgId)
			fmt.Fprintf(fTestGo, "Field1: \"bar\",\n")
			fmt.Fprintf(fTestGo, "Field2: 456,\n")
			fmt.Fprintf(fTestGo, "}, res.JSON%s%s)\n", strings.ToUpper(statusCode[:1]), statusCode[1:])
		case content.content == "application/test+json":
			fmt.Fprintf(fTestGo, "assert.Equal(t, &pkg%d.TestSchema{\n", pkgId)
			fmt.Fprintf(fTestGo, "Field1: \"bar\",\n")
			fmt.Fprintf(fTestGo, "Field2: 456,\n")
			fmt.Fprintf(fTestGo, "}, res.ApplicationtestJSON%s%s)\n", strings.ToUpper(statusCode[:1]), statusCode[1:])
		default:
			fmt.Fprintf(fTestGo, "assert.Equal(t, []byte(\"bar\"), res.Body)\n")
		}
		fmt.Fprintf(fTestGo, "}\n")
		fmt.Fprintf(fTestGo, "\n")
	}
}
