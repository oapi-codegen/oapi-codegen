package codegen

import (
	"encoding/json"
	"fmt"
)

const (
	// extPropGoType overrides the generated type definition.
	extPropGoType = "x-go-type"
	// extPropGoImport specifies the module to import which provides above type
	extPropGoImport = "x-go-type-import"
	// extGoName is used to override a field name
	extGoName = "x-go-name"
	// extGoTypeName is used to override a generated typename for something.
	extGoTypeName       = "x-go-type-name"
	extPropGoJsonIgnore = "x-go-json-ignore"
	extPropOmitEmpty    = "x-omitempty"
	extPropExtraTags    = "x-oapi-codegen-extra-tags"
)

func extString(extPropValue interface{}) (string, error) {
	raw, ok := extPropValue.(json.RawMessage)
	if !ok {
		return "", fmt.Errorf("failed to convert type: %T", extPropValue)
	}
	var str string
	if err := json.Unmarshal(raw, &str); err != nil {
		return "", fmt.Errorf("failed to unmarshal json: %w", err)
	}

	return str, nil
}
func extTypeName(extPropValue interface{}) (string, error) {
	return extString(extPropValue)
}

func extParseGoFieldName(extPropValue interface{}) (string, error) {
	return extString(extPropValue)
}

func extParseOmitEmpty(extPropValue interface{}) (bool, error) {
	raw, ok := extPropValue.(json.RawMessage)
	if !ok {
		return false, fmt.Errorf("failed to convert type: %T", extPropValue)
	}

	var omitEmpty bool
	if err := json.Unmarshal(raw, &omitEmpty); err != nil {
		return false, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	return omitEmpty, nil
}

func extExtraTags(extPropValue interface{}) (map[string]string, error) {
	raw, ok := extPropValue.(json.RawMessage)
	if !ok {
		return nil, fmt.Errorf("failed to convert type: %T", extPropValue)
	}
	var tags map[string]string
	if err := json.Unmarshal(raw, &tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}
	return tags, nil
}

func extParseGoJsonIgnore(extPropValue interface{}) (bool, error) {
	raw, ok := extPropValue.(json.RawMessage)
	if !ok {
		return false, fmt.Errorf("failed to convert type: %T", extPropValue)
	}

	var goJsonIgnore bool
	if err := json.Unmarshal(raw, &goJsonIgnore); err != nil {
		return false, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	return goJsonIgnore, nil
}
