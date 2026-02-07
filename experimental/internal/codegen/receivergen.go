package codegen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/oapi-codegen/oapi-codegen-exp/experimental/internal/codegen/templates"
)

// ReceiverTemplateData is passed to receiver templates.
type ReceiverTemplateData struct {
	Prefix      string                 // "Webhook" or "Callback"
	PrefixLower string                 // "webhook" or "callback"
	Operations  []*OperationDescriptor // Operations to generate for
}

// ReceiverGenerator generates receiver code from operation descriptors.
// It is parameterized by prefix to support both webhooks and callbacks.
type ReceiverGenerator struct {
	tmpl       *template.Template
	prefix     string // "Webhook" or "Callback"
	serverType string
}

// NewReceiverGenerator creates a new receiver generator for the specified server type.
func NewReceiverGenerator(prefix string, serverType string) (*ReceiverGenerator, error) {
	if serverType == "" {
		return nil, fmt.Errorf("%s receiver requires a server type to be set", prefix)
	}

	tmpl := template.New("receiver").Funcs(templates.Funcs())

	// Get receiver templates for the specified server type
	receiverTemplates, err := getReceiverTemplates(serverType)
	if err != nil {
		return nil, err
	}

	// Parse receiver-specific templates
	for _, ct := range receiverTemplates {
		content, err := templates.TemplateFS.ReadFile("files/" + ct.Template)
		if err != nil {
			return nil, fmt.Errorf("failed to read receiver template %s: %w", ct.Template, err)
		}
		_, err = tmpl.New(ct.Name).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse receiver template %s: %w", ct.Template, err)
		}
	}

	// Parse shared templates (errors, param_types)
	for _, st := range templates.SharedServerTemplates {
		content, err := templates.TemplateFS.ReadFile("files/" + st.Template)
		if err != nil {
			return nil, fmt.Errorf("failed to read shared template %s: %w", st.Template, err)
		}
		_, err = tmpl.New(st.Name).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse shared template %s: %w", st.Template, err)
		}
	}

	return &ReceiverGenerator{
		tmpl:       tmpl,
		prefix:     prefix,
		serverType: serverType,
	}, nil
}

// getReceiverTemplates returns the receiver templates for the specified server type.
func getReceiverTemplates(serverType string) (map[string]templates.ReceiverTemplate, error) {
	switch serverType {
	case ServerTypeStdHTTP:
		return templates.StdHTTPReceiverTemplates, nil
	case ServerTypeChi:
		return templates.ChiReceiverTemplates, nil
	case ServerTypeEcho:
		return templates.EchoReceiverTemplates, nil
	case ServerTypeEchoV4:
		return templates.EchoV4ReceiverTemplates, nil
	case ServerTypeGin:
		return templates.GinReceiverTemplates, nil
	case ServerTypeGorilla:
		return templates.GorillaReceiverTemplates, nil
	case ServerTypeFiber:
		return templates.FiberReceiverTemplates, nil
	case ServerTypeIris:
		return templates.IrisReceiverTemplates, nil
	default:
		return nil, fmt.Errorf("unsupported server type for receiver: %q", serverType)
	}
}

func (g *ReceiverGenerator) templateData(ops []*OperationDescriptor) ReceiverTemplateData {
	return ReceiverTemplateData{
		Prefix:      g.prefix,
		PrefixLower: strings.ToLower(g.prefix),
		Operations:  ops,
	}
}

// GenerateReceiver generates the receiver interface and handler functions.
func (g *ReceiverGenerator) GenerateReceiver(ops []*OperationDescriptor) (string, error) {
	var buf bytes.Buffer

	if err := g.tmpl.ExecuteTemplate(&buf, "receiver", g.templateData(ops)); err != nil {
		return "", fmt.Errorf("generating receiver code: %w", err)
	}

	return buf.String(), nil
}

// GenerateParamTypes generates the parameter struct types.
func (g *ReceiverGenerator) GenerateParamTypes(ops []*OperationDescriptor) (string, error) {
	var buf bytes.Buffer
	if err := g.tmpl.ExecuteTemplate(&buf, "param_types", ops); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateErrors generates error types (shared with server).
func (g *ReceiverGenerator) GenerateErrors() (string, error) {
	var buf bytes.Buffer
	if err := g.tmpl.ExecuteTemplate(&buf, "errors", nil); err != nil {
		return "", err
	}
	return buf.String(), nil
}
