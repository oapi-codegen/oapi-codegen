package codegen

import (
	"testing"

	"gopkg.in/yaml.v3"
)

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

func TestGenerationOptions_ServerYAML(t *testing.T) {
	t.Run("unmarshal server field", func(t *testing.T) {
		yamlContent := `
generation:
  server: std-http
`
		var cfg Configuration
		err := yaml.Unmarshal([]byte(yamlContent), &cfg)
		if err != nil {
			t.Fatalf("yaml.Unmarshal failed: %v", err)
		}
		if cfg.Generation.Server != ServerTypeStdHTTP {
			t.Errorf("Server = %q, want %q", cfg.Generation.Server, ServerTypeStdHTTP)
		}
	})

	t.Run("unmarshal empty server field", func(t *testing.T) {
		yamlContent := `
generation:
  no-models: true
`
		var cfg Configuration
		err := yaml.Unmarshal([]byte(yamlContent), &cfg)
		if err != nil {
			t.Fatalf("yaml.Unmarshal failed: %v", err)
		}
		if cfg.Generation.Server != "" {
			t.Errorf("Server = %q, want empty string", cfg.Generation.Server)
		}
	})

	t.Run("marshal server field", func(t *testing.T) {
		cfg := Configuration{
			PackageName: "test",
			Generation: GenerationOptions{
				Server: ServerTypeStdHTTP,
			},
		}
		data, err := yaml.Marshal(&cfg)
		if err != nil {
			t.Fatalf("yaml.Marshal failed: %v", err)
		}
		if got := string(data); !contains(got, "server: std-http") {
			t.Errorf("Marshaled YAML does not contain 'server: std-http':\n%s", got)
		}
	})

	t.Run("omit empty server field", func(t *testing.T) {
		cfg := Configuration{
			PackageName: "test",
			Generation:  GenerationOptions{},
		}
		data, err := yaml.Marshal(&cfg)
		if err != nil {
			t.Fatalf("yaml.Marshal failed: %v", err)
		}
		if got := string(data); contains(got, "server:") {
			t.Errorf("Marshaled YAML should not contain 'server:' when empty:\n%s", got)
		}
	})
}

func TestServerTypeConstants(t *testing.T) {
	if ServerTypeStdHTTP != "std-http" {
		t.Errorf("ServerTypeStdHTTP = %q, want %q", ServerTypeStdHTTP, "std-http")
	}
}

// contains is a simple helper for string containment check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
