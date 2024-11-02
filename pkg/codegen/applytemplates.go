package codegen

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen/openapiv3"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen/schema"
)

// GenerateIrisServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateIrisServer(t *template.Template, operations []schema.OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"iris/iris-interface.tmpl", "iris/iris-middleware.tmpl", "iris/iris-handler.tmpl"}, t, operations)
}

// GenerateChiServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateChiServer(t *template.Template, operations []schema.OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"chi/chi-interface.tmpl", "chi/chi-middleware.tmpl", "chi/chi-handler.tmpl"}, t, operations)
}

// GenerateFiberServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateFiberServer(t *template.Template, operations []schema.OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"fiber/fiber-interface.tmpl", "fiber/fiber-middleware.tmpl", "fiber/fiber-handler.tmpl"}, t, operations)
}

// GenerateEchoServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateEchoServer(t *template.Template, operations []schema.OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"echo/echo-interface.tmpl", "echo/echo-wrappers.tmpl", "echo/echo-register.tmpl"}, t, operations)
}

// GenerateGinServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateGinServer(t *template.Template, operations []schema.OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"gin/gin-interface.tmpl", "gin/gin-wrappers.tmpl", "gin/gin-register.tmpl"}, t, operations)
}

// GenerateGorillaServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateGorillaServer(t *template.Template, operations []schema.OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"gorilla/gorilla-interface.tmpl", "gorilla/gorilla-middleware.tmpl", "gorilla/gorilla-register.tmpl"}, t, operations)
}

// GenerateStdHTTPServer generates all the go code for the ServerInterface as well as
// all the wrapper functions around our handlers.
func GenerateStdHTTPServer(t *template.Template, operations []schema.OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"stdhttp/std-http-interface.tmpl", "stdhttp/std-http-middleware.tmpl", "stdhttp/std-http-handler.tmpl"}, t, operations)
}

func GenerateStrictServer(t *template.Template, operations []schema.OperationDefinition, opts openapiv3.Configuration) (string, error) {

	var templates []string

	if opts.Generate.ChiServer || opts.Generate.GorillaServer || opts.Generate.StdHTTPServer {
		templates = append(templates, "strict/strict-interface.tmpl", "strict/strict-http.tmpl")
	}
	if opts.Generate.EchoServer {
		templates = append(templates, "strict/strict-interface.tmpl", "strict/strict-echo.tmpl")
	}
	if opts.Generate.GinServer {
		templates = append(templates, "strict/strict-interface.tmpl", "strict/strict-gin.tmpl")
	}
	if opts.Generate.FiberServer {
		templates = append(templates, "strict/strict-fiber-interface.tmpl", "strict/strict-fiber.tmpl")
	}
	if opts.Generate.IrisServer {
		templates = append(templates, "strict/strict-iris-interface.tmpl", "strict/strict-iris.tmpl")
	}

	return GenerateTemplates(templates, t, operations)
}

func GenerateStrictResponses(t *template.Template, responses []schema.ResponseDefinition) (string, error) {
	return GenerateTemplates([]string{"strict/strict-responses.tmpl"}, t, responses)
}

// GenerateClient uses the template engine to generate the function which registers our wrappers
// as Echo path handlers.
func GenerateClient(t *template.Template, ops []schema.OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"client.tmpl"}, t, ops)
}

// GenerateClientWithResponses generates a client which extends the basic client which does response
// unmarshaling.
func GenerateClientWithResponses(t *template.Template, ops []schema.OperationDefinition) (string, error) {
	return GenerateTemplates([]string{"client-with-responses.tmpl"}, t, ops)
}

// GenerateTemplates used to generate templates
func GenerateTemplates(templates []string, t *template.Template, ops interface{}) (string, error) {
	var generatedTemplates []string
	for _, tmpl := range templates {
		var buf bytes.Buffer
		w := bufio.NewWriter(&buf)

		if err := t.ExecuteTemplate(w, tmpl, ops); err != nil {
			return "", fmt.Errorf("error generating %s: %s", tmpl, err)
		}
		if err := w.Flush(); err != nil {
			return "", fmt.Errorf("error flushing output buffer for %s: %s", tmpl, err)
		}
		generatedTemplates = append(generatedTemplates, buf.String())
	}

	return strings.Join(generatedTemplates, "\n"), nil
}
