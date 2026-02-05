package codegen

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/oapi-codegen/oapi-codegen/experimental/internal/codegen/templates"
)

// ServerGenerator generates server code from operation descriptors.
type ServerGenerator struct {
	tmpl *template.Template
}

// NewServerGenerator creates a new server generator.
func NewServerGenerator() (*ServerGenerator, error) {
	tmpl := template.New("server").Funcs(templates.Funcs())

	// Parse StdHTTP templates
	for _, st := range templates.StdHTTPServerTemplates {
		content, err := templates.TemplateFS.ReadFile("files/" + st.Template)
		if err != nil {
			return nil, err
		}
		_, err = tmpl.New(st.Name).Parse(string(content))
		if err != nil {
			return nil, err
		}
	}

	// Parse shared templates
	for _, st := range templates.SharedServerTemplates {
		content, err := templates.TemplateFS.ReadFile("files/" + st.Template)
		if err != nil {
			return nil, err
		}
		_, err = tmpl.New(st.Name).Parse(string(content))
		if err != nil {
			return nil, err
		}
	}

	return &ServerGenerator{tmpl: tmpl}, nil
}

// GenerateInterface generates the ServerInterface.
func (g *ServerGenerator) GenerateInterface(ops []*OperationDescriptor) (string, error) {
	var buf bytes.Buffer
	if err := g.tmpl.ExecuteTemplate(&buf, "interface", ops); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateHandler generates the HTTP handler and routing code.
func (g *ServerGenerator) GenerateHandler(ops []*OperationDescriptor) (string, error) {
	var buf bytes.Buffer
	if err := g.tmpl.ExecuteTemplate(&buf, "handler", ops); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateWrapper generates the ServerInterfaceWrapper.
func (g *ServerGenerator) GenerateWrapper(ops []*OperationDescriptor) (string, error) {
	var buf bytes.Buffer
	if err := g.tmpl.ExecuteTemplate(&buf, "wrapper", ops); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateErrors generates the error types.
func (g *ServerGenerator) GenerateErrors() (string, error) {
	var buf bytes.Buffer
	if err := g.tmpl.ExecuteTemplate(&buf, "errors", nil); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateParamTypes generates the parameter struct types.
func (g *ServerGenerator) GenerateParamTypes(ops []*OperationDescriptor) (string, error) {
	var buf bytes.Buffer
	if err := g.tmpl.ExecuteTemplate(&buf, "param_types", ops); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateServer generates server code based on the specified server type.
// Returns empty string if serverType is empty (no server generation requested).
// Returns an error if serverType is not supported.
func (g *ServerGenerator) GenerateServer(serverType string, ops []*OperationDescriptor) (string, error) {
	switch serverType {
	case "":
		// Empty string means no server generation
		return "", nil
	case ServerTypeStdHTTP:
		return g.generateStdHTTPServer(ops)
	default:
		return "", fmt.Errorf("unsupported server type: %q (supported: %q)", serverType, ServerTypeStdHTTP)
	}
}

// generateStdHTTPServer generates the complete StdHTTP server code.
func (g *ServerGenerator) generateStdHTTPServer(ops []*OperationDescriptor) (string, error) {
	var buf bytes.Buffer

	// Generate interface
	iface, err := g.GenerateInterface(ops)
	if err != nil {
		return "", err
	}
	buf.WriteString(iface)
	buf.WriteString("\n")

	// Generate param types
	paramTypes, err := g.GenerateParamTypes(ops)
	if err != nil {
		return "", err
	}
	buf.WriteString(paramTypes)
	buf.WriteString("\n")

	// Generate wrapper
	wrapper, err := g.GenerateWrapper(ops)
	if err != nil {
		return "", err
	}
	buf.WriteString(wrapper)
	buf.WriteString("\n")

	// Generate handler
	handler, err := g.GenerateHandler(ops)
	if err != nil {
		return "", err
	}
	buf.WriteString(handler)
	buf.WriteString("\n")

	// Generate errors
	errors, err := g.GenerateErrors()
	if err != nil {
		return "", err
	}
	buf.WriteString(errors)

	return buf.String(), nil
}
