// Package codegen provides extension handling for OpenAPI x- properties.
package codegen

import (
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi/orderedmap"
	"go.yaml.in/yaml/v4"
)

// Extension names - new naming convention with x-oapi-codegen- prefix
const (
	// ExtTypeOverride specifies an external type to use instead of generating one.
	// Format: "TypeName" or "TypeName;import/path" or "TypeName;alias import/path"
	ExtTypeOverride = "x-oapi-codegen-type-override"

	// ExtNameOverride overrides the generated field name.
	ExtNameOverride = "x-oapi-codegen-name-override"

	// ExtTypeNameOverride overrides the generated type name.
	ExtTypeNameOverride = "x-oapi-codegen-type-name-override"

	// ExtSkipOptionalPointer skips pointer wrapping for optional fields.
	ExtSkipOptionalPointer = "x-oapi-codegen-skip-optional-pointer"

	// ExtJSONIgnore excludes the field from JSON marshaling (json:"-").
	ExtJSONIgnore = "x-oapi-codegen-json-ignore"

	// ExtOmitEmpty explicitly controls the omitempty JSON tag.
	ExtOmitEmpty = "x-oapi-codegen-omitempty"

	// ExtOmitZero adds omitzero to the JSON tag (Go 1.24+ encoding/json/v2).
	ExtOmitZero = "x-oapi-codegen-omitzero"

	// ExtEnumVarNames overrides the generated enum constant names.
	ExtEnumVarNames = "x-oapi-codegen-enum-varnames"

	// ExtDeprecatedReason provides a deprecation reason for documentation.
	ExtDeprecatedReason = "x-oapi-codegen-deprecated-reason"

	// ExtOrder controls field ordering in generated structs.
	ExtOrder = "x-oapi-codegen-order"
)

// Legacy extension names for backwards compatibility
const (
	legacyExtGoType                = "x-go-type"
	legacyExtGoTypeImport          = "x-go-type-import"
	legacyExtGoName                = "x-go-name"
	legacyExtGoTypeName            = "x-go-type-name"
	legacyExtGoTypeSkipOptionalPtr = "x-go-type-skip-optional-pointer"
	legacyExtGoJSONIgnore          = "x-go-json-ignore"
	legacyExtOmitEmpty             = "x-omitempty"
	legacyExtOmitZero              = "x-omitzero"
	legacyExtEnumVarNames          = "x-enum-varnames"
	legacyExtEnumNames             = "x-enumNames" // Alternative name
	legacyExtDeprecatedReason      = "x-deprecated-reason"
	legacyExtOrder                 = "x-order"
)

// TypeOverride represents an external type override with optional import.
type TypeOverride struct {
	TypeName    string // The Go type name (e.g., "uuid.UUID")
	ImportPath  string // Import path (e.g., "github.com/google/uuid")
	ImportAlias string // Optional import alias (e.g., "foo" for `import foo "..."`)
}

// Extensions holds parsed extension values for a schema or property.
type Extensions struct {
	TypeOverride        *TypeOverride // External type to use
	NameOverride        string        // Override field name
	TypeNameOverride    string        // Override generated type name
	SkipOptionalPointer *bool         // Skip pointer for optional fields
	JSONIgnore          *bool         // Exclude from JSON
	OmitEmpty           *bool         // Control omitempty
	OmitZero            *bool         // Control omitzero
	EnumVarNames        []string      // Override enum constant names
	DeprecatedReason    string        // Deprecation reason
	Order               *int          // Field ordering
}

