package openapi

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

const (
	// extPropGoType overrides the generated type definition.
	extPropGoType = "x-go-type"
	// extPropGoTypeSkipOptionalPointer specifies that optional fields should
	// be the type itself instead of a pointer to the type.
	extPropGoTypeSkipOptionalPointer = "x-go-type-skip-optional-pointer"
	// extPropGoImport specifies the module to import which provides above type
	extPropGoImport = "x-go-type-import"
	// extGoName is used to override a field name
	extGoName = "x-go-name"
	// extGoTypeName is used to override a generated typename for something.
	extGoTypeName        = "x-go-type-name"
	extPropGoJsonIgnore  = "x-go-json-ignore"
	extPropOmitEmpty     = "x-omitempty"
	extPropExtraTags     = "x-oapi-codegen-extra-tags"
	extEnumVarNames      = "x-enum-varnames"
	extEnumNames         = "x-enumNames"
	extDeprecationReason = "x-deprecated-reason"
)

func extString(extPropValue *yaml.Node) (string, error) {
	if extPropValue.Kind != yaml.ScalarNode || extPropValue.Tag != "!!str" {
		return "", fmt.Errorf("expected scalar node with tag !!str, got %T", extPropValue)
	}
	var str string
	err := extPropValue.Decode(&str)
	if err != nil {
		return "", fmt.Errorf("failed to convert type: %T: %w", extPropValue, err)
	}
	return str, nil
}

func extTypeName(extPropValue *yaml.Node) (string, error) {
	return extString(extPropValue)
}

func extParsePropGoTypeSkipOptionalPointer(extPropValue *yaml.Node) (bool, error) {
	if extPropValue.Kind != yaml.ScalarNode || extPropValue.Tag != "!!bool" {
		return false, fmt.Errorf("expected scalar node with tag !!str, got %T", extPropValue)
	}
	var goTypeSkipOptionalPointer bool
	err := extPropValue.Decode(&goTypeSkipOptionalPointer)
	if err != nil {
		return false, fmt.Errorf("failed to convert type: %T: %w", extPropValue, err)
	}
	return goTypeSkipOptionalPointer, nil
}

func extParseGoFieldName(extPropValue *yaml.Node) (string, error) {
	return extString(extPropValue)
}

func extParseOmitEmpty(extPropValue *yaml.Node) (bool, error) {
	var omitEmpty bool
	err := extPropValue.Decode(&extPropValue)
	if err != nil {
		return false, fmt.Errorf("failed to convert type: %T: %w", extPropValue, err)
	}
	return omitEmpty, nil
}

func extExtraTags(extPropValue *yaml.Node) (map[string]string, error) {
	var tags map[string]string
	err := extPropValue.Decode(&tags)
	if err != nil {
		return nil, fmt.Errorf("failed to convert type: %T: %w", extPropValue, err)
	}
	return tags, nil
}

func extParseGoJsonIgnore(extPropValue *yaml.Node) (bool, error) {
	var goJsonIgnore bool
	err := extPropValue.Decode(&goJsonIgnore)
	if err != nil {
		return false, fmt.Errorf("failed to convert type: %T: %w", extPropValue, err)
	}
	return goJsonIgnore, nil
}

func extParseEnumVarNames(extPropValue []*yaml.Node) ([]string, error) {
	names := make([]string, len(extPropValue))
	for i, v := range extPropValue {
		names[i] = v.Value
	}
	return names, nil
}

func extParseDeprecationReason(extPropValue *yaml.Node) (string, error) {
	return extString(extPropValue)
}
