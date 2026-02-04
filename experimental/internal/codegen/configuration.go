package codegen

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"sort"
)

type Configuration struct {
	// PackageName which will be used in all generated files
	PackageName string `yaml:"package"`
	// Output specifies the output file path
	Output string `yaml:"output"`
	// TypeMapping allows customizing OpenAPI type/format to Go type mappings
	TypeMapping TypeMapping `yaml:"type-mapping,omitempty"`
	// NameMangling configures how OpenAPI names are converted to Go identifiers
	NameMangling NameMangling `yaml:"name-mangling,omitempty"`
	// NameSubstitutions allows direct overrides of generated names
	NameSubstitutions NameSubstitutions `yaml:"name-substitutions,omitempty"`
	// ImportMapping maps external spec file paths to Go package import paths.
	// Example: {"../common/api.yaml": "github.com/org/project/common"}
	// Use "-" as the value to indicate types should be in the current package.
	ImportMapping map[string]string `yaml:"import-mapping,omitempty"`
	// ContentTypes is a list of regexp patterns for media types to generate models for.
	// Only request/response bodies with matching content types will have types generated.
	// Defaults to common JSON and YAML types if not specified.
	ContentTypes []string `yaml:"content-types,omitempty"`
}

// DefaultContentTypes returns the default list of content type patterns.
// These match common JSON and YAML media types.
func DefaultContentTypes() []string {
	return []string{
		`^application/json$`,
		`^application/.*\+json$`,
	}
}

// ApplyDefaults merges user configuration on top of default values.
func (c *Configuration) ApplyDefaults() {
	c.TypeMapping = DefaultTypeMapping.Merge(c.TypeMapping)
	c.NameMangling = DefaultNameMangling().Merge(c.NameMangling)
	if len(c.ContentTypes) == 0 {
		c.ContentTypes = DefaultContentTypes()
	}
}

// ContentTypeMatcher checks if content types match configured patterns.
type ContentTypeMatcher struct {
	patterns []*regexp.Regexp
}

// NewContentTypeMatcher creates a matcher from a list of regexp patterns.
// Invalid patterns are silently ignored.
func NewContentTypeMatcher(patterns []string) *ContentTypeMatcher {
	m := &ContentTypeMatcher{
		patterns: make([]*regexp.Regexp, 0, len(patterns)),
	}
	for _, p := range patterns {
		if re, err := regexp.Compile(p); err == nil {
			m.patterns = append(m.patterns, re)
		}
	}
	return m
}

// Matches returns true if the content type matches any of the configured patterns.
func (m *ContentTypeMatcher) Matches(contentType string) bool {
	for _, re := range m.patterns {
		if re.MatchString(contentType) {
			return true
		}
	}
	return false
}

// ExternalImport represents an external package import with its alias.
type ExternalImport struct {
	Alias string // Short alias for use in generated code (e.g., "ext_a1b2c3")
	Path  string // Full import path (e.g., "github.com/org/project/common")
}

// ImportResolver resolves external references to Go package imports.
type ImportResolver struct {
	mapping map[string]ExternalImport // spec file path -> import info
}

// NewImportResolver creates an ImportResolver from the configuration's import mapping.
func NewImportResolver(importMapping map[string]string) *ImportResolver {
	resolver := &ImportResolver{
		mapping: make(map[string]ExternalImport),
	}

	for specPath, pkgPath := range importMapping {
		if pkgPath == "-" {
			// "-" means current package, no import needed
			resolver.mapping[specPath] = ExternalImport{Alias: "", Path: ""}
		} else {
			resolver.mapping[specPath] = ExternalImport{
				Alias: hashImportAlias(pkgPath),
				Path:  pkgPath,
			}
		}
	}

	return resolver
}

// Resolve looks up an external spec file path and returns its import info.
// Returns nil if the path is not in the mapping.
func (r *ImportResolver) Resolve(specPath string) *ExternalImport {
	if imp, ok := r.mapping[specPath]; ok {
		return &imp
	}
	return nil
}

// AllImports returns all external imports sorted by alias.
func (r *ImportResolver) AllImports() []ExternalImport {
	var imports []ExternalImport
	for _, imp := range r.mapping {
		if imp.Path != "" { // Skip current package markers
			imports = append(imports, imp)
		}
	}
	sort.Slice(imports, func(i, j int) bool {
		return imports[i].Alias < imports[j].Alias
	})
	return imports
}

// hashImportAlias generates a short, deterministic alias from an import path.
// Uses first 8 characters of SHA256 hash prefixed with "ext_".
func hashImportAlias(importPath string) string {
	h := sha256.Sum256([]byte(importPath))
	return "ext_" + hex.EncodeToString(h[:])[:8]
}
