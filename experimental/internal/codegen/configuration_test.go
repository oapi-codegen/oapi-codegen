package codegen

import "testing"

func TestContentTypeMatcher(t *testing.T) {
	tests := []struct {
		name        string
		patterns    []string
		contentType string
		want        bool
	}{
		// Default patterns - JSON only (YAML not supported without custom unmarshalers)
		{"json exact", DefaultContentTypes(), "application/json", true},
		{"json+suffix", DefaultContentTypes(), "application/vnd.api+json", true},
		{"problem+json", DefaultContentTypes(), "application/problem+json", true},

		// YAML not in defaults (would need custom unmarshalers)
		{"yaml not default", DefaultContentTypes(), "application/yaml", false},
		{"text/yaml not default", DefaultContentTypes(), "text/yaml", false},

		// Non-matching
		{"text/plain", DefaultContentTypes(), "text/plain", false},
		{"text/html", DefaultContentTypes(), "text/html", false},
		{"application/xml", DefaultContentTypes(), "application/xml", false},
		{"application/octet-stream", DefaultContentTypes(), "application/octet-stream", false},
		{"multipart/form-data", DefaultContentTypes(), "multipart/form-data", false},
		{"image/png", DefaultContentTypes(), "image/png", false},

		// Custom patterns
		{"custom xml", []string{`^application/xml$`}, "application/xml", true},
		{"custom xml no match", []string{`^application/xml$`}, "application/json", false},
		{"custom wildcard", []string{`^text/.*`}, "text/plain", true},
		{"custom wildcard html", []string{`^text/.*`}, "text/html", true},
		{"custom yaml", []string{`^application/yaml$`}, "application/yaml", true},

		// Empty patterns
		{"empty patterns", []string{}, "application/json", false},

		// Invalid pattern (silently ignored)
		{"invalid pattern", []string{`[invalid`}, "application/json", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewContentTypeMatcher(tt.patterns)
			got := m.Matches(tt.contentType)
			if got != tt.want {
				t.Errorf("Matches(%q) = %v, want %v", tt.contentType, got, tt.want)
			}
		})
	}
}

func TestDefaultContentTypes(t *testing.T) {
	defaults := DefaultContentTypes()
	if len(defaults) == 0 {
		t.Error("DefaultContentTypes() returned empty slice")
	}

	// Verify all patterns are valid regexps
	m := NewContentTypeMatcher(defaults)
	if len(m.patterns) != len(defaults) {
		t.Errorf("Some default patterns failed to compile: got %d patterns, want %d",
			len(m.patterns), len(defaults))
	}
}
