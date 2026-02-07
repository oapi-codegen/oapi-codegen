package codegen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/oapi-codegen/oapi-codegen/experimental/internal/codegen/templates"
)

// ClientGenerator generates client code from operation descriptors.
type ClientGenerator struct {
	tmpl           *template.Template
	schemaIndex    map[string]*SchemaDescriptor
	generateSimple bool
	modelsPackage  *ModelsPackage
}

// NewClientGenerator creates a new client generator.
// modelsPackage can be nil if models are in the same package.
func NewClientGenerator(schemaIndex map[string]*SchemaDescriptor, generateSimple bool, modelsPackage *ModelsPackage) (*ClientGenerator, error) {
	tmpl := template.New("client").Funcs(templates.Funcs()).Funcs(clientFuncs(schemaIndex, modelsPackage))

	// Parse client templates
	for _, ct := range templates.ClientTemplates {
		content, err := templates.TemplateFS.ReadFile("files/" + ct.Template)
		if err != nil {
			return nil, err
		}
		_, err = tmpl.New(ct.Name).Parse(string(content))
		if err != nil {
			return nil, err
		}
	}

	// Parse shared templates (param_types is shared with server)
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

	return &ClientGenerator{
		tmpl:           tmpl,
		schemaIndex:    schemaIndex,
		generateSimple: generateSimple,
		modelsPackage:  modelsPackage,
	}, nil
}

// clientFuncs returns template functions specific to client generation.
func clientFuncs(schemaIndex map[string]*SchemaDescriptor, modelsPackage *ModelsPackage) template.FuncMap {
	return template.FuncMap{
		"pathFmt":                        pathFmt,
		"isSimpleOperation":              isSimpleOperation,
		"simpleOperationSuccessResponse": simpleOperationSuccessResponse,
		"errorResponseForOperation":      errorResponseForOperation,
		"goTypeForContent": func(content *ResponseContentDescriptor) string {
			return goTypeForContent(content, schemaIndex, modelsPackage)
		},
		"modelsPkg": func() string {
			return modelsPackage.Prefix()
		},
	}
}

// pathFmt converts a path with {param} placeholders to a format string.
// Example: "/pets/{petId}" -> "/pets/%s"
func pathFmt(path string) string {
	result := path
	for {
		start := strings.Index(result, "{")
		if start == -1 {
			break
		}
		end := strings.Index(result, "}")
		if end == -1 {
			break
		}
		result = result[:start] + "%s" + result[end+1:]
	}
	return result
}

// isSimpleOperation returns true if an operation has a single JSON success response type.
// "Simple" operations can have typed wrapper methods in SimpleClient.
func isSimpleOperation(op *OperationDescriptor) bool {
	// Must have responses
	if len(op.Responses) == 0 {
		return false
	}

	// Count success responses (2xx or default that could be success)
	var successResponses []*ResponseDescriptor
	for _, r := range op.Responses {
		if strings.HasPrefix(r.StatusCode, "2") {
			successResponses = append(successResponses, r)
		}
	}

	// Must have exactly one success response
	if len(successResponses) != 1 {
		return false
	}

	success := successResponses[0]

	// Must have at least one content type and exactly one JSON content type
	// (i.e., if there are multiple content types, we can't have a simple typed response)
	if len(success.Contents) == 0 {
		return false
	}
	if len(success.Contents) != 1 {
		return false
	}

	// The single content type must be JSON
	return success.Contents[0].IsJSON
}

// simpleOperationSuccessResponse returns the single success response for a simple operation.
func simpleOperationSuccessResponse(op *OperationDescriptor) *ResponseDescriptor {
	for _, r := range op.Responses {
		if strings.HasPrefix(r.StatusCode, "2") {
			return r
		}
	}
	return nil
}

// errorResponseForOperation returns the error response (default or 4xx/5xx) if one exists.
func errorResponseForOperation(op *OperationDescriptor) *ResponseDescriptor {
	// First, look for a default response
	for _, r := range op.Responses {
		if r.StatusCode == "default" {
			if len(r.Contents) > 0 && r.Contents[0].IsJSON {
				return r
			}
		}
	}
	// Then look for a 4xx or 5xx response
	for _, r := range op.Responses {
		if strings.HasPrefix(r.StatusCode, "4") || strings.HasPrefix(r.StatusCode, "5") {
			if len(r.Contents) > 0 && r.Contents[0].IsJSON {
				return r
			}
		}
	}
	return nil
}

