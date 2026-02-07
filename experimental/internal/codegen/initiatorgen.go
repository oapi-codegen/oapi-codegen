package codegen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/oapi-codegen/oapi-codegen/experimental/internal/codegen/templates"
)

// InitiatorTemplateData is passed to initiator templates.
type InitiatorTemplateData struct {
	Prefix      string                 // "Webhook" or "Callback"
	PrefixLower string                 // "webhook" or "callback"
	Operations  []*OperationDescriptor // Operations to generate for
}

// InitiatorGenerator generates initiator (sender) code from operation descriptors.
// It is parameterized by prefix to support both webhooks and callbacks.
type InitiatorGenerator struct {
	tmpl           *template.Template
	prefix         string // "Webhook" or "Callback"
	schemaIndex    map[string]*SchemaDescriptor
	generateSimple bool
	modelsPackage  *ModelsPackage
}

// NewInitiatorGenerator creates a new initiator generator.
func NewInitiatorGenerator(prefix string, schemaIndex map[string]*SchemaDescriptor, generateSimple bool, modelsPackage *ModelsPackage) (*InitiatorGenerator, error) {
	tmpl := template.New("initiator").Funcs(templates.Funcs()).Funcs(clientFuncs(schemaIndex, modelsPackage))

	// Parse initiator templates
	for _, pt := range templates.InitiatorTemplates {
		content, err := templates.TemplateFS.ReadFile("files/" + pt.Template)
		if err != nil {
			return nil, fmt.Errorf("failed to read initiator template %s: %w", pt.Template, err)
		}
		_, err = tmpl.New(pt.Name).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse initiator template %s: %w", pt.Template, err)
		}
	}

	// Parse shared templates (param_types)
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

	return &InitiatorGenerator{
		tmpl:           tmpl,
		prefix:         prefix,
		schemaIndex:    schemaIndex,
		generateSimple: generateSimple,
		modelsPackage:  modelsPackage,
	}, nil
}

func (g *InitiatorGenerator) templateData(ops []*OperationDescriptor) InitiatorTemplateData {
	return InitiatorTemplateData{
		Prefix:      g.prefix,
		PrefixLower: strings.ToLower(g.prefix),
		Operations:  ops,
	}
}

// GenerateBase generates the base initiator types and helpers.
func (g *InitiatorGenerator) GenerateBase(ops []*OperationDescriptor) (string, error) {
	var buf bytes.Buffer
	if err := g.tmpl.ExecuteTemplate(&buf, "initiator_base", g.templateData(ops)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateInterface generates the InitiatorInterface.
func (g *InitiatorGenerator) GenerateInterface(ops []*OperationDescriptor) (string, error) {
	var buf bytes.Buffer
	if err := g.tmpl.ExecuteTemplate(&buf, "initiator_interface", g.templateData(ops)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateMethods generates the Initiator methods.
func (g *InitiatorGenerator) GenerateMethods(ops []*OperationDescriptor) (string, error) {
	var buf bytes.Buffer
	if err := g.tmpl.ExecuteTemplate(&buf, "initiator_methods", g.templateData(ops)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateRequestBuilders generates the request builder functions.
func (g *InitiatorGenerator) GenerateRequestBuilders(ops []*OperationDescriptor) (string, error) {
	var buf bytes.Buffer
	if err := g.tmpl.ExecuteTemplate(&buf, "initiator_request_builders", g.templateData(ops)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateSimple generates the SimpleInitiator with typed responses.
func (g *InitiatorGenerator) GenerateSimple(ops []*OperationDescriptor) (string, error) {
	var buf bytes.Buffer
	if err := g.tmpl.ExecuteTemplate(&buf, "initiator_simple", g.templateData(ops)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateParamTypes generates the parameter struct types.
func (g *InitiatorGenerator) GenerateParamTypes(ops []*OperationDescriptor) (string, error) {
	var buf bytes.Buffer
	if err := g.tmpl.ExecuteTemplate(&buf, "param_types", ops); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateRequestBodyTypes generates type aliases for request bodies.
func (g *InitiatorGenerator) GenerateRequestBodyTypes(ops []*OperationDescriptor) string {
	var buf bytes.Buffer
	pkgPrefix := g.modelsPackage.Prefix()

	for _, op := range ops {
		for _, body := range op.Bodies {
			if !body.IsJSON {
				continue
			}
			var targetType string
			if body.Schema != nil {
				if body.Schema.Ref != "" {
					if target, ok := g.schemaIndex[body.Schema.Ref]; ok {
						targetType = pkgPrefix + target.ShortName
					}
				} else if body.Schema.ShortName != "" {
					targetType = pkgPrefix + body.Schema.ShortName
				}
			}
			if targetType == "" {
				targetType = "interface{}"
			}
			buf.WriteString(fmt.Sprintf("type %s = %s\n\n", body.GoTypeName, targetType))
		}
	}

	return buf.String()
}

// GenerateInitiator generates the complete initiator code.
func (g *InitiatorGenerator) GenerateInitiator(ops []*OperationDescriptor) (string, error) {
	var buf bytes.Buffer

	// Generate request body type aliases first
	bodyTypes := g.GenerateRequestBodyTypes(ops)
	buf.WriteString(bodyTypes)

	// Generate base initiator
	base, err := g.GenerateBase(ops)
	if err != nil {
		return "", fmt.Errorf("generating base initiator: %w", err)
	}
	buf.WriteString(base)
	buf.WriteString("\n")

	// Generate interface
	iface, err := g.GenerateInterface(ops)
	if err != nil {
		return "", fmt.Errorf("generating initiator interface: %w", err)
	}
	buf.WriteString(iface)
	buf.WriteString("\n")

	// Generate param types
	paramTypes, err := g.GenerateParamTypes(ops)
	if err != nil {
		return "", fmt.Errorf("generating param types: %w", err)
	}
	buf.WriteString(paramTypes)
	buf.WriteString("\n")

	// Generate methods
	methods, err := g.GenerateMethods(ops)
	if err != nil {
		return "", fmt.Errorf("generating initiator methods: %w", err)
	}
	buf.WriteString(methods)
	buf.WriteString("\n")

	// Generate request builders
	builders, err := g.GenerateRequestBuilders(ops)
	if err != nil {
		return "", fmt.Errorf("generating request builders: %w", err)
	}
	buf.WriteString(builders)
	buf.WriteString("\n")

	// Generate simple initiator if requested
	if g.generateSimple {
		simple, err := g.GenerateSimple(ops)
		if err != nil {
			return "", fmt.Errorf("generating simple initiator: %w", err)
		}
		buf.WriteString(simple)
	}

	return buf.String(), nil
}