// ParseExtensions extracts extension values from a schema's extensions map.
// It supports both new (x-oapi-codegen-*) and legacy (x-go-*) extension names,
// logging deprecation warnings for legacy names.
func ParseExtensions(extensions *orderedmap.Map[string, *yaml.Node], path string) (*Extensions, error) {
	if extensions == nil {
		return &Extensions{}, nil
	}

	ext := &Extensions{}

	// Legacy type override needs special handling: x-go-type and x-go-type-import
	// are separate extensions that must be combined
	var legacyGoType string
	var legacyGoTypeImport any

	for pair := extensions.First(); pair != nil; pair = pair.Next() {
		key := pair.Key()
		node := pair.Value()
		if node == nil {
			continue
		}

		val := decodeYAMLNode(node)

		switch key {
		case ExtTypeOverride:
			override, err := parseTypeOverride(val)
			if err != nil {
				return nil, fmt.Errorf("parsing %s: %w", key, err)
			}
			ext.TypeOverride = override

		case legacyExtGoType:
			if s, ok := val.(string); ok {
				legacyGoType = s
			}

		case legacyExtGoTypeImport:
			legacyGoTypeImport = val

		case ExtNameOverride, legacyExtGoName:
			s, err := asString(val, key)
			if err != nil {
				return nil, err
			}
			ext.NameOverride = s

		case ExtTypeNameOverride, legacyExtGoTypeName:
			s, err := asString(val, key)
			if err != nil {
				return nil, err
			}
			ext.TypeNameOverride = s

		case ExtSkipOptionalPointer, legacyExtGoTypeSkipOptionalPtr:
			b, err := asBool(val, key)
			if err != nil {
				return nil, err
			}
			ext.SkipOptionalPointer = &b

		case ExtJSONIgnore, legacyExtGoJSONIgnore:
			b, err := asBool(val, key)
			if err != nil {
				return nil, err
			}
			ext.JSONIgnore = &b

		case ExtOmitEmpty, legacyExtOmitEmpty:
			b, err := asBool(val, key)
			if err != nil {
				return nil, err
			}
			ext.OmitEmpty = &b

		case ExtOmitZero, legacyExtOmitZero:
			b, err := asBool(val, key)
			if err != nil {
				return nil, err
			}
			ext.OmitZero = &b

		case ExtEnumVarNames, legacyExtEnumVarNames, legacyExtEnumNames:
			s, err := asStringSlice(val, key)
			if err != nil {
				return nil, err
			}
			ext.EnumVarNames = s

		case ExtDeprecatedReason, legacyExtDeprecatedReason:
			s, err := asString(val, key)
			if err != nil {
				return nil, err
			}
			ext.DeprecatedReason = s

		case ExtOrder, legacyExtOrder:
			i, err := asInt(val, key)
			if err != nil {
				return nil, err
			}
			ext.Order = &i

		default:
			// Unknown extension - ignore
		}
	}

	// Combine legacy x-go-type and x-go-type-import if no new-style override was set
	if ext.TypeOverride == nil && legacyGoType != "" {
		ext.TypeOverride = buildLegacyTypeOverride(legacyGoType, legacyGoTypeImport)
	}

	return ext, nil
}

// hasExtension checks if an extension exists by either the new or legacy name.
// This is used to check for extensions before fully parsing them.
func hasExtension(extensions *orderedmap.Map[string, *yaml.Node], newName, legacyName string) bool {
	if extensions == nil {
		return false
	}

	for pair := extensions.First(); pair != nil; pair = pair.Next() {
		key := pair.Key()
		if key == newName || key == legacyName {
			return true
		}
	}
	return false
}

// decodeYAMLNode converts a yaml.Node to a Go value.
func decodeYAMLNode(node *yaml.Node) any {
	if node == nil {
		return nil
	}

	var result any
	if err := node.Decode(&result); err != nil {
		return nil
	}
	return result
}

// parseTypeOverride parses the new combined type override format.
// Format: "TypeName" or "TypeName;import/path" or "TypeName;alias import/path"
func parseTypeOverride(val any) (*TypeOverride, error) {
	str, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("expected string, got %T", val)
	}

	override := &TypeOverride{}
	parts := strings.SplitN(str, ";", 2)
	override.TypeName = strings.TrimSpace(parts[0])

	if len(parts) == 2 {
		importPart := strings.TrimSpace(parts[1])
		importParts := strings.SplitN(importPart, " ", 2)
		if len(importParts) == 2 {
			override.ImportAlias = strings.TrimSpace(importParts[0])
			override.ImportPath = strings.TrimSpace(importParts[1])
		} else {
			override.ImportPath = importPart
		}
	}

	return override, nil
}

// buildLegacyTypeOverride combines legacy x-go-type and x-go-type-import values.
func buildLegacyTypeOverride(typeName string, importVal any) *TypeOverride {
	override := &TypeOverride{
		TypeName: typeName,
	}

	if importVal == nil {
		return override
	}

	// Legacy import can be a string or an object with path/name
	switch v := importVal.(type) {
	case string:
		override.ImportPath = v
	case map[string]any:
		if p, ok := v["path"].(string); ok {
			override.ImportPath = p
		}
		if name, ok := v["name"].(string); ok {
			override.ImportAlias = name
		}
	}

	return override
}

// Type conversion helpers that include the extension name in error messages

func asString(val any, extName string) (string, error) {
	if val == nil {
		return "", nil
	}
	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("parsing %s: expected string, got %T", extName, val)
	}
	return str, nil
}

func asBool(val any, extName string) (bool, error) {
	if val == nil {
		return false, nil
	}
	b, ok := val.(bool)
	if !ok {
		return false, fmt.Errorf("parsing %s: expected bool, got %T", extName, val)
	}
	return b, nil
}

func asInt(val any, extName string) (int, error) {
	if val == nil {
		return 0, nil
	}
	switch v := val.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("parsing %s: expected int, got %T", extName, val)
	}
}

func asStringSlice(val any, extName string) ([]string, error) {
	if val == nil {
		return nil, nil
	}
	slice, ok := val.([]any)
	if !ok {
		return nil, fmt.Errorf("parsing %s: expected array, got %T", extName, val)
	}
	result := make([]string, len(slice))
	for i, v := range slice {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("parsing %s: expected string at index %d, got %T", extName, i, v)
		}
		result[i] = str
	}
	return result, nil
}