// goTypeForContent returns the Go type for a response content descriptor.
// If modelsPackage is set, type names are prefixed with the package name.
func goTypeForContent(content *ResponseContentDescriptor, schemaIndex map[string]*SchemaDescriptor, modelsPackage *ModelsPackage) string {
	if content == nil || content.Schema == nil {
		return "interface{}"
	}

	pkgPrefix := modelsPackage.Prefix()

	// If the schema has a reference, look it up
	if content.Schema.Ref != "" {
		if target, ok := schemaIndex[content.Schema.Ref]; ok {
			return pkgPrefix + target.ShortName
		}
	}

	// Check if this is an array schema with items that have a reference
	if content.Schema.Schema != nil && content.Schema.Schema.Items != nil {
		itemProxy := content.Schema.Schema.Items.A
		if itemProxy != nil && itemProxy.IsReference() {
			ref := itemProxy.GetReference()
			if target, ok := schemaIndex[ref]; ok {
				return "[]" + pkgPrefix + target.ShortName
			}
		}
	}

	// If the schema has a short name, use it
	if content.Schema.ShortName != "" {
		return pkgPrefix + content.Schema.ShortName
	}

	// Fall back to the stable name
	if content.Schema.StableName != "" {
		return pkgPrefix + content.Schema.StableName
	}

	// Try to derive from the schema itself
	if content.Schema.Schema != nil {
		return schemaToGoType(content.Schema.Schema)
	}

	return "interface{}"
}

// GenerateBase generates the base client types and helpers.
func (g *ClientGenerator) GenerateBase() (string, error) {
	var buf bytes.Buffer
	if err := g.tmpl.ExecuteTemplate(&buf, "base", nil); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateInterface generates the ClientInterface.
func (g *ClientGenerator) GenerateInterface(ops []*OperationDescriptor) (string, error) {
	var buf bytes.Buffer
	if err := g.tmpl.ExecuteTemplate(&buf, "interface", ops); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateMethods generates the Client methods.
func (g *ClientGenerator) GenerateMethods(ops []*OperationDescriptor) (string, error) {
	var buf bytes.Buffer
	if err := g.tmpl.ExecuteTemplate(&buf, "methods", ops); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateRequestBuilders generates the request builder functions.
func (g *ClientGenerator) GenerateRequestBuilders(ops []*OperationDescriptor) (string, error) {
	var buf bytes.Buffer
	if err := g.tmpl.ExecuteTemplate(&buf, "request_builders", ops); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateSimple generates the SimpleClient with typed responses.
func (g *ClientGenerator) GenerateSimple(ops []*OperationDescriptor) (string, error) {
	var buf bytes.Buffer
	if err := g.tmpl.ExecuteTemplate(&buf, "simple", ops); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateParamTypes generates the parameter struct types.
func (g *ClientGenerator) GenerateParamTypes(ops []*OperationDescriptor) (string, error) {
	var buf bytes.Buffer
	if err := g.tmpl.ExecuteTemplate(&buf, "param_types", ops); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateRequestBodyTypes generates type aliases for request bodies.
func (g *ClientGenerator) GenerateRequestBodyTypes(ops []*OperationDescriptor) string {
	var buf bytes.Buffer
	pkgPrefix := g.modelsPackage.Prefix()

	for _, op := range ops {
		for _, body := range op.Bodies {
			if !body.IsJSON {
				continue
			}
			// Get the underlying type for this request body
			var targetType string
			if body.Schema != nil {
				if body.Schema.Ref != "" {
					// Reference to a component schema
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

			// Generate type alias: type addPetJSONRequestBody = models.NewPet
			buf.WriteString(fmt.Sprintf("type %s = %s\n\n", body.GoTypeName, targetType))
		}
	}

	return buf.String()
}

// GenerateClient generates the complete client code.
func (g *ClientGenerator) GenerateClient(ops []*OperationDescriptor) (string, error) {
	var buf bytes.Buffer

	// Generate request body type aliases first
	bodyTypes := g.GenerateRequestBodyTypes(ops)
	buf.WriteString(bodyTypes)

	// Generate base client
	base, err := g.GenerateBase()
	if err != nil {
		return "", fmt.Errorf("generating base client: %w", err)
	}
	buf.WriteString(base)
	buf.WriteString("\n")

	// Generate interface
	iface, err := g.GenerateInterface(ops)
	if err != nil {
		return "", fmt.Errorf("generating client interface: %w", err)
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
		return "", fmt.Errorf("generating client methods: %w", err)
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

	// Generate simple client if requested
	if g.generateSimple {
		simple, err := g.GenerateSimple(ops)
		if err != nil {
			return "", fmt.Errorf("generating simple client: %w", err)
		}
		buf.WriteString(simple)
	}

	return buf.String(), nil
}
