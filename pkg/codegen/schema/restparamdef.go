package schema

import (
	"fmt"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
)

type ParameterDefinition struct {
	ParamName string // The original json parameter name, eg param_name
	In        string // Where the parameter is defined - path, header, cookie, query
	Required  bool   // Is this a required parameter?
	Spec      *openapi3.Parameter
	Schema    Schema
}

// TypeDef is here as an adapter after a large refactoring so that I don't
// have to update all the templates. It returns the type definition for a parameter,
// without the leading '*' for optional ones.
func (pd ParameterDefinition) TypeDef() string {
	typeDecl := pd.Schema.TypeDecl()
	return typeDecl
}

// JsonTag generates the JSON annotation to map GoType to json type name. If Parameter
// Foo is marshaled to json as "foo", this will create the annotation
// 'json:"foo"'
func (pd *ParameterDefinition) JsonTag() string {
	if pd.Required {
		return fmt.Sprintf("`json:\"%s\"`", pd.ParamName)
	} else {
		return fmt.Sprintf("`json:\"%s,omitempty\"`", pd.ParamName)
	}
}

func (pd *ParameterDefinition) IsJson() bool {
	p := pd.Spec
	if len(p.Content) == 1 {
		for k := range p.Content {
			if util.IsMediaTypeJson(k) {
				return true
			}
		}
	}
	return false
}

func (pd *ParameterDefinition) IsPassThrough() bool {
	p := pd.Spec
	if len(p.Content) > 1 {
		return true
	}
	if len(p.Content) == 1 {
		return !pd.IsJson()
	}
	return false
}

func (pd *ParameterDefinition) IsStyled() bool {
	p := pd.Spec
	return p.Schema != nil
}

func (pd *ParameterDefinition) Style() string {
	style := pd.Spec.Style
	if style == "" {
		in := pd.Spec.In
		switch in {
		case "path", "header":
			return "simple"
		case "query", "cookie":
			return "form"
		default:
			panic("unknown parameter format")
		}
	}
	return style
}

func (pd *ParameterDefinition) Explode() bool {
	if pd.Spec.Explode == nil {
		in := pd.Spec.In
		switch in {
		case "path", "header":
			return false
		case "query", "cookie":
			return true
		default:
			panic("unknown parameter format")
		}
	}
	return *pd.Spec.Explode
}

func (pd ParameterDefinition) GoVariableName() string {
	name := LowercaseFirstCharacters(pd.GoName())
	if IsGoKeyword(name) {
		name = "p" + UppercaseFirstCharacter(name)
	}
	if unicode.IsNumber([]rune(name)[0]) {
		name = "n" + name
	}
	return name
}

func (pd ParameterDefinition) GoName() string {
	goName := pd.ParamName
	if extension, ok := pd.Spec.Extensions[extGoName]; ok {
		if extGoFieldName, err := extParseGoFieldName(extension); err == nil {
			goName = extGoFieldName
		}
	}
	return SchemaNameToTypeName(goName)
}

func (pd ParameterDefinition) IndirectOptional() bool {
	return !pd.Required && !pd.Schema.SkipOptionalPointer
}

type ParameterDefinitions []ParameterDefinition

func (p ParameterDefinitions) FindByName(name string) *ParameterDefinition {
	for _, param := range p {
		if param.ParamName == name {
			return &param
		}
	}
	return nil
}
