package codegen

import (
	"sort"
	"strings"

	"github.com/oapi-codegen/oapi-codegen-exp/experimental/internal/codegen/templates"
)

// CodegenContext is a centralized tracker for imports, helpers, param functions,
// and custom type templates needed during code generation. Code at any depth
// can call its registration methods; the final output assembly queries it to
// emit exactly what was requested.
type CodegenContext struct {
	imports     map[string]string // path -> alias
	helpers     map[string]bool   // helper template names needed (e.g., "marshal_form")
	params      map[string]bool   // param style keys (e.g., "style_simple", "bind_form_explode")
	customTypes map[string]bool   // custom type template names (e.g., "nullable", "Date")
}

// NewCodegenContext creates a new CodegenContext.
func NewCodegenContext() *CodegenContext {
	return &CodegenContext{
		imports:     make(map[string]string),
		helpers:     make(map[string]bool),
		params:      make(map[string]bool),
		customTypes: make(map[string]bool),
	}
}

// --- Import registration ---

// AddImport records an import path needed by the generated code.
func (c *CodegenContext) AddImport(path string) {
	if path != "" {
		c.imports[path] = ""
	}
}

// AddImportAlias records an import path with an alias.
func (c *CodegenContext) AddImportAlias(path, alias string) {
	if path != "" {
		c.imports[path] = alias
	}
}

// AddImports adds multiple imports from a map[path]alias.
func (c *CodegenContext) AddImports(imports map[string]string) {
	for path, alias := range imports {
		c.AddImportAlias(path, alias)
	}
}

// --- Helper registration ---

// NeedHelper records that a helper template is needed (e.g., "marshal_form").
func (c *CodegenContext) NeedHelper(name string) {
	if name != "" {
		c.helpers[name] = true
	}
}

// NeedParam records that a param style/explode combination is needed.
// It records both the style (serialization) and bind (deserialization) keys.
func (c *CodegenContext) NeedParam(style string, explode bool) {
	styleKey := templates.ParamStyleKey("style_", style, explode)
	bindKey := templates.ParamStyleKey("bind_", style, explode)
	c.params[styleKey] = true
	c.params[bindKey] = true
}

// NeedCustomType records that a custom type template is needed (e.g., "nullable", "Date").
func (c *CodegenContext) NeedCustomType(name string) {
	if name != "" {
		c.customTypes[name] = true
	}
}

// --- Query methods ---

// Imports returns the collected imports as a map[path]alias.
func (c *CodegenContext) Imports() map[string]string {
	return c.imports
}

// RequiredHelpers returns the sorted list of helper template names needed.
func (c *CodegenContext) RequiredHelpers() []string {
	result := make([]string, 0, len(c.helpers))
	for name := range c.helpers {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

// RequiredParams returns the sorted list of param style keys needed.
func (c *CodegenContext) RequiredParams() []string {
	result := make([]string, 0, len(c.params))
	for key := range c.params {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

// HasAnyParams returns true if any param functions are needed.
func (c *CodegenContext) HasAnyParams() bool {
	return len(c.params) > 0
}

// RequiredCustomTypes returns the sorted list of custom type template names needed.
func (c *CodegenContext) RequiredCustomTypes() []string {
	result := make([]string, 0, len(c.customTypes))
	for name := range c.customTypes {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

// --- Convenience methods (mirror TypeGenerator's API for easy migration) ---

// AddJSONImport adds encoding/json import.
func (c *CodegenContext) AddJSONImport() {
	c.AddImport("encoding/json")
}

// AddJSONImports adds encoding/json and fmt imports.
func (c *CodegenContext) AddJSONImports() {
	c.AddImport("encoding/json")
	c.AddImport("fmt")
}

// --- Param template/import resolution ---

// GetRequiredParamTemplates returns the list of param templates needed, with
// the helpers template first if any params are used.
func (c *CodegenContext) GetRequiredParamTemplates() []templates.ParamTemplate {
	if !c.HasAnyParams() {
		return nil
	}

	var result []templates.ParamTemplate
	result = append(result, templates.ParamHelpersTemplate)

	keys := c.RequiredParams()
	for _, key := range keys {
		tmpl, ok := templates.ParamTemplates[key]
		if !ok {
			continue
		}
		result = append(result, tmpl)
	}

	return result
}

// GetRequiredParamImports returns all imports needed for used param functions.
func (c *CodegenContext) GetRequiredParamImports() []templates.Import {
	if !c.HasAnyParams() {
		return nil
	}

	importSet := make(map[string]templates.Import)

	for _, imp := range templates.ParamHelpersTemplate.Imports {
		importSet[imp.Path] = imp
	}

	for key := range c.params {
		tmpl, ok := templates.ParamTemplates[key]
		if !ok {
			continue
		}
		for _, imp := range tmpl.Imports {
			importSet[imp.Path] = imp
		}
	}

	result := make([]templates.Import, 0, len(importSet))
	for _, imp := range importSet {
		result = append(result, imp)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Path < result[j].Path
	})

	return result
}

// NeedFormHelper is a convenience method that checks operations for form-encoded
// bodies and registers the "marshal_form" helper if any are found.
func (c *CodegenContext) NeedFormHelper(ops []*OperationDescriptor) {
	for _, op := range ops {
		for _, body := range op.Bodies {
			if body.IsFormEncoded && body.GenerateTyped {
				c.NeedHelper("marshal_form")
				return
			}
		}
	}
}

// AddTemplateImports adds all imports declared by the given template import slices.
func (c *CodegenContext) AddTemplateImports(imports []templates.Import) {
	for _, imp := range imports {
		c.AddImportAlias(imp.Path, imp.Alias)
	}
}

// loadCustomType loads a custom type template and returns its code and imports.
// This is the same logic as the standalone loadCustomType function but integrated
// so that imports are registered directly on the context.
func (c *CodegenContext) loadAndRegisterCustomType(templateName string) string {
	typeName := strings.TrimSuffix(templateName, ".tmpl")

	var typeDef templates.TypeTemplate
	var found bool

	for name, def := range templates.TypeTemplates {
		if def.Template == "types/"+templateName || strings.ToLower(name) == typeName {
			typeDef = def
			found = true
			break
		}
	}

	if !found {
		return ""
	}

	content, err := templates.TemplateFS.ReadFile("files/" + typeDef.Template)
	if err != nil {
		return ""
	}

	code := string(content)
	if idx := strings.Index(code, "}}"); idx != -1 {
		code = strings.TrimLeft(code[idx+2:], "\n")
	}

	// Register imports directly on context
	for _, imp := range typeDef.Imports {
		c.AddImportAlias(imp.Path, imp.Alias)
	}

	return code
}
