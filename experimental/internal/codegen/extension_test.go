package codegen

import (
	"testing"

	"github.com/pb33f/libopenapi/orderedmap"
	"go.yaml.in/yaml/v4"
)

func TestParseTypeOverride(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantType   string
		wantPath   string
		wantAlias  string
	}{
		{
			name:     "simple type",
			input:    "int64",
			wantType: "int64",
		},
		{
			name:     "type with import",
			input:    "uuid.UUID;github.com/google/uuid",
			wantType: "uuid.UUID",
			wantPath: "github.com/google/uuid",
		},
		{
			name:      "type with aliased import",
			input:     "foo.Type;foo github.com/bar/foo/v2",
			wantType:  "foo.Type",
			wantPath:  "github.com/bar/foo/v2",
			wantAlias: "foo",
		},
		{
			name:     "type with spaces",
			input:    " decimal.Decimal ; github.com/shopspring/decimal ",
			wantType: "decimal.Decimal",
			wantPath: "github.com/shopspring/decimal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTypeOverride(tt.input)
			if err != nil {
				t.Fatalf("parseTypeOverride() error = %v", err)
			}
			if got.TypeName != tt.wantType {
				t.Errorf("TypeName = %q, want %q", got.TypeName, tt.wantType)
			}
			if got.ImportPath != tt.wantPath {
				t.Errorf("ImportPath = %q, want %q", got.ImportPath, tt.wantPath)
			}
			if got.ImportAlias != tt.wantAlias {
				t.Errorf("ImportAlias = %q, want %q", got.ImportAlias, tt.wantAlias)
			}
		})
	}
}

func TestParseExtensions(t *testing.T) {
	// Create a test extensions map
	extensions := orderedmap.New[string, *yaml.Node]()

	// Add a type override extension
	typeOverrideNode := &yaml.Node{}
	typeOverrideNode.SetString("uuid.UUID;github.com/google/uuid")
	extensions.Set(ExtTypeOverride, typeOverrideNode)

	// Add a name override extension
	nameOverrideNode := &yaml.Node{}
	nameOverrideNode.SetString("CustomFieldName")
	extensions.Set(ExtNameOverride, nameOverrideNode)

	// Add omitempty extension
	omitEmptyNode := &yaml.Node{}
	if err := omitEmptyNode.Encode(true); err != nil {
		t.Fatalf("Failed to encode omitEmptyNode: %v", err)
	}
	extensions.Set(ExtOmitEmpty, omitEmptyNode)

	ext, err := ParseExtensions(extensions, "#/test/path")
	if err != nil {
		t.Fatalf("ParseExtensions() error = %v", err)
	}

	// Check type override
	if ext.TypeOverride == nil {
		t.Fatal("TypeOverride should not be nil")
	}
	if ext.TypeOverride.TypeName != "uuid.UUID" {
		t.Errorf("TypeOverride.TypeName = %q, want %q", ext.TypeOverride.TypeName, "uuid.UUID")
	}
	if ext.TypeOverride.ImportPath != "github.com/google/uuid" {
		t.Errorf("TypeOverride.ImportPath = %q, want %q", ext.TypeOverride.ImportPath, "github.com/google/uuid")
	}

	// Check name override
	if ext.NameOverride != "CustomFieldName" {
		t.Errorf("NameOverride = %q, want %q", ext.NameOverride, "CustomFieldName")
	}

	// Check omitempty
	if ext.OmitEmpty == nil || *ext.OmitEmpty != true {
		t.Errorf("OmitEmpty = %v, want true", ext.OmitEmpty)
	}
}

func TestParseExtensionsLegacy(t *testing.T) {
	// Create a test extensions map with legacy names
	extensions := orderedmap.New[string, *yaml.Node]()

	// Add legacy x-go-type extension
	goTypeNode := &yaml.Node{}
	goTypeNode.SetString("time.Time")
	extensions.Set("x-go-type", goTypeNode)

	// Add legacy x-go-type-import extension
	goImportNode := &yaml.Node{}
	goImportNode.SetString("time")
	extensions.Set("x-go-type-import", goImportNode)

	// Add legacy x-go-name extension
	goNameNode := &yaml.Node{}
	goNameNode.SetString("LegacyFieldName")
	extensions.Set("x-go-name", goNameNode)

	ext, err := ParseExtensions(extensions, "#/test/path")
	if err != nil {
		t.Fatalf("ParseExtensions() error = %v", err)
	}

	// Check type override (from legacy)
	if ext.TypeOverride == nil {
		t.Fatal("TypeOverride should not be nil")
	}
	if ext.TypeOverride.TypeName != "time.Time" {
		t.Errorf("TypeOverride.TypeName = %q, want %q", ext.TypeOverride.TypeName, "time.Time")
	}
	if ext.TypeOverride.ImportPath != "time" {
		t.Errorf("TypeOverride.ImportPath = %q, want %q", ext.TypeOverride.ImportPath, "time")
	}

	// Check name override (from legacy)
	if ext.NameOverride != "LegacyFieldName" {
		t.Errorf("NameOverride = %q, want %q", ext.NameOverride, "LegacyFieldName")
	}
}

func TestParseExtensionsEnumVarNames(t *testing.T) {
	extensions := orderedmap.New[string, *yaml.Node]()

	// Add enum var names as a sequence
	enumNamesNode := &yaml.Node{
		Kind: yaml.SequenceNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "Active"},
			{Kind: yaml.ScalarNode, Value: "Inactive"},
			{Kind: yaml.ScalarNode, Value: "Pending"},
		},
	}
	extensions.Set(ExtEnumVarNames, enumNamesNode)

	ext, err := ParseExtensions(extensions, "#/test/path")
	if err != nil {
		t.Fatalf("ParseExtensions() error = %v", err)
	}

	if len(ext.EnumVarNames) != 3 {
		t.Fatalf("EnumVarNames length = %d, want 3", len(ext.EnumVarNames))
	}
	expected := []string{"Active", "Inactive", "Pending"}
	for i, name := range ext.EnumVarNames {
		if name != expected[i] {
			t.Errorf("EnumVarNames[%d] = %q, want %q", i, name, expected[i])
		}
	}
}
