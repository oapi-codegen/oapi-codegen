package codegen

import (
	"fmt"
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
	extPropOmitZero      = "x-omitzero"
	extPropExtraTags     = "x-oapi-codegen-extra-tags"
	extEnumVarNames      = "x-enum-varnames"
	extEnumNames         = "x-enumNames"
	extDeprecationReason = "x-deprecated-reason"
	extOrder             = "x-order"
	// extOapiCodegenOnlyHonourGoName is to be used to explicitly enforce the generation of a field as the `x-go-name` extension has describe it.
	// This is intended to be used alongside the `allow-unexported-struct-field-names` Compatibility option
	extOapiCodegenOnlyHonourGoName = "x-oapi-codegen-only-honour-go-name"
)

func extString(extPropValue interface{}) (string, error) {
	str, ok := extPropValue.(string)
	if !ok {
		return "", fmt.Errorf("failed to convert type: %T", extPropValue)
	}
	return str, nil
}

func extTypeName(extPropValue interface{}) (string, error) {
	return extString(extPropValue)
}

func extParsePropGoTypeSkipOptionalPointer(extPropValue interface{}) (bool, error) {
	goTypeSkipOptionalPointer, ok := extPropValue.(bool)
	if !ok {
		return false, fmt.Errorf("failed to convert type: %T", extPropValue)
	}
	return goTypeSkipOptionalPointer, nil
}

func extParseGoFieldName(extPropValue interface{}) (string, error) {
	return extString(extPropValue)
}

func extParseOmitEmpty(extPropValue interface{}) (bool, error) {
	omitEmpty, ok := extPropValue.(bool)
	if !ok {
		return false, fmt.Errorf("failed to convert type: %T", extPropValue)
	}
	return omitEmpty, nil
}

func extParseOmitZero(extPropValue interface{}) (bool, error) {
	omitZero, ok := extPropValue.(bool)
	if !ok {
		return false, fmt.Errorf("failed to convert type: %T", extPropValue)
	}
	return omitZero, nil
}

func extExtraTags(extPropValue interface{}) (map[string]string, error) {
	tagsI, ok := extPropValue.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to convert type: %T", extPropValue)
	}
	tags := make(map[string]string, len(tagsI))
	for k, v := range tagsI {
		vs, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("failed to convert type: %T", v)
		}
		tags[k] = vs
	}
	return tags, nil
}

func extParseGoJsonIgnore(extPropValue interface{}) (bool, error) {
	goJsonIgnore, ok := extPropValue.(bool)
	if !ok {
		return false, fmt.Errorf("failed to convert type: %T", extPropValue)
	}
	return goJsonIgnore, nil
}

func extParseEnumVarNames(extPropValue interface{}) ([]string, error) {
	namesI, ok := extPropValue.([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to convert type: %T", extPropValue)
	}
	names := make([]string, len(namesI))
	for i, v := range namesI {
		vs, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("failed to convert type: %T", v)
		}
		names[i] = vs
	}
	return names, nil
}

func extParseDeprecationReason(extPropValue interface{}) (string, error) {
	return extString(extPropValue)
}

func extParseOapiCodegenOnlyHonourGoName(extPropValue interface{}) (bool, error) {
	onlyHonourGoName, ok := extPropValue.(bool)
	if !ok {
		return false, fmt.Errorf("failed to convert type: %T", extPropValue)
	}
	return onlyHonourGoName, nil
}
