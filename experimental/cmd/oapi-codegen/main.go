package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"

	"gopkg.in/yaml.v3"

	"github.com/oapi-codegen/oapi-codegen/experimental/internal/codegen"
)

func main() {
	configPath := flag.String("config", "", "path to configuration file")
	flagPackage := flag.String("package", "", "Go package name for generated code")
	flagOutput := flag.String("output", "", "output file path (default: <spec-basename>.gen.go)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <spec-path-or-url>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  spec-path-or-url    path or URL to OpenAPI spec file\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	specPath := flag.Arg(0)

	// Load the OpenAPI spec from file or URL
	specData, err := loadSpec(specPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading spec: %v\n", err)
		os.Exit(1)
	}

	// Configure libopenapi to skip resolving external references.
	// We handle external $refs via import mappings — the referenced specs
	// don't need to be fetched or parsed. See pb33f/libopenapi#519.
	docConfig := datamodel.NewDocumentConfiguration()
	docConfig.SkipExternalRefResolution = true

	doc, err := libopenapi.NewDocumentWithConfiguration(specData, docConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing spec: %v\n", err)
		os.Exit(1)
	}

	// Parse config if provided, otherwise use empty config
	var cfg codegen.Configuration
	if *configPath != "" {
		configData, err := os.ReadFile(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading config: %v\n", err)
			os.Exit(1)
		}
		if err := yaml.Unmarshal(configData, &cfg); err != nil {
			fmt.Fprintf(os.Stderr, "error parsing config: %v\n", err)
			os.Exit(1)
		}
	}

	// Flags override config file values
	if *flagPackage != "" {
		cfg.PackageName = *flagPackage
	}
	if *flagOutput != "" {
		cfg.Output = *flagOutput
	}

	// Default output to <spec-basename>.gen.go
	if cfg.Output == "" {
		// For URLs, extract the filename from the URL path
		baseName := specPath
		if u, err := url.Parse(specPath); err == nil && u.Scheme != "" && u.Host != "" {
			baseName = u.Path
		}
		base := filepath.Base(baseName)
		ext := filepath.Ext(base)
		cfg.Output = strings.TrimSuffix(base, ext) + ".gen.go"
	}

	// Default package name from output file
	if cfg.PackageName == "" {
		cfg.PackageName = "api"
	}

	// Generate code
	code, err := codegen.Generate(doc, specData, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating code: %v\n", err)
		os.Exit(1)
	}

	// Write output
	if err := os.WriteFile(cfg.Output, []byte(code), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s\n", cfg.Output)
}

// loadSpec loads an OpenAPI spec from a file path or URL.
func loadSpec(specPath string) ([]byte, error) {
	u, err := url.Parse(specPath)
	if err == nil && u.Scheme != "" && u.Host != "" {
		return loadSpecFromURL(u.String())
	}
	return os.ReadFile(specPath)
}

// loadSpecFromURL fetches an OpenAPI spec from an HTTP(S) URL.
func loadSpecFromURL(specURL string) ([]byte, error) {
	resp, err := http.Get(specURL) //nolint:gosec // URL comes from user-provided spec path
	if err != nil {
		return nil, fmt.Errorf("fetching spec from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching spec from URL: HTTP %d %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading spec from URL: %w", err)
	}
	return data, nil
}
